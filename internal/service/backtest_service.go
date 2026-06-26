package service

import (
	"context"
	"fmt"
	"gidh-edge/internal/client"
	"gidh-edge/internal/repo"
	"io"
	"net/http"
	"os"
	"strings"
)

type BacktestService struct {
	engine    *client.HTTPEngineClient
	repo      repo.MarketDataRepo
	backupDir string
}

func NewBacktestService(e *client.HTTPEngineClient, r repo.MarketDataRepo, backupDir string) *BacktestService {
	return &BacktestService{engine: e, repo: r, backupDir: backupDir}
}

func (s *BacktestService) ProxyBacktestRequest(ctx context.Context, method, uri string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, uri, body, headers)
}

func (s *BacktestService) GetAvailableDates(ctx context.Context) ([]string, error) {
	// 1. Get dates from DB that have DNA
	dnaDates, err := s.repo.GetDNADates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DNA dates: %w", err)
	}

	// 2. Get dates from DB that have Instrument Profiles
	profileDates, err := s.repo.GetInstrumentProfileDates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instrument profile dates: %w", err)
	}

	// 3. Read the backup directory
	files, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var availableDates []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		// Pattern: backup_YYYY-MM-DD.tar.xz
		if strings.HasPrefix(name, "backup_") && strings.HasSuffix(name, ".tar.xz") {
			// Extract YYYY-MM-DD
			dateStr := strings.TrimPrefix(name, "backup_")
			dateStr = strings.TrimSuffix(dateStr, ".tar.xz")

			// 4. Check if this date exists in BOTH our DNA map AND the Profile map
			if dnaDates[dateStr] && profileDates[dateStr] {
				availableDates = append(availableDates, dateStr)
			}
		}
	}

	return availableDates, nil
}
