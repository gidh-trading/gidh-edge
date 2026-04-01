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
	engine client.EngineClient
}

func NewSnapshotService(r repo.MarketDataRepo, e client.EngineClient) *SnapshotService {
	return &SnapshotService{
		repo:   r,
		engine: e,
	}
}

func (s *SnapshotService) GetFullDaySnapshot(ctx context.Context, token uint32, date time.Time) (models.Snapshot, error) {
	// 1. Get Memory from DB (Historical finalized bars)
	history, err := s.repo.GetHistory(ctx, token, date)
	if err != nil {
		logger.Errorf("Failed to fetch history from DB: %v", err)
		return models.Snapshot{}, err
	}

	// 2. Get Pulse from Engine (Attempt to get live un-finalized bars)
	liveState, err := s.engine.GetActiveState(ctx, token)
	if err != nil {
		// Post-Mortem Availability: Do not fail if Engine is down
		logger.Warnf("Ingestion Engine unreachable for token %d: %v. Returning history only.", token, err)
		return models.Snapshot{
			History: history,
			Active:  []models.Bar{},
			Profile: []float64{},
		}, nil
	}

	// 3. Return Stitched Result (Merging Memory and Pulse)
	return models.Snapshot{
		History: history,
		Active:  liveState.Bars,
		Profile: liveState.Profile,
	}, nil
}
