package repo

import (
	"context"
	"gidh-edge/internal/models"
	"time"
)

type MarketDataRepo interface {
	GetDNADates(ctx context.Context) (map[string]bool, error)
	GetAvailable(ctx context.Context, date time.Time) ([]models.Instrument, error)
	GetInstruments(ctx context.Context, date time.Time) ([]models.Instrument, error) // Added
	GetBarsHistory(ctx context.Context, token uint32, date time.Time, interval string) ([]models.Bar, error)
	GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error)
	GetVolumeProfiles(ctx context.Context, token uint32, date time.Time, limit int) ([]models.VolumeProfile, error)
	GetInstrumentProfiles(ctx context.Context) ([]models.InstrumentProfile, error)
}
