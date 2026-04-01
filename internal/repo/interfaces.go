package repo

import (
	"context"
	"gidh-edge/internal/models"
	"time"
)

type MarketDataRepo interface {
	GetHistory(ctx context.Context, token uint32, date time.Time) ([]models.Bar, error)
	GetAnomalies(ctx context.Context, token uint32, date time.Time) ([]models.Anomaly, error)
}
