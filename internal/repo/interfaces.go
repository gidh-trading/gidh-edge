package repo

import (
	"context"
	"gidh-edge/internal/models"
	"time"
)

type InstrumentRepo interface {
	GetAvailable(ctx context.Context, date time.Time) ([]models.Instrument, error)
}

type MarketDataRepo interface {
	GetHistory(ctx context.Context, token uint32, date time.Time) ([]models.Bar, error)
	GetAnomalies(ctx context.Context, token uint32, date time.Time) ([]models.Anomaly, error)
}
