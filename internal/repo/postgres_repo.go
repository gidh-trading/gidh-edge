package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/pkg/logger"
	"time"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) GetAvailable(ctx context.Context, date time.Time) ([]models.Instrument, error) {
	query := `SELECT DISTINCT c.instrument_token, c.stock_name 
              FROM instrument_configs c 
              INNER JOIN gidh_bars b ON c.instrument_token = b.instrument_token 
              WHERE b.timestamp::date = $1
              ORDER BY c.stock_name`
	rows, err := r.db.QueryContext(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Instrument
	for rows.Next() {
		var i models.Instrument
		rows.Scan(&i.Token, &i.StockName)
		list = append(list, i)
	}
	return list, nil
}

func (r *PostgresRepo) GetInstruments(ctx context.Context, date time.Time) ([]models.Instrument, error) {
	query := `SELECT DISTINCT instrument_token, stock_name 
              FROM instrument_configs 
              ORDER BY stock_name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Instrument
	for rows.Next() {
		var i models.Instrument
		if err := rows.Scan(&i.Token, &i.StockName); err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *PostgresRepo) GetBarsHistory(ctx context.Context, token uint32, date time.Time, timeframe string) ([]models.Bar, error) {
	// 1. Query fetches the single consolidated analytics JSONB column
	query := `
       SELECT 
          timestamp, instrument_token, stock_name, timeframe, 
          open, high, low, close, volume, tick_count, vwap,
          poc, vah, val, total_buy_qty, total_sell_qty, change_pct, analytics
       FROM gidh_bars 
       WHERE instrument_token = $1 
         AND timeframe = $3
         AND timestamp::date = (
             SELECT MAX(timestamp::date) 
             FROM gidh_bars 
             WHERE instrument_token = $1 AND timestamp::date = $2::date
         )
       ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"), timeframe)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bars []models.Bar
	for rows.Next() {
		var b models.Bar
		// Declare a temporary slice of bytes to hold the raw JSON data string from database
		var analyticsBytes []byte

		// 2. Scan directly into fields, catching the raw json array in our byte buffer
		err := rows.Scan(
			&b.Timestamp,
			&b.InstrumentToken,
			&b.StockName,
			&b.Timeframe,
			&b.Open,
			&b.High,
			&b.Low,
			&b.Close,
			&b.Volume,
			&b.TickCount,
			&b.VWAP,
			&b.POC,
			&b.VAH,
			&b.VAL,
			&b.TotalBuyQty,
			&b.TotalSellQty,
			&b.ChangePct,
			&analyticsBytes,
		)
		if err != nil {
			logger.Errorf("failed to scan bar rows fields: %v", err)
			return nil, err
		}

		// 3. 🔥 Unmarshal the json string straight into the internal struct property
		if len(analyticsBytes) > 0 {
			if err := json.Unmarshal(analyticsBytes, &b.Analytics); err != nil {
				logger.Errorf("failed to unmarshal nested analytics JSONB: %v", err)
				return nil, err
			}
		}

		bars = append(bars, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bars, nil
}

func (r *PostgresRepo) GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error) {
	var dna models.MarketDNA
	var hvnsJSON, lvnsJSON, bucketsJSON []byte

	// Filter by specific date to support backtesting
	query := `SELECT instrument_token, stock_name, trading_date, poc_5d, vah_5d, val_5d, macro_hvns, macro_lvns 
              FROM gidh_market_dna 
              WHERE instrument_token = $1 AND trading_date = $2::date`

	err := r.db.QueryRowContext(ctx, query, token, date.Format("2006-01-02")).Scan(
		&dna.InstrumentToken, &dna.StockName, &dna.TradingDate, &dna.POC, &dna.VAH, &dna.VAL, &hvnsJSON, &lvnsJSON,
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
              WHERE instrument_token = $1 AND trading_date = $2::date::date
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

func (r *PostgresRepo) GetDNADates(ctx context.Context) (map[string]bool, error) {
	query := `SELECT DISTINCT trading_date FROM gidh_market_dna`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make(map[string]bool)
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		dates[t.Format("2006-01-02")] = true
	}
	return dates, nil
}

func (r *PostgresRepo) GetPricePotential(ctx context.Context, stockName string, interval string) ([]models.PricePotential, error) {
	// 🎯 Note: 'timeframe' is queried in the WHERE clause, but aliased as 'interval' for the Go Struct mapper
	query := `
		SELECT stock_name, timeframe AS interval, p97, p90, p75, p50, p25
		FROM public.gidh_bars_price_potential
		WHERE stock_name = $1 AND timeframe = $2;`

	rows, err := r.db.QueryContext(ctx, query, stockName, interval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.PricePotential
	for rows.Next() {
		var p models.PricePotential
		err := rows.Scan(
			&p.StockName,
			&p.Interval, // Maps smoothly here!
			&p.P97,
			&p.P90,
			&p.P75,
			&p.P50,
			&p.P25,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
