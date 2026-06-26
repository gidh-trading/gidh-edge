package service

import (
	"context"
	"fmt"
	"gidh-edge/internal/client"
	"gidh-edge/internal/repo"
	"io"
	"net/http"
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

	// 3. Find dates present in BOTH maps
	var availableDates []string

	// Iterate through DNA dates and check if they exist in the profile map
	for dateStr := range dnaDates {
		if profileDates[dateStr] {
			availableDates = append(availableDates, dateStr)
		}
	}

	return availableDates, nil
}
