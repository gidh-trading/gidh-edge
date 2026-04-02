package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"gidh-edge/internal/models"
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

func (r *PostgresRepo) GetHistory(ctx context.Context, token uint32, date time.Time) ([]models.Bar, error) {
	// Updated query to include vwap, poc, vah, and val
	query := `SELECT timestamp, open, high, low, close, volume, vwap, poc, vah, val 
              FROM gidh_bars 
              WHERE instrument_token = $1 AND timestamp::date = $2 
              ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"))
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
	query := `SELECT timestamp, anomaly_type, gidh_score, message FROM gidh_anomalies 
              WHERE instrument_token = $1 AND timestamp::date = $2 ORDER BY timestamp ASC`
	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Anomaly
	for rows.Next() {
		var a models.Anomaly
		rows.Scan(&a.Timestamp, &a.Type, &a.Severity, &a.Message)
		list = append(list, a)
	}
	return list, nil
}

func (r *PostgresRepo) GetBaseline(ctx context.Context, token uint32, date time.Time) (*models.Baseline, error) {
	var b models.Baseline
	var hvnsJSON, bucketsJSON []byte
	query := `SELECT instrument_token, symbol, trading_date, poc_5d, vah_5d, val_5d, macro_hvns, time_buckets 
              FROM market_dna WHERE instrument_token = $1 AND trading_date = $2::date`
	err := r.db.QueryRowContext(ctx, query, token, date.Format("2006-01-02")).Scan(
		&b.Token, &b.Symbol, &b.Date, &b.POC, &b.VAH, &b.VAL, &hvnsJSON, &bucketsJSON,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(hvnsJSON, &b.MacroHVNs)
	json.Unmarshal(bucketsJSON, &b.TimeBuckets)
	return &b, nil
}
