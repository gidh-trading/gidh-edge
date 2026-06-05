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

type BarAnalytics struct {
	VolumeRank int    `json:"volume_rank"`
	TickRank   int    `json:"tick_rank"`
	PriceRank  int    `json:"price_rank"`
	RangeRank  int    `json:"range_rank"`
	Direction  string `json:"direction"`
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

	TotalBuyQty  float64 `json:"total_buy_qty"`
	TotalSellQty float64 `json:"total_sell_qty"`
	ChangePct    float64 `json:"change_pct"`

	// Analytical Metadata Container
	Analytics BarAnalytics `json:"analytics"`
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

type PricePotential struct {
	StockName string  `json:"stock_name"`
	Interval  string  `json:"interval"`
	P97       float64 `json:"p97"`
	P90       float64 `json:"p90"`
	P75       float64 `json:"p75"`
	P50       float64 `json:"p50"`
	P25       float64 `json:"p25"`
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
	OrderID   string    `json:"order_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"` // BUY, SELL
	OrderType string    `json:"order_type"`
	Qty       int       `json:"qty"`
	FilledQty int       `json:"filled_qty"` // Explicitly snake_case for UI progress bar streams
	Price     float64   `json:"price"`
	Status    string    `json:"status"` // PENDING, COMPLETE, CANCELLED, REJECTED
	Timestamp time.Time `json:"timestamp"`
	UserEmail string    `json:"user_email,omitempty"`
}

type Position struct {
	TradingDate   time.Time `json:"trading_date"`
	Symbol        string    `json:"symbol"`
	Product       string    `json:"product"`
	Side          string    `json:"side"` // LONG, SHORT, or empty "" if flat
	NetQuantity   int       `json:"net_quantity"`
	AveragePrice  float64   `json:"average_price"`
	RealizedPnL   float64   `json:"realized_pnl"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`  // Computed dynamically per tick on backend
	TargetPrice   float64   `json:"target_price"`    // Syncs visual chart target boundaries
	StopLossPrice float64   `json:"stop_loss_price"` // Syncs visual chart floor boundaries
}

// internal/models/models.go

type MockTrade struct {
	Symbol          string  `json:"symbol"`
	Exchange        string  `json:"exchange"`         // Default "NSE"
	TransactionType string  `json:"transaction_type"` // BUY or SELL
	Product         string  `json:"product"`          // CNC or MIS
	OrderType       string  `json:"order_type"`       // MARKET or LIMIT
	Quantity        int     `json:"quantity"`
	Price           float64 `json:"price"`
}

type VirtualContractNoteRequest struct {
	Trades []MockTrade `json:"trades"`
}

type EnrichedMockTrade struct {
	Timestamp       time.Time `json:"timestamp"`
	Side            string    `json:"side"`
	Symbol          string    `json:"symbol"`
	Exchange        string    `json:"exchange"`
	Quantity        int       `json:"quantity"`
	AveragePrice    float64   `json:"average_price"`
	AllocatedCharge float64   `json:"allocated_charge"` // Total transactional friction for this leg
}

type VirtualContractNoteResponse struct {
	Summary struct {
		Brokerage              float64 `json:"brokerage"`
		STT                    float64 `json:"stt"`
		StampDuty              float64 `json:"stamp_duty"`
		ExchangeTurnoverCharge float64 `json:"exchange_turnover_charge"`
		SEBITurnoverCharge     float64 `json:"sebi_turnover_charge"`
		GST                    float64 `json:"gst"`
		TotalCharges           float64 `json:"total_charges"`
	} `json:"summary"`
	Trades []EnrichedMockTrade `json:"trades"`
}
