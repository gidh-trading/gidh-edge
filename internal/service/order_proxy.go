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

func (s *OrderService) GetActiveOrders(ctx context.Context, uid string, token uint32, date string) ([]models.Order, error) {
	return s.engine.GetActiveOrders(ctx, uid, token, date)
}

func (s *OrderService) SubmitOrder(ctx context.Context, req models.OrderRequest, uid string) (*models.Order, error) {
	return s.engine.SubmitOrder(ctx, req, uid)
}

func (s *OrderService) UpdateRisk(ctx context.Context, orderID string, sl, tp float64) error {
	return s.engine.UpdateOrderRisk(ctx, orderID, sl, tp)
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID string) error {
	return s.engine.CancelOrder(ctx, orderID)
}

func (s *OrderService) ExitPosition(ctx context.Context, req models.ExitRequest, uid string) error {
	return s.engine.ExitPosition(ctx, req, uid)
}

func (s *OrderService) GetActivePositions(ctx context.Context, uid string) ([]models.Position, error) {
	return s.engine.GetActivePositions(ctx, uid)
}
