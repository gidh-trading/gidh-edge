// internal/service/order_service.go

package service

import (
	"context"
	"gidh-edge/internal/client"
	"io"
	"net/http"
)

type OrderService struct {
	engine *client.HTTPEngineClient
}

func NewOrderService(e *client.HTTPEngineClient) *OrderService {
	return &OrderService{engine: e}
}

// ProxyOrder forwards the raw request directly to the backend engine
func (s *OrderService) ProxyOrder(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	// The URI should match the endpoint expected by your backend engine
	return s.engine.ForwardRawRequest(ctx, method, "/api/orders/place", body, headers)
}

// ProxyPositions forwards the request to fetch open positions from the backend engine
func (s *OrderService) ProxyPositions(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/positions", body, headers)
}

func (s *OrderService) ProxyOrderModify(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/orders/modify", body, headers)
}

func (s *OrderService) ProxyOrderCancel(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/orders/cancel", body, headers)
}

func (s *OrderService) ProxyPositionMetadata(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/positions/metadata", body, headers)
}

func (s *OrderService) ProxyPositionExit(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/positions/exit", body, headers)
}
