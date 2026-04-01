// internal/client/engine.go
package client

import (
	"context"
	"gidh-edge/internal/models"
)

// EngineState reflects the JSON structure returned by the Ingestion Engine's /api/active-state
type EngineState struct {
	Symbol        string                `json:"symbol"`
	ActiveBars    map[string]models.Bar `json:"active_bars"`    // Map of interval -> forming candle
	VolumeProfile models.VolumeProfile  `json:"volume_profile"` // Full auction state
}

type EngineClient interface {
	GetActiveState(ctx context.Context, token uint32) (EngineState, error)
}
