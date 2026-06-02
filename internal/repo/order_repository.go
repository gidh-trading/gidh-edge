package repo

import (
	"context"
	"database/sql"
	"time"

	"gidh-edge/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// GetHistoricalOrders fetches orders matching a specific trading date from the database
func (r *OrderRepository) GetHistoricalOrders(ctx context.Context, tradingDate time.Time) ([]models.OrderBookEntry, error) {
	// 1. Updated query matching the columns of your updated gidh_orders table
	query := `
		SELECT order_id, symbol, product, side, order_type, quantity, 
		       filled_qty, price, status, timestamp, user_email
		FROM gidh_orders
		WHERE trading_date = $1::date;`

	rows, err := r.db.QueryContext(ctx, query, tradingDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.OrderBookEntry
	for rows.Next() {
		var o models.OrderBookEntry

		var discardedProduct string

		err := rows.Scan(
			&o.OrderID, &o.Symbol, &discardedProduct, &o.Side, &o.OrderType, &o.Qty,
			&o.FilledQty, &o.Price, &o.Status, &o.Timestamp, &o.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetHistoricalPositions fetches position snapshots matching a specific trading date from the database
func (r *OrderRepository) GetHistoricalPositions(ctx context.Context, tradingDate time.Time) ([]models.Position, error) {
	query := `
		SELECT trading_date, symbol, product, side, net_quantity, avg_price, realized_pnl, target_price, stop_loss_price
		FROM gidh_positions
		WHERE trading_date = $1::date;`

	rows, err := r.db.QueryContext(ctx, query, tradingDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		var p models.Position
		err := rows.Scan(
			&p.TradingDate, &p.Symbol, &p.Product, &p.Side, &p.NetQuantity,
			&p.AveragePrice, &p.RealizedPnL, &p.TargetPrice, &p.StopLossPrice, // Scan values here
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}
