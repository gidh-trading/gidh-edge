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

func (c *HTTPEngineClient) GetActiveOrders(ctx context.Context, uid string, token uint32, date string) ([]models.Order, error) {
	// UPDATED: Append token and date as query parameters to match the Engine's new endpoint logic
	url := fmt.Sprintf("%s/api/engine/orders/active?token=%d&date=%s", c.baseURL, token, date)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-Firebase-UID", uid)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("engine unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("engine error (status %d): %s", resp.StatusCode, string(errorBody))
	}

	var orders []models.Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("failed to decode engine response: %w", err)
	}

	// Ensure we return an empty slice instead of nil to keep the frontend clean
	if orders == nil {
		return []models.Order{}, nil
	}

	return orders, nil
}

func (c *HTTPEngineClient) UpdateOrderRisk(ctx context.Context, orderID string, sl, tp float64) error {
	url := fmt.Sprintf("%s/api/engine/orders/modify", c.baseURL)
	payload, _ := json.Marshal(map[string]any{
		"id":          orderID,
		"stop_loss":   sl,
		"take_profit": tp,
	})

	req, _ := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("engine failed to update risk: status %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPEngineClient) CancelOrder(ctx context.Context, orderID string) error {
	url := fmt.Sprintf("%s/api/engine/orders/cancel?id=%s", c.baseURL, orderID)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("engine failed to cancel order: status %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPEngineClient) ExitPosition(ctx context.Context, req models.ExitRequest, uid string) error {
	url := fmt.Sprintf("%s/api/engine/orders/exit", c.baseURL)

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Firebase-UID", uid)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("engine unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("engine exit failed: %s", string(errorBody))
	}

	return nil
}
