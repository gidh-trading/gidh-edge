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
