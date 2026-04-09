package models

import "time"

type OrderRequest struct {
	Token       uint32  `json:"token"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"` // BUY/SELL
	OrderType   string  `json:"order_type"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	TakeProfit  float64 `json:"take_profit"`
	StopLoss    float64 `json:"stop_loss"`
	IsBacktest  bool    `json:"is_backtest"`
	FirebaseUID string  `json:"-"` // Set from Auth context
}

type Order struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"order_id"`
	InstrumentToken uint32    `json:"instrument_token"`
	Symbol          string    `json:"symbol"`
	Status          string    `json:"status"`
	Side            string    `json:"side"`
	Quantity        int       `json:"quantity"`
	IsBacktest      bool      `json:"is_backtest"`
	CreatedAt       time.Time `json:"created_at"`
}
