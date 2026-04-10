package client

import (
	"bytes"
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

func (c *HTTPEngineClient) SubmitOrder(ctx context.Context, req models.OrderRequest, uid string) (*models.Order, error) {
	url := fmt.Sprintf("%s/api/engine/orders/submit", c.baseURL)

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Firebase-UID", uid)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("engine unreachable: %w", err)
	}
	defer resp.Body.Close()

	// Handle backend rejections (like 409 Conflict)
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", string(errorBody)) // Pass backend error string back
	}

	var order models.Order

	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, fmt.Errorf("failed to decode engine response: %w", err)
	}
	return &order, nil
}

func (c *HTTPEngineClient) GetActiveOrders(ctx context.Context, uid string) ([]models.Order, error) {
	url := fmt.Sprintf("%s/api/engine/orders/active", c.baseURL)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.Header.Set("X-Firebase-UID", uid)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("engine unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", string(errorBody))
	}

	var orders []models.Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("failed to decode engine response: %w", err)
	}
	return orders, nil
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
