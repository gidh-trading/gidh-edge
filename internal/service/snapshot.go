package service

import (
	"context"
	"gidh-edge/internal/client"
	"gidh-edge/internal/models"
	"gidh-edge/internal/repo"
	"gidh-edge/pkg/logger"
	"time"
)

type SnapshotService struct {
	repo   repo.MarketDataRepo
	engine *client.HTTPEngineClient
}

func NewSnapshotService(r repo.MarketDataRepo, e *client.HTTPEngineClient) *SnapshotService {
	return &SnapshotService{repo: r, engine: e}
}

func (s *SnapshotService) GetFullDaySnapshot(ctx context.Context, token uint32, date time.Time) (models.Snapshot, error) {
	// 1. Rehydrate History from DB (Memory)
	history, _ := s.repo.GetHistory(ctx, token, date)
	anomalies, _ := s.repo.GetAnomalies(ctx, token, date)

	// 2. Fetch Live Pulse from Engine RAM
	live, err := s.engine.GetActiveState(ctx, token)
	if err != nil {
		logger.Warnf("Engine unreachable, using Post-Mortem history only: %v", err)
		return models.Snapshot{HistoryBars: history, HistoryAnomalies: anomalies}, nil
	}

	activeBars := make([]models.Bar, 0, len(live.ActiveBars))
	for _, b := range live.ActiveBars {
		activeBars = append(activeBars, b)
	}

	return models.Snapshot{
		HistoryBars:      history,
		HistoryAnomalies: anomalies,
		ActiveBars:       activeBars,
		VolumeProfile:    live.VolumeProfile,
	}, nil
}
