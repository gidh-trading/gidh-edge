package service

import (
	"context"
	"gidh-edge/internal/client"
	"gidh-edge/internal/models"
	"gidh-edge/internal/repo"
	"io"
	"net/http"
	"time"
)

type OrderService struct {
	engine *client.HTTPEngineClient
	repo   *repo.OrderRepository // Injected clean repository layer
}

func NewOrderService(e *client.HTTPEngineClient, repo *repo.OrderRepository) *OrderService {
	return &OrderService{engine: e, repo: repo}
}

// GetHistoricalOrders proxies the call down to the repository layer
func (s *OrderService) GetHistoricalOrders(ctx context.Context, tradingDate time.Time) ([]models.OrderBookEntry, error) {
	return s.repo.GetHistoricalOrders(ctx, tradingDate)
}

// GetHistoricalPositions proxies the call down to the repository layer
func (s *OrderService) GetHistoricalPositions(ctx context.Context, tradingDate time.Time) ([]models.Position, error) {
	return s.repo.GetHistoricalPositions(ctx, tradingDate)
}

// --- Proxy Order System calls remain completely unchanged below ---

func (s *OrderService) ProxyOrder(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/orders/place", body, headers)
}

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
