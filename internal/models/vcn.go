package models

import "time"

type ItemizedCharges struct {
	Brokerage              float64 `json:"brokerage"`
	STT                    float64 `json:"stt"`
	StampDuty              float64 `json:"stamp_duty"`
	ExchangeTurnoverCharge float64 `json:"exchange_turnover_charge"`
	SebiTurnoverCharge     float64 `json:"sebi_turnover_charge"`
	GST                    float64 `json:"gst"`
	TotalCharges           float64 `json:"total_charges"`
}

type OrderRecord struct {
	OrderID   string    `json:"order_id"`
	Symbol    string    `json:"symbol"`
	Product   string    `json:"product"`
	Side      string    `json:"side"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
	Charges   float64   `json:"allocated_charge"`
}

type VirtualContractNotePayload struct {
	Summary ItemizedCharges `json:"summary"`
	Trades  []OrderRecord   `json:"trades"`
}
