package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"gidh-edge/internal/models"
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
              WHERE b.timestamp::date = $1`
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

	query := `SELECT timestamp, open, high, low, close, volume, vwap, poc, vah, val 
              FROM gidh_bars 
              WHERE instrument_token = $1 AND timestamp::date = $2 AND interval = $3
              ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"), dbInterval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bars []models.Bar
	for rows.Next() {
		var b models.Bar
		// Updated Scan to include the 4 new fields
		err := rows.Scan(
			&b.Timestamp,
			&b.Open,
			&b.High,
			&b.Low,
			&b.Close,
			&b.Volume,
			&b.VWAP,
			&b.POC,
			&b.VAH,
			&b.VAL,
		)
		if err != nil {
			return nil, err
		}
		bars = append(bars, b)
	}
	return bars, nil
}

func (r *PostgresRepo) GetAnomalies(ctx context.Context, token uint32, date time.Time) ([]models.Anomaly, error) {
	query := `
        SELECT 
            period_start, 
            last_updated_at, 
            anomaly_type, 
            upgrade_count,
            effort_score, 
            result_score, 
            divergence_score,
            price_value, 
            price_baseline,
            dist_poc_pct, 
            dist_vah_pct, 
            dist_val_pct
        FROM gidh_anomalies 
        WHERE instrument_token = $1 
          AND period_start::date = $2 
        ORDER BY period_start ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Anomaly
	for rows.Next() {
		var a models.Anomaly
		err := rows.Scan(
			&a.PeriodStart,
			&a.LastUpdatedAt,
			&a.Type,
			&a.UpgradeCount,
			&a.EffortScore,
			&a.ResultScore,
			&a.DivergenceScore,
			&a.PriceValue,
			&a.PriceBaseline,
			&a.DistPOCPct,
			&a.DistVAHPct,
			&a.DistVALPct,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *PostgresRepo) GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error) {
	var dna models.MarketDNA
	var hvnsJSON, bucketsJSON []byte

	// Filter by specific date to support backtesting
	query := `SELECT instrument_token, stock_name, trading_date, poc_5d, vah_5d, val_5d, macro_hvns, time_buckets 
              FROM gidh_market_dna 
              WHERE instrument_token = $1 AND trading_date = $2::date`

	err := r.db.QueryRowContext(ctx, query, token, date.Format("2006-01-02")).Scan(
		&dna.InstrumentToken, &dna.Symbol, &dna.TradingDate, &dna.POC, &dna.VAH, &dna.VAL, &hvnsJSON, &bucketsJSON,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(hvnsJSON, &dna.MacroHVNs)
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
