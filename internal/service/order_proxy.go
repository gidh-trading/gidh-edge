package service

import (
	"context"
	"gidh-edge/internal/client"
	"gidh-edge/internal/models"
)

type OrderService struct {
	engine *client.HTTPEngineClient
}

func NewOrderService(e *client.HTTPEngineClient) *OrderService {
	return &OrderService{engine: e}
}

func (s *OrderService) SubmitOrder(ctx context.Context, req models.OrderRequest, uid string) (*models.Order, error) {
	// Lean gateway: No DB checks here, just forward to the engine
	return s.engine.SubmitOrder(ctx, req, uid)
}

func (s *OrderService) GetActiveOrders(ctx context.Context, isBacktest bool, uid string) ([]models.Order, error) {
	return s.engine.GetActiveOrders(ctx, isBacktest, uid)
}
