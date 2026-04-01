package client

import (
	"context"
	"encoding/json"
	"fmt"
	"gidh-edge/internal/models"
	"net/http"
	"time"
)

type EngineState struct {
	ActiveBars    map[string]models.Bar `json:"active_bars"`
	VolumeProfile models.VolumeProfile  `json:"volume_profile"`
}

type HTTPEngineClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPEngineClient(url string) *HTTPEngineClient {
	return &HTTPEngineClient{baseURL: url, client: &http.Client{Timeout: 2 * time.Second}}
}

func (c *HTTPEngineClient) GetActiveState(ctx context.Context, token uint32) (EngineState, error) {
	url := fmt.Sprintf("%s/api/active-state?token=%d", c.baseURL, token)
	resp, err := c.client.Get(url)
	if err != nil {
		return EngineState{}, err
	}
	defer resp.Body.Close()

	var state EngineState
	json.NewDecoder(resp.Body).Decode(&state)
	return state, nil
}
