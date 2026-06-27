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

// FetchVCNPayload collects completed trades and itemizes charges filtered by user and trading date
func (r *OrderRepository) FetchVCNPayload(ctx context.Context, userEmail string, dateStr string) (*models.VirtualContractNotePayload, error) {
	var query string
	var rows *sql.Rows
	var err error

	// If no date is passed from UI, default to CURRENT_DATE matching your schema default
	if dateStr == "" {
		query = `
            SELECT order_id, symbol, product, side, quantity, price, timestamp 
            FROM gidh_orders 
            WHERE user_email = $1 AND status = 'COMPLETE' AND trading_date = CURRENT_DATE
            ORDER BY timestamp ASC`
		rows, err = r.db.QueryContext(ctx, query, userEmail)
	} else {
		query = `
            SELECT order_id, symbol, product, side, quantity, price, timestamp 
            FROM gidh_orders 
            WHERE user_email = $1 AND status = 'COMPLETE' AND trading_date = $2
            ORDER BY timestamp ASC`
		rows, err = r.db.QueryContext(ctx, query, userEmail, dateStr)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payload models.VirtualContractNotePayload
	payload.Trades = make([]models.OrderRecord, 0)

	for rows.Next() {
		var o models.OrderRecord
		if err := rows.Scan(&o.OrderID, &o.Symbol, &o.Product, &o.Side, &o.Quantity, &o.Price, &o.Timestamp); err != nil {
			return nil, err
		}

		// Apply standard NSE contract transaction cost formulas
		turnover := float64(o.Quantity) * o.Price

		brokerage := turnover * 0.0003
		if brokerage > 20.0 {
			brokerage = 20.0
		}
		stt := turnover * 0.00025
		exchangeFees := turnover * 0.0000345
		sebiOverhead := turnover * 0.000001

		// Stamp duty is only charged on the BUY side
		stampDuty := 0.0
		if o.Side == "BUY" {
			stampDuty = turnover * 0.00003 // 0.003% for Intraday (MIS)
		}

		gst := (brokerage + exchangeFees + sebiOverhead) * 0.18
		totalCharges := brokerage + stt + exchangeFees + sebiOverhead + gst

		o.Charges = totalCharges
		payload.Trades = append(payload.Trades, o)

		// Accumulate Global Session Summary
		payload.Summary.Brokerage += brokerage
		payload.Summary.STT += stt
		payload.Summary.StampDuty += stampDuty
		payload.Summary.ExchangeTurnoverCharge += exchangeFees
		payload.Summary.SebiTurnoverCharge += sebiOverhead
		payload.Summary.GST += gst
		payload.Summary.TotalCharges += totalCharges
	}

	return &payload, nil
}
