package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gidh-edge/internal/models"
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
	return &HTTPEngineClient{baseURL: url, client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPEngineClient) GetActiveState(ctx context.Context, token uint32, interval string) (EngineState, error) {
	url := fmt.Sprintf("%s/api/engine/active-state?token=%d&interval=%s", c.baseURL, token, interval)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return EngineState{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return EngineState{}, fmt.Errorf("engine returned status: %d", resp.StatusCode)
	}

	var state EngineState
	json.NewDecoder(resp.Body).Decode(&state)
	return state, nil
}

func (c *HTTPEngineClient) ForwardRawRequest(ctx context.Context, method, uri string, body io.Reader, headers http.Header) (*http.Response, error) {
	fullURL := c.baseURL + uri
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// Forward necessary headers (like Auth) from the UI to the backend
	if contentType := headers.Get("Content-Type"); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	if auth := headers.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	return c.client.Do(req)
}
