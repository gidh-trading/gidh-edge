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
	query := `
		SELECT order_id, symbol, product, side, order_type, quantity, 
		       filled_qty, price, status, timestamp, target_price, sl_price
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
		err := rows.Scan(
			&o.OrderID, &o.Symbol, &o.Product, &o.Side, &o.OrderType, &o.Qty,
			&o.FilledQty, &o.Price, &o.Status, &o.Timestamp, &o.TargetPrice, &o.StopLossPrice,
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
		SELECT trading_date, symbol, product, side, net_quantity, avg_price, realized_pnl
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
			&p.TradingDate, &p.Symbol, &p.Product, &p.Side, &p.NetQuantity, &p.AveragePrice, &p.RealizedPnL,
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
