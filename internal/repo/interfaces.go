package repo

import (
	"context"
	"gidh-edge/internal/models"
	"time"
)

type MarketDataRepo interface {
	GetHistory(ctx context.Context, token uint32, date time.Time, interval string) ([]models.Bar, error)
	GetAnomalies(ctx context.Context, token uint32, date time.Time) ([]models.AnomalyEvent, error)
	GetMarketDNA(ctx context.Context, token uint32, date time.Time) (*models.MarketDNA, error)
	GetVolumeProfiles(ctx context.Context, token uint32, date time.Time, limit int) ([]models.VolumeProfile, error)
}
