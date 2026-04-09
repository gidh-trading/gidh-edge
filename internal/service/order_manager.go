package service

import (
	"context"
	"fmt"
	"gidh-edge/internal/models"
	"gidh-edge/internal/repo"
	"sync"
)

type OrderExecutor interface {
	PlaceOrder(ctx context.Context, req models.OrderRequest) (string, error)
	CancelOrder(ctx context.Context, orderID string) error
}

type OrderManager struct {
	repo          *repo.PostgresRepo
	liveExecutor  OrderExecutor
	paperExecutor OrderExecutor
	mu            sync.Mutex
}

func NewOrderManager(r *repo.PostgresRepo, live, paper OrderExecutor) *OrderManager {
	return &OrderManager{
		repo:          r,
		liveExecutor:  live,
		paperExecutor: paper,
	}
}

func (m *OrderManager) HandleOrder(ctx context.Context, req models.OrderRequest) (*models.Order, error) {
	// 1. Logic Check: DB will enforce the Unique Index constraint,
	// but we can check here for better error messaging.
	existing, _ := m.repo.GetActiveOrder(ctx, req.Token, req.IsBacktest, req.FirebaseUID)
	if existing != nil {
		return nil, fmt.Errorf("active order already exists for %s", req.Symbol)
	}

	// 2. Route to correct executor
	var executor OrderExecutor = m.liveExecutor
	if req.IsBacktest {
		executor = m.paperExecutor
	}

	// 3. Execute with Zerodha or Paper Engine
	orderID, err := executor.PlaceOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. Record in DB
	order := &models.Order{
		OrderID:         orderID,
		InstrumentToken: req.Token,
		Symbol:          req.Symbol,
		Side:            req.Side,
		Quantity:        req.Quantity,
		Status:          "OPEN",
		IsBacktest:      req.IsBacktest,
	}

	if err := m.repo.SaveOrder(ctx, order, req.FirebaseUID); err != nil {
		// If DB save fails after API call, this is a critical sync error
		return nil, fmt.Errorf("order placed but failed to log: %v", err)
	}

	return order, nil
}
