package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/pkg/logger"
	"strings"
	"time"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) GetAvailable(ctx context.Context, date time.Time) ([]models.Instrument, error) {
	query := `SELECT DISTINCT c.instrument_token, c.symbol 
              FROM instrument_configs c 
              INNER JOIN gidh_bars b ON c.instrument_token = b.instrument_token 
              WHERE b.timestamp::date = $1
              ORDER BY c.symbol`
	rows, err := r.db.QueryContext(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Instrument
	for rows.Next() {
		var i models.Instrument
		rows.Scan(&i.Token, &i.Symbol)
		list = append(list, i)
	}
	return list, nil
}

func (r *PostgresRepo) GetHistory(ctx context.Context, token uint32, date time.Time, interval string) ([]models.Bar, error) {

	// Map UI interval (1m, 5m) to DB format (1m0s, 5m0s)
	dbInterval := interval
	if interval != "" && !strings.HasSuffix(interval, "0s") {
		dbInterval = interval + "0s"
	}

	query := `
       SELECT 
          timestamp, open, high, low, close, volume,
          cvd,cvd_divergence, effort_score,result_score,
          pulse_score,peak_trade_sign, total_buy_qty, total_sell_qty, 
          vwap, poc, vah, val
       FROM gidh_bars 
       WHERE instrument_token = $1 
         AND interval = $3
         AND timestamp::date = (
             SELECT MAX(timestamp::date) 
             FROM gidh_bars 
             WHERE instrument_token = $1 AND timestamp::date <= $2::date
         )
       ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"), dbInterval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bars []models.Bar
	for rows.Next() {
		var b models.Bar
		// Updated Scan to include support_levels and resistance_levels
		err := rows.Scan(
			&b.Timestamp,
			&b.Open,
			&b.High,
			&b.Low,
			&b.Close,
			&b.Volume,
			&b.CVD,
			&b.CVDDivergence,
			&b.EffortScore,
			&b.ResultScore,
			&b.PulseScore,
			&b.PeakTradeSign,
			&b.TotalBuyQty,
			&b.TotalSellQty,
			&b.VWAP,
			&b.POC,
			&b.VAH,
			&b.VAL,
		)
		if err != nil {
			logger.Errorf("failed to scan bar: %v", err)
			return nil, err
		}

		bars = append(bars, b)
	}
	return bars, nil
}

func (r *PostgresRepo) GetAnomalies(ctx context.Context, token uint32, date time.Time, interval string) ([]models.AnomalyEvent, error) {
	query := `
        SELECT 
            period_start, instrument_token, interval, symbol, anomaly_type, 
            time_key, last_updated_at, direction, effort_score, 
            result_score, pulse_score, intensity, price_value
        FROM gidh_anomalies 
        WHERE instrument_token = $1 
          AND period_start::date = $2::date
          AND interval = $3
        ORDER BY period_start ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"), interval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anomalies []models.AnomalyEvent
	for rows.Next() {
		var a models.AnomalyEvent
		err := rows.Scan(
			&a.PeriodStart, &a.InstrumentToken, &a.Interval, &a.Symbol, &a.Type,
			&a.TimeKey, &a.LastUpdatedAt, &a.Direction, &a.EffortScore,
			&a.ResultScore, &a.PulseScore, &a.Intensity, &a.PriceValue,
		)
		if err != nil {
			logger.Errorf("Error scanning anomaly row: %v", err)
			continue
		}
		anomalies = append(anomalies, a)
	}

	return anomalies, nil
}

func (r *PostgresRepo) GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error) {
	var dna models.MarketDNA
	var hvnsJSON, lvnsJSON, bucketsJSON []byte

	// Filter by specific date to support backtesting
	query := `SELECT instrument_token, stock_name, trading_date, poc_5d, vah_5d, val_5d, macro_hvns, macro_lvns, time_buckets 
              FROM gidh_market_dna 
              WHERE instrument_token = $1 AND trading_date = $2::date`

	err := r.db.QueryRowContext(ctx, query, token, date.Format("2006-01-02")).Scan(
		&dna.InstrumentToken, &dna.Symbol, &dna.TradingDate, &dna.POC, &dna.VAH, &dna.VAL, &hvnsJSON, &lvnsJSON, &bucketsJSON,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(hvnsJSON, &dna.MacroHVNs)
	json.Unmarshal(lvnsJSON, &dna.MacroLVNs)
	json.Unmarshal(bucketsJSON, &dna.TimeBuckets)

	return &dna, nil
}

func (r *PostgresRepo) GetVolumeProfiles(ctx context.Context, token uint32, date time.Time, limit int) ([]models.VolumeProfile, error) {
	// 1. Update the query to include the hvns and lvns columns
	query := `SELECT stock_name, instrument_token, trading_date, poc, vah, val, nodes, hvns, lvns 
              FROM gidh_volume_profiles 
              WHERE instrument_token = $1 AND trading_date <= $2::date
              ORDER BY trading_date DESC LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []models.VolumeProfile
	for rows.Next() {
		var vp models.VolumeProfile
		var nodesJSON, hvnsJSON, lvnsJSON []byte // Added hvnsJSON and lvnsJSON

		// 2. Update the Scan to include the new JSON columns
		err := rows.Scan(
			&vp.StockName,
			&vp.InstrumentToken,
			&vp.TradingDate,
			&vp.POC,
			&vp.VAH,
			&vp.VAL,
			&nodesJSON,
			&hvnsJSON, // New
			&lvnsJSON, // New
		)
		if err != nil {
			return nil, err
		}

		// 3. Unmarshal all three structural layers
		json.Unmarshal(nodesJSON, &vp.Nodes)
		json.Unmarshal(hvnsJSON, &vp.HVNs) // New
		json.Unmarshal(lvnsJSON, &vp.LVNs) // New

		profiles = append(profiles, vp)
	}

	return profiles, nil
}

func (r *PostgresRepo) GetActiveOrder(ctx context.Context, token uint32, isBacktest bool, uid string) (*models.Order, error) {
	query := `SELECT order_id, status FROM orders 
              WHERE instrument_token = $1 AND status IN ('OPEN', 'TRIGGER PENDING') 
              AND is_backtest = $2`

	params := []interface{}{token, isBacktest}
	if isBacktest {
		query += " AND firebase_uid = $3"
		params = append(params, uid)
	}

	var o models.Order
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&o.OrderID, &o.Status)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *PostgresRepo) SaveOrder(ctx context.Context, o *models.Order, uid string) error {
	query := `INSERT INTO orders (instrument_token, symbol, order_id, side, quantity, status, is_backtest, firebase_uid) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query, o.InstrumentToken, o.Symbol, o.OrderID, o.Side, o.Quantity, o.Status, o.IsBacktest, uid)
	return err
}

func (r *PostgresRepo) GetActiveOrders(ctx context.Context, isBacktest bool, uid string) ([]models.Order, error) {
	query := `SELECT id, order_id, instrument_token, symbol, side, status, quantity, is_backtest, created_at 
              FROM orders 
              WHERE status IN ('OPEN', 'TRIGGER PENDING') AND is_backtest = $1`

	params := []interface{}{isBacktest}
	if isBacktest {
		query += " AND firebase_uid = $2"
		params = append(params, uid)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.OrderID, &o.InstrumentToken, &o.Symbol, &o.Side, &o.Status, &o.Quantity, &o.IsBacktest, &o.CreatedAt)
		if err != nil {
			continue
		}
		list = append(list, o)
	}
	return list, nil
}
