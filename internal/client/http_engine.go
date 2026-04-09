package client

import (
	"bytes"
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

func (c *HTTPEngineClient) GetActiveState(ctx context.Context, token uint32, interval string) (EngineState, error) {
	url := fmt.Sprintf("%s/api/engine/active-state?token=%d&interval=%s", c.baseURL, token, interval)
	resp, err := c.client.Get(url)
	if err != nil {
		return EngineState{}, err
	}
	defer resp.Body.Close()

	var state EngineState
	json.NewDecoder(resp.Body).Decode(&state)
	return state, nil
}

func (c *HTTPEngineClient) SubmitOrder(ctx context.Context, req models.OrderRequest, uid string) (*models.Order, error) {
	url := fmt.Sprintf("%s/api/engine/orders/submit", c.baseURL)

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Firebase-UID", uid) // Forward the UID

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("engine returned error: %d", resp.StatusCode)
	}

	var order models.Order
	json.NewDecoder(resp.Body).Decode(&order)
	return &order, nil
}

func (c *HTTPEngineClient) GetActiveOrders(ctx context.Context, isBacktest bool, uid string) ([]models.Order, error) {
	url := fmt.Sprintf("%s/api/engine/orders/active?backtest=%t", c.baseURL, isBacktest)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.Header.Set("X-Firebase-UID", uid)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var orders []models.Order
	json.NewDecoder(resp.Body).Decode(&orders)
	return orders, nil
}
