package service

import (
	"context"
	"gidh-edge/internal/client"
	"gidh-edge/internal/models"
	"gidh-edge/internal/repo"
	"time"
)

type EdgeService struct {
	repo   *repo.PostgresRepo
	engine client.EngineClient
}

func NewEdgeService(r *repo.PostgresRepo, e client.EngineClient) *EdgeService {
	return &EdgeService{repo: r, engine: e}
}

func (s *EdgeService) GetCalendar(ctx context.Context) ([]string, error) {
	return s.repo.GetCalendar(ctx)
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

func (s *EdgeService) GetBaselines(ctx context.Context, token uint32, date time.Time) (models.Baseline, error) {
	return s.repo.GetBaselines(ctx, token, date)
}

func (s *EdgeService) GetEngineStatus(ctx context.Context) string {
	_, err := s.engine.GetActiveState(ctx, 0) // Health-check ping
	if err != nil {
		return "post-mortem"
	}
	return "active"
}
