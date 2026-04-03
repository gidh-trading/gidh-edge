package service

import (
	"context"
	"gidh-edge/internal/client"
	"gidh-edge/internal/models"
	"gidh-edge/internal/repo"
	"gidh-edge/pkg/logger"
	"sort"
	"time"
)

type SnapshotService struct {
	repo   repo.MarketDataRepo
	engine *client.HTTPEngineClient
}

func NewSnapshotService(r repo.MarketDataRepo, e *client.HTTPEngineClient) *SnapshotService {
	return &SnapshotService{repo: r, engine: e}
}

func (s *SnapshotService) GetFullDaySnapshot(ctx context.Context, token uint32, date time.Time, interval string) (models.Snapshot, error) {
	history, _ := s.repo.GetHistory(ctx, token, date, interval)
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

	// ---------------------------------------------------------
	// LIVE DATA INTEGRATION & SANITIZATION (T>0 Merge)
	// ---------------------------------------------------------
	live, err := s.engine.GetActiveState(ctx, token, interval)
	var activeBars []models.Bar

	if err == nil {
		// 1. Create a map of History by Timestamp for O(1) deduplication
		historyMap := make(map[time.Time]int)
		for i, b := range history {
			historyMap[b.Timestamp] = i
		}

		// 2. Process active bars from the Engine
		for _, b := range live.ActiveBars {
			if idx, exists := historyMap[b.Timestamp]; exists {
				// Collision: The DB and the Live Engine both have this bar.
				// We overwrite the DB version with the Engine's version,
				// as the Engine has the latest live ticks/volume.
				history[idx] = b
			} else {
				// It's a completely new bar not yet saved to DB
				activeBars = append(activeBars, b)
			}
		}

		// 3. CRITICAL: Go map iteration is random. We MUST sort the activeBars
		// chronologically before sending them to the UI to satisfy charting libraries.
		sort.Slice(activeBars, func(i, j int) bool {
			return activeBars[i].Timestamp.Before(activeBars[j].Timestamp)
		})
	} else {
		logger.Warnf("Failed to fetch active state from engine for token %d: %v", token, err)
	}

	snapshot := models.Snapshot{
		HistoryBars:      history,
		HistoryAnomalies: anomalies,
		MarketDNA:        dna,
		VolumeProfiles:   profiles,
		ActiveBars:       activeBars, // Now strictly sorted and deduped
	}

	return snapshot, nil
}
