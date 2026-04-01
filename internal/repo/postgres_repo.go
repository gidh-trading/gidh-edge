package repo

import (
	"context"
	"database/sql"
	"gidh-edge/internal/models"
	"time"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// GetAvailable identifies stocks with recorded pulses (bars) for a specific date.
// It joins instrument_configs with gidh_bars to provide a "Data-Proven" list.
func (r *PostgresRepo) GetAvailable(ctx context.Context, date time.Time) ([]models.Instrument, error) {
	query := `
		SELECT DISTINCT c.instrument_token, c.symbol 
		FROM instrument_configs c
		INNER JOIN gidh_bars b ON c.instrument_token = b.instrument_token
		WHERE b.timestamp::date = $1
		ORDER BY c.symbol ASC`

	rows, err := r.db.QueryContext(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instruments []models.Instrument
	for rows.Next() {
		var i models.Instrument
		if err := rows.Scan(&i.Token, &i.Symbol); err != nil {
			return nil, err
		}
		instruments = append(instruments, i)
	}
	return instruments, nil
}

// GetHistory fetches the "Memory" of the session (finalized candles).
func (r *PostgresRepo) GetHistory(ctx context.Context, token uint32, date time.Time) ([]models.Bar, error) {
	// Note: instrument_token is the standard name in gidh_bars
	query := `
		SELECT timestamp, open, high, low, close, volume 
		FROM gidh_bars 
		WHERE instrument_token = $1 AND timestamp::date = $2 
		ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.Bar
	for rows.Next() {
		var b models.Bar
		if err := rows.Scan(&b.Timestamp, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume); err != nil {
			return nil, err
		}
		history = append(history, b)
	}
	return history, nil
}

// GetAnomalies fetches historical "Grains" to populate the initial Bubble Map.
func (r *PostgresRepo) GetAnomalies(ctx context.Context, token uint32, date time.Time) ([]models.Anomaly, error) {
	// Mapping: anomaly_type -> Type, gidh_score -> Severity
	query := `
		SELECT timestamp, anomaly_type, gidh_score, message 
		FROM gidh_anomalies 
		WHERE instrument_token = $1 AND timestamp::date = $2
		ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, token, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anomalies []models.Anomaly
	for rows.Next() {
		var a models.Anomaly
		if err := rows.Scan(&a.Timestamp, &a.Type, &a.Severity, &a.Message); err != nil {
			return nil, err
		}
		anomalies = append(anomalies, a)
	}
	return anomalies, nil
}

func (r *PostgresRepo) GetCalendar(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT timestamp::date::text FROM gidh_bars ORDER BY 1 DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		rows.Scan(&d)
		dates = append(dates, d)
	}
	return dates, nil
}

func (r *PostgresRepo) GetBaselines(ctx context.Context, token uint32, date time.Time) (models.Baseline, error) {
	var b models.Baseline
	query := `SELECT vah, val, poc FROM gidh_baselines WHERE token = $1 AND date = $2`
	err := r.db.QueryRowContext(ctx, query, token, date.Format("2006-01-02")).Scan(&b.VAH, &b.VAL, &b.POC)
	if err == sql.ErrNoRows {
		return b, nil // Return empty if not calculated yet
	}
	return b, err
}
