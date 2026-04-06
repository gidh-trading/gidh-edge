package service

import (
	"context"
	"gidh-edge/internal/client"
	"gidh-edge/internal/models"
	"gidh-edge/internal/repo"
	"gidh-edge/pkg/logger"
	"time"
)

type EdgeService struct {
	repo   *repo.PostgresRepo
	engine *client.HTTPEngineClient
}

func NewEdgeService(r *repo.PostgresRepo, e *client.HTTPEngineClient) *EdgeService {
	return &EdgeService{repo: r, engine: e}
}

func (s *EdgeService) GetInstruments(ctx context.Context, date time.Time) ([]models.Instrument, error) {
	return s.repo.GetAvailable(ctx, date)
}

func (s *EdgeService) GetHistoryBars(ctx context.Context, token uint32, date time.Time, interval string) ([]models.Bar, error) {
	return s.repo.GetHistory(ctx, token, date, interval)
}

func (s *EdgeService) GetHistorySignals(ctx context.Context, token uint32, date time.Time) ([]models.AnomalyEvent, error) {
	return s.repo.GetAnomalies(ctx, token, date)
}

func (s *EdgeService) GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error) {
	dna, err := s.repo.GetMarketDNA(ctx, token, date)
	if err != nil {
		logger.Errorf("Market DNA not found for token %d on %v", token, date)
		return nil, err
	}
	return dna, nil
}

func (s *EdgeService) GetEngineStatus(ctx context.Context) string {
	_, err := s.engine.GetActiveState(ctx, 0, "1m") // Health-check ping
	if err != nil {
		return "post-mortem"
	}
	return "active"
}
