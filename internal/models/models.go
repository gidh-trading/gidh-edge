package models

import (
	"encoding/json"
	"math"
	"time"
)

type Instrument struct {
	Token  uint32 `json:"token"`
	Symbol string `json:"symbol"`
}

type Bar struct {
	Timestamp     time.Time `json:"timestamp"`
	Open          float64   `json:"open"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	Close         float64   `json:"close"`
	Volume        int64     `json:"volume"`
	CVD           float64   `json:"cvd"`
	CVDDivergence float64   `json:"cvd_divergence"` // -1 to 1 Heatmap value

	// Physics Scores
	EffortScore   float64 `json:"effort_score"`
	ResultScore   float64 `json:"result_score"`
	PulseScore    float64 `json:"pulse_score"`
	PeakTradeSign int     `json:"peak_trade_sign"`

	IsClosed bool `json:"is_closed"`

	TotalBuyQty             int64 `json:"total_buy_qty"`
	TotalSellQty            int64 `json:"total_sell_qty"`
	TotalExecutedBuyVolume  int64 `json:"total_executed_buy_volume"`
	TotalExecutedSellVolume int64 `json:"total_executed_sell_volume"`

	VWAP float64 `json:"vwap"`
	POC  float64 `json:"poc"`
	VAH  float64 `json:"vah"`
	VAL  float64 `json:"val"`
}

type Wall struct {
	Price    float64 `json:"price"`
	Quantity int64   `json:"quantity"`
	Orders   int     `json:"orders"`
	Side     string  `json:"side"` // "buy" or "sell"

	// --- Existing State ---
	IsConcrete bool `json:"is_concrete"`
	IsBroken   bool `json:"is_broken"`

	// --- NEW: Iceberg Tracking ---
	AbsorbedVolume int64 `json:"absorbed_volume"` // Total hidden volume eaten
	HitCount       int   `json:"hit_count"`       // Number of times aggressor hit and reloaded
	IsIceberg      bool  `json:"is_iceberg"`      // Flag for UI rendering
}

type AnomalyType string

type AnomalyEvent struct {
	TimeKey         string      `json:"time_key"`
	PeriodStart     time.Time   `json:"period_start"`
	Interval        string      `json:"interval"`
	LastUpdatedAt   time.Time   `json:"last_updated_at"`
	InstrumentToken uint32      `json:"instrument_token"`
	Symbol          string      `json:"symbol"`
	Type            AnomalyType `json:"type"`
	Direction       int         `json:"direction"`

	// Core Physics Triad
	EffortScore float64 `json:"effort_score"`
	ResultScore float64 `json:"result_score"`
	PulseScore  float64 `json:"pulse_score"`

	Intensity  float64 `json:"intensity"`
	PriceValue float64 `json:"price_value"`
}

type VPNode struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
}

func (v *VPNode) UnmarshalJSON(data []byte) error {
	type Alias VPNode
	aux := &struct {
		Volume float64 `json:"volume"`
		*Alias
	}{
		Alias: (*Alias)(v),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	v.Volume = int64(math.Round(aux.Volume))
	return nil
}

type VPExtrema struct {
	Price    float64 `json:"price"`
	Volume   int64   `json:"volume"`
	Strength float64 `json:"strength"` // 0-100 scale
}

type VolumeProfile struct {
	StockName       string      `json:"stock_name"`
	InstrumentToken uint32      `json:"instrument_token"`
	TradingDate     time.Time   `json:"trading_date"`
	POC             float64     `json:"poc"`
	VAH             float64     `json:"vah"`
	VAL             float64     `json:"val"`
	Nodes           []VPNode    `json:"nodes"`
	HVNs            []VPExtrema `json:"hvns"`
	LVNs            []VPExtrema `json:"lvns"`
}

type MarketDNA struct {
	InstrumentToken uint32          `json:"instrument_token"`
	Symbol          string          `json:"symbol"` // Maps to stock_name in DB
	TradingDate     time.Time       `json:"trading_date"`
	POC             float64         `json:"poc_5d"`
	VAH             float64         `json:"vah_5d"`
	VAL             float64         `json:"val_5d"`
	MacroHVNs       []VPExtrema     `json:"macro_hvns"`
	MacroLVNs       []VPExtrema     `json:"macro_lvns"`
	TimeBuckets     []TimeBucketDNA `json:"time_buckets"`
}

type TimeBucketDNA struct {
	TimeKey        string  `json:"time_key"`
	MedianVol      float64 `json:"median_vol"`
	MADVol         float64 `json:"mad_vol"`
	SurgeThreshold float64 `json:"surge_threshold"`
	MedianRange    float64 `json:"median_range"`
	MADRange       float64 `json:"mad_range"`
}

type Snapshot struct {
	HistoryBars      []Bar           `json:"history_bars"`
	HistoryAnomalies []AnomalyEvent  `json:"history_anomalies"`
	ActiveBars       []Bar           `json:"active_bars"`
	MarketDNA        *MarketDNA      `json:"market_dna"`
	VolumeProfiles   []VolumeProfile `json:"volume_profiles"`
}

// OrderRequest matches the backend Engine's expectation
type OrderRequest struct {
	Token       uint32  `json:"token"`
	Symbol      string  `json:"symbol"`
	TradingDate string  `json:"trading_date"`
	Side        string  `json:"side"`
	OrderType   string  `json:"order_type"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	StopLoss    float64 `json:"stop_loss"`
	TakeProfit  float64 `json:"take_profit"`
}

type ExitRequest struct {
	ID        string  `json:"id"`
	OrderType string  `json:"order_type"` // "MARKET" or "LIMIT"
	Price     float64 `json:"price"`      // Required if LIMIT
}

type Position struct {
	ID              string    `json:"id"`
	FirebaseUID     string    `json:"firebase_uid"`
	InstrumentToken uint32    `json:"instrument_token"`
	Symbol          string    `json:"symbol"`
	TradingDate     time.Time `json:"trading_date"`
	Side            string    `json:"side"`          // "BUY" or "SELL"
	Quantity        int       `json:"quantity"`      // Net quantity
	AveragePrice    float64   `json:"average_price"` // Weighted average entry
	Status          string    `json:"status"`        // "ACTIVE" or "CLOSED"
	RealizedPnL     float64   `json:"realized_pnl"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Order matches the updated DB schema from the backend
type Order struct {
	ID              string                 `json:"id"`
	OrderID         string                 `json:"order_id"`
	InstrumentToken uint32                 `json:"instrument_token"`
	Symbol          string                 `json:"symbol"`
	TradingDate     time.Time              `json:"trading_date"`
	FirebaseUID     string                 `json:"firebase_uid"`
	Side            string                 `json:"side"`
	OrderType       string                 `json:"order_type"`
	Quantity        int                    `json:"quantity"`
	Price           float64                `json:"price"`
	StopLoss        float64                `json:"stop_loss"`
	TakeProfit      float64                `json:"take_profit"`
	Status          string                 `json:"status"`
	EntryPrice      float64                `json:"entry_price"`
	ExitPrice       float64                `json:"exit_price"`
	Stats           map[string]interface{} `json:"stats"`
	CreatedAt       time.Time              `json:"created_at"`
}

// JSONResponse is the standard Edge API response wrapper
type JSONResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
