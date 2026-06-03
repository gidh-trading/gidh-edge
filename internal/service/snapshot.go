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

func (s *SnapshotService) GetFullDaySnapshot(ctx context.Context, token uint32, date time.Time, interval string) (models.Snapshot, error) {
	// 1. Fetch historical data from the repository
	history, _ := s.repo.GetBarsHistory(ctx, token, date, interval)

	dna, err := s.repo.GetMarketDNA(ctx, token, date)
	if err != nil {
		logger.Warnf("Failed to fetch Market DNA for token %d on %v: %v", token, date, err)
	}

	// 4. Fetch Volume Profiles ending at the specific backtesting date
	profiles, err := s.repo.GetVolumeProfiles(ctx, token, date, 5)
	if err != nil {
		logger.Warnf("Failed to fetch Volume Profiles for token %d on %v: %v", token, date, err)
	}

	// 🎯 5. Extract stock name from history and query the view
	var potential []models.PricePotential
	if len(history) > 0 {
		stockName := history[0].StockName
		potential, err = s.repo.GetPricePotential(ctx, stockName, interval)
		if err != nil {
			logger.Warnf("Failed to fetch Price Potential metrics for %s (%s): %v", stockName, interval, err)
		}
	}

	// Ensure it initializes to an empty slice instead of null if no data is found
	if potential == nil {
		potential = []models.PricePotential{}
	}

	// 6. Construct and return the full snapshot
	snapshot := models.Snapshot{
		HistoryBars:    history,
		MarketDNA:      dna,
		VolumeProfiles: profiles,
		PricePotential: potential, // 🎯 Included in the response
	}

	return snapshot, nil
}
