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

// 1. Duplicate the Physics Hierarchy from the backend
var anomalySeverity = map[models.AnomalyType]int{
	"ENERGY_PROPAGATION": 4,
	"ENERGY_SINK":        3,
	"ENERGY_LEAK":        2,
	"VOLUME_SURGE":       1,
}

type SnapshotService struct {
	repo   repo.MarketDataRepo
	engine *client.HTTPEngineClient
}

func NewSnapshotService(r repo.MarketDataRepo, e *client.HTTPEngineClient) *SnapshotService {
	return &SnapshotService{repo: r, engine: e}
}

func (s *SnapshotService) GetFullDaySnapshot(ctx context.Context, token uint32, date time.Time, interval string) (models.Snapshot, error) {
	// 1. Fetch historical data from the repository
	history, _ := s.repo.GetHistory(ctx, token, date, interval)

	// 2. Fetch PERFECTLY ALIGNED anomalies directly from the database
	// NOTE: Ensure your repo.GetAnomalies signature is updated to accept 'interval'
	intervalAnomalies, _ := s.repo.GetAnomalies(ctx, token, date, interval)

	// 3. Fetch DNA for the specific backtesting date
	dna, err := s.repo.GetMarketDNA(ctx, token, date)
	if err != nil {
		logger.Warnf("Failed to fetch Market DNA for token %d on %v: %v", token, date, err)
	}

	// 4. Fetch Volume Profiles ending at the specific backtesting date
	profiles, err := s.repo.GetVolumeProfiles(ctx, token, date, 5)
	if err != nil {
		logger.Warnf("Failed to fetch Volume Profiles for token %d on %v: %v", token, date, err)
	}

	// ---------------------------------------------------------
	// 5. LIVE DATA INTEGRATION & SANITIZATION (T>0 Merge)
	// ---------------------------------------------------------
	live, err := s.engine.GetActiveState(ctx, token, interval)
	var activeBars []models.Bar

	if err == nil {
		historyMap := make(map[time.Time]int)
		for i, b := range history {
			historyMap[b.Timestamp] = i
		}

		for _, b := range live.ActiveBars {
			if idx, exists := historyMap[b.Timestamp]; exists {
				// Overwrite historical placeholder with live finalized bar
				history[idx] = b
			} else {
				// Collect bars that haven't been persisted to history yet
				activeBars = append(activeBars, b)
			}
		}

		sort.Slice(activeBars, func(i, j int) bool {
			return activeBars[i].Timestamp.Before(activeBars[j].Timestamp)
		})
	} else {
		logger.Warnf("Failed to fetch active state from engine for token %d: %v", token, err)
	}

	// 6. Construct and return the full snapshot
	snapshot := models.Snapshot{
		HistoryBars:      history,
		HistoryAnomalies: intervalAnomalies, // Now natively filtered by 1m, 5m, 15m etc.
		MarketDNA:        dna,
		VolumeProfiles:   profiles,
		ActiveBars:       activeBars,
	}

	return snapshot, nil
}
