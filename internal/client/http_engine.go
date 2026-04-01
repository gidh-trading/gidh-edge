package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPEngineClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPEngineClient(baseURL string) *HTTPEngineClient {
	return &HTTPEngineClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second, // Fast timeout for "Pulse" data
		},
	}
}

func (c *HTTPEngineClient) GetActiveState(ctx context.Context, token uint32) (EngineState, error) {
	url := fmt.Sprintf("%s/api/active-state?token=%d", c.baseURL, token)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return EngineState{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return EngineState{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return EngineState{}, fmt.Errorf("engine returned status: %d", resp.StatusCode)
	}

	var state EngineState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return EngineState{}, err
	}

	return state, nil
}
