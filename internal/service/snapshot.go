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

	// Fetch DNA for the specific backtesting date
	dna, err := s.repo.GetMarketDNA(ctx, token, date)
	if err != nil {
		logger.Warnf("Failed to fetch Market DNA for token %d on %v: %v", token, date, err)
	}

	// Fetch Volume Profiles ending at the specific backtesting date
	profiles, err := s.repo.GetVolumeProfiles(ctx, token, date, 5)
	if err != nil {
		logger.Warnf("Failed to fetch Volume Profiles for token %d on %v: %v", token, date, err)
	}

	snapshot := models.Snapshot{
		HistoryBars:      history,
		HistoryAnomalies: anomalies,
		MarketDNA:        dna,
		VolumeProfiles:   profiles,
	}

	// Live data check (usually skipped in deep backtesting but kept for hybrid modes)
	live, err := s.engine.GetActiveState(ctx, token)
	if err == nil {
		activeBars := make([]models.Bar, 0, len(live.ActiveBars))
		for _, b := range live.ActiveBars {
			activeBars = append(activeBars, b)
		}
		snapshot.ActiveBars = activeBars
	}

	return snapshot, nil
}
