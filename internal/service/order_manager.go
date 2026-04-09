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

func (m *OrderManager) SubmitOrder(ctx context.Context, req models.OrderRequest, uid string) (*models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Check for existing active orders
	existing, _ := m.repo.GetActiveOrder(ctx, req.Token, req.IsBacktest, uid)
	if existing != nil {
		return nil, fmt.Errorf("active order already exists for %s", req.Symbol)
	}

	// 2. Select Executor
	var executor OrderExecutor = m.liveExecutor
	if req.IsBacktest {
		executor = m.paperExecutor
	}

	// 3. Execute
	orderID, err := executor.PlaceOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. Save to DB
	order := &models.Order{
		OrderID:         orderID,
		InstrumentToken: req.Token,
		Symbol:          req.Symbol,
		Side:            req.Side,
		Quantity:        req.Quantity,
		Status:          "OPEN",
		IsBacktest:      req.IsBacktest,
	}

	if err := m.repo.SaveOrder(ctx, order, uid); err != nil {
		return nil, fmt.Errorf("order placed but failed to log: %v", err)
	}

	return order, nil
}

func (m *OrderManager) GetActiveOrders(ctx context.Context, isBacktest bool, uid string) ([]models.Order, error) {
	return m.repo.GetActiveOrders(ctx, isBacktest, uid)
}
