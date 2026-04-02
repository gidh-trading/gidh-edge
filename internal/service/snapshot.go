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
	history, _ := s.repo.GetHistory(ctx, token, date)
	anomalies, _ := s.repo.GetAnomalies(ctx, token, date)

	// 1. Try Live Data first
	live, err := s.engine.GetActiveState(ctx, token)
	if err == nil {
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

	// 2. Fallback: Get historical profile from DB
	logger.Warnf("Engine offline, fetching profile from DB for token %d", token)

	snapshot := models.Snapshot{
		HistoryBars:      history,
		HistoryAnomalies: anomalies,
	}

	dbProfile, err := s.repo.GetVolumeProfile(ctx, token, date)
	if err == nil && dbProfile != nil {
		snapshot.VolumeProfile = *dbProfile
	}

	return snapshot, nil
}
