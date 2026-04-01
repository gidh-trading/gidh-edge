package models

import "time"

type Instrument struct {
	Token  uint32 `json:"token"`
	Symbol string `json:"symbol"`
}

type Bar struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
}

type Anomaly struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
}

type Snapshot struct {
	History []Bar     `json:"history"`
	Active  []Bar     `json:"active"`
	Profile []float64 `json:"profile"` // Volume Profile
}

type Baseline struct {
	Token uint32    `json:"token"`
	VAH   float64   `json:"vah"` // Value Area High
	VAL   float64   `json:"val"` // Value Area Low
	POC   float64   `json:"poc"` // Point of Control
	Date  time.Time `json:"date"`
}

type HealthStatus struct {
	Database string `json:"database"`
	Engine   string `json:"engine"`
	Status   string `json:"status"` // "healthy" or "degraded"
}

// Standard Response Wrapper
type JSONResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
