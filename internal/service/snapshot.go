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

// internal/service/snapshot.go

func (s *SnapshotService) GetFullDaySnapshot(ctx context.Context, token uint32, date time.Time) (models.Snapshot, error) {
	// 1. Get Memory from DB (Historical finalized bars)
	history, err := s.repo.GetHistory(ctx, token, date)
	if err != nil {
		logger.Errorf("Failed to fetch history from DB: %v", err)
		return models.Snapshot{}, err
	}

	// 2. Get Pulse from Engine (Live un-finalized state)
	liveState, err := s.engine.GetActiveState(ctx, token)
	if err != nil {
		// Post-Mortem Mode: fallback to history if engine is unreachable
		logger.Debugf("Engine unreachable for token %d: %v. Returning history only.", token, err)
		return models.Snapshot{
			History: history,
			Active:  []models.Bar{},
		}, nil
	}

	// 3. Flatten the ActiveBars map into a slice for the UI
	activeSlice := make([]models.Bar, 0, len(liveState.ActiveBars))
	for _, bar := range liveState.ActiveBars {
		activeSlice = append(activeSlice, bar)
	}

	// 4. Return Stitched Result
	return models.Snapshot{
		History: history,
		Active:  activeSlice,
		Profile: liveState.VolumeProfile,
	}, nil
}
