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

func (r *PostgresRepo) GetHistory(ctx context.Context, token uint32, date time.Time, timeframe string) ([]models.Bar, error) {

	// No more manual string manipulation needed if the DB stores '1m' or '5m' directly
	query := `
       SELECT 
          timestamp, open, high, low, close, volume,
          vwap, poc, vah, val, buy_volume, sell_volume,
          total_vol_energy, buy_vol_energy, sell_vol_energy,
          total_rng_energy, buy_rng_energy, sell_rng_energy
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
			&b.BuyVolume,      // Added
			&b.SellVolume,     // Added
			&b.TotalVolEnergy, // New Energy Metric
			&b.BuyVolEnergy,   // New Energy Metric
			&b.SellVolEnergy,  // New Energy Metric
			&b.TotalRngEnergy, // New Energy Metric
			&b.BuyRngEnergy,   // New Energy Metric
			&b.SellRngEnergy,  // New Energy Metric
		)
		if err != nil {
			logger.Errorf("failed to scan bar: %v", err)
			return nil, err
		}

		bars = append(bars, b)
	}

	return bars, nil
}

func (r *PostgresRepo) GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error) {
	var dna models.MarketDNA
	var hvnsJSON, lvnsJSON, bucketsJSON []byte

	// Filter by specific date to support backtesting
	query := `SELECT instrument_token, stock_name, trading_date, poc_5d, vah_5d, val_5d, macro_hvns, macro_lvns, time_buckets 
              FROM gidh_market_dna 
              WHERE instrument_token = $1 AND trading_date = $2::date`

	err := r.db.QueryRowContext(ctx, query, token, date.Format("2006-01-02")).Scan(
		&dna.InstrumentToken, &dna.StockName, &dna.TradingDate, &dna.POC, &dna.VAH, &dna.VAL, &hvnsJSON, &lvnsJSON, &bucketsJSON,
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
