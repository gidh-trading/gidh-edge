package repo

import (
	"context"
	"gidh-edge/internal/models"
	"time"
)

type MarketDataRepo interface {
	GetAvailable(ctx context.Context, date time.Time) ([]models.Instrument, error)
	GetInstruments(ctx context.Context, date time.Time) ([]models.Instrument, error) // Added
	GetHistory(ctx context.Context, token uint32, date time.Time, interval string) ([]models.Bar, error)
	GetAnomalies(ctx context.Context, token uint32, date time.Time, interval string) ([]models.AnomalyEvent, error)
	GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error)
	GetVolumeProfiles(ctx context.Context, token uint32, date time.Time, limit int) ([]models.VolumeProfile, error)
}
