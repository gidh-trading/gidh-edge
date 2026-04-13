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

func (s *OrderService) ProxyOrderRequest(ctx context.Context, method, uri string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, uri, body, headers)
}
