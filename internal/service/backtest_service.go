// internal/service/backtest_service.go

package service

import (
	"context"
	"fmt"
	"gidh-edge/internal/client"
	"gidh-edge/internal/repo"
	"gidh-edge/pkg/postgres" // Import your postgres wrapper
	"io"
	"net/http"
	"time"
)

type BacktestService struct {
	engine     *client.HTTPEngineClient
	repo       repo.MarketDataRepo
	backupDir  string
	liveDBConn string // Store the live database string reference
}

func NewBacktestService(e *client.HTTPEngineClient, r repo.MarketDataRepo, backupDir string, liveDBConn string) *BacktestService {
	return &BacktestService{engine: e, repo: r, backupDir: backupDir, liveDBConn: liveDBConn}
}

func (s *BacktestService) ProxyBacktestRequest(ctx context.Context, method, uri string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, uri, body, headers)
}

func (s *BacktestService) GetAvailableDates(ctx context.Context) ([]string, error) {
	// 1. Get dates from local Backtest DB that have DNA
	dnaDates, err := s.repo.GetDNADates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DNA dates: %w", err)
	}

	// 2. Get dates from local Backtest DB that have Instrument Profiles
	profileDates, err := s.repo.GetInstrumentProfileDates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instrument profile dates: %w", err)
	}

	// 3. Get valid live tick dates directly from the Live Hypertable DB
	liveTickDates, err := s.fetchLiveTickDates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch live tick dates: %w", err)
	}

	// 4. Find dates present in ALL THREE data sets (DNA + Profile + Live Ticks)
	var availableDates []string

	for dateStr := range dnaDates {
		if profileDates[dateStr] && liveTickDates[dateStr] {
			availableDates = append(availableDates, dateStr)
		}
	}

	return availableDates, nil
}

// Helper method to establish a transient/pooled connection to the live DB and fetch dates
func (s *BacktestService) fetchLiveTickDates(ctx context.Context) (map[string]bool, error) {
	dates := make(map[string]bool)
	if s.liveDBConn == "" {
		return dates, fmt.Errorf("LIVE_DATABASE_URL is not configured")
	}

	// Open connection to the live environment database
	liveDB, err := postgres.New(s.liveDBConn)
	if err != nil {
		return nil, fmt.Errorf("unable to reach live database pipeline: %w", err)
	}
	defer liveDB.Close()

	// High-performance query against the pre-aggregated continuous materialized view
	query := `
		SELECT trading_date 
		FROM mv_live_tick_dates 
		WHERE tick_count > 0
		ORDER BY trading_date DESC;`

	rows, err := liveDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query live tick materialized view: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		dates[t.Format("2006-01-02")] = true
	}

	return dates, rows.Err()
}
