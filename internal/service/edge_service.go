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

func (s *EdgeService) GetHistoryBars(ctx context.Context, token uint32, date time.Time) ([]models.Bar, error) {
	return s.repo.GetHistory(ctx, token, date)
}

func (s *EdgeService) GetHistorySignals(ctx context.Context, token uint32, date time.Time) ([]models.Anomaly, error) {
	return s.repo.GetAnomalies(ctx, token, date)
}

func (s *EdgeService) GetBaseline(ctx context.Context, token uint32, date time.Time) (*models.Baseline, error) {
	// Subtract 1 day because baseline DNA is calculated and stored for the previous day
	lookupDate := date.AddDate(0, 0, -1)

	baseline, err := s.repo.GetBaseline(ctx, token, lookupDate)
	if err != nil {
		logger.Errorf("Baseline not found for token %d on %v (requested for %v)", token, lookupDate, date)
		return nil, err
	}
	return baseline, nil
}

func (s *EdgeService) GetEngineStatus(ctx context.Context) string {
	_, err := s.engine.GetActiveState(ctx, 0) // Health-check ping
	if err != nil {
		return "post-mortem"
	}
	return "active"
}
