package models

import (
	"encoding/json"
	"math"
	"time"
)

type Instrument struct {
	Token     uint32 `json:"token"`
	StockName string `json:"stock_name"`
}

type UIDominantAnomaly struct {
	IsPresent bool    `json:"is_present"`
	Type      string  `json:"type"` // "WHALE" or "ICEBERG"
	P         float64 `json:"p"`    // Price Bin level mapping
	V         float64 `json:"v"`    // Total Volume accumulated inside bucket
	D         float64 `json:"d"`    // Aggressive Volume net delta flow
	I         float64 `json:"i"`    // Volume weighted intensity footprint mapping
}

type AnomalyType int

type AnomalySnapshot struct {
	Timestamp  time.Time   `json:"ts"`
	Type       AnomalyType `json:"type"` //
	Direction  int         `json:"dir"`  // -1 = Sell, 1 = Buy
	VolumeRank int         `json:"vol_rank"`
	PriceRank  int         `json:"price_rank"`
}

// PeakAnomalyMetrics remains a strict Go struct for compiler safety,
// but serializes out as a flat JSON dictionary object for database/sockets.
type PeakAnomalyMetrics struct {
	PeakVolumeRank      int `json:"peak_volume_rank"`
	PeakPriceRank       int `json:"peak_price_rank"`
	PeakTickRank        int `json:"peak_tick_rank"`
	MaxAnomalyDirection int `json:"max_anomaly_direction"`
	MaxAbsorptionSignal int `json:"max_absorption_signal"`
}

type Bar struct {
	Timestamp       time.Time `json:"timestamp"`
	InstrumentToken int32     `json:"instrument_token"`
	StockName       string    `json:"stock_name"`
	Timeframe       string    `json:"timeframe"`

	// ---- Pure OHLC ----
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`

	// ---- Aggregated Quantities ----
	Volume    float64 `json:"volume"`
	TickCount int64   `json:"tick_count"`
	VWAP      float64 `json:"vwap"`

	// ---- Auction Framework Elements ----
	POC float64 `json:"poc"`
	VAH float64 `json:"vah"`
	VAL float64 `json:"val"`

	// ---- Dynamic Structural Strategy Blocks ----
	Peaks             PeakAnomalyMetrics `json:"peaks"`
	SignificantEvents []AnomalySnapshot  `json:"significant_events,omitempty"`

	TotalBuyQty  float64 `json:"total_buy_qty"`
	TotalSellQty float64 `json:"total_sell_qty"`
	ChangePct    float64 `json:"change_pct"`
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
	StockName       string          `json:"stock_name"`
	TradingDate     time.Time       `json:"trading_date"`
	POC             float64         `json:"poc_5d"`
	VAH             float64         `json:"vah_5d"`
	VAL             float64         `json:"val_5d"`
	MacroHVNs       []VPExtrema     `json:"macro_hvns"`
	MacroLVNs       []VPExtrema     `json:"macro_lvns"`
	TimeBuckets     []TimeBucketDNA `json:"time_buckets"`
}

type TimeBucketDNA struct {
	MinuteIndex int `json:"minute_index"`

	VolumeMean float64 `json:"volume_mean"`
	VolumeStd  float64 `json:"volume_std"`

	RangeMean float64 `json:"range_mean"`
	RangeStd  float64 `json:"range_std"`

	// Optional future extensions
	VolumeP95 float64 `json:"volume_p95,omitempty"`
	RangeP95  float64 `json:"range_p95,omitempty"`
}

type Snapshot struct {
	HistoryBars    []Bar           `json:"history_bars"`
	MarketDNA      *MarketDNA      `json:"market_dna"`
	VolumeProfiles []VolumeProfile `json:"volume_profiles"`
}

// JSONResponse is the standard Edge API response wrapper
type JSONResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// gidh-edge/internal/models/models.go

type OrderBookEntry struct {
	OrderID       string    `json:"order_id"`
	Symbol        string    `json:"symbol"`
	Product       string    `json:"product"`
	Side          string    `json:"side"`
	OrderType     string    `json:"order_type"`
	Qty           int       `json:"qty"`
	FilledQty     int       `json:"filled_qty"`
	Price         float64   `json:"price"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
	TargetPrice   float64   `json:"target_price,omitempty"`
	StopLossPrice float64   `json:"stop_loss_price,omitempty"`
	TradingDate   time.Time `json:"trading_date"`
	UserEmail     string    `json:"user_email,omitempty"`
}

type Position struct {
	TradingDate   time.Time `json:"trading_date"`
	Symbol        string    `json:"symbol"`
	Product       string    `json:"product"`
	Side          string    `json:"side"`
	NetQuantity   int       `json:"net_quantity"`
	AveragePrice  float64   `json:"average_price"` // 🧠 Fixed: changed from "avg_price"
	RealizedPnL   float64   `json:"realized_pnl"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`  // 🧠 Added field
	TargetPrice   float64   `json:"target_price"`    // 🧠 Added field
	StopLossPrice float64   `json:"stop_loss_price"` // 🧠 Added field
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}
