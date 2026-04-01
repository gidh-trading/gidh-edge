package client

import (
	"context"
	"gidh-edge/internal/models"
)

// EngineState represents the current state from the Ingestion Engine
type EngineState struct {
	Bars    []models.Bar `json:"bars"`
	Profile []float64    `json:"profile"`
}

type EngineClient interface {
	GetActiveState(ctx context.Context, token uint32) (EngineState, error)
}
