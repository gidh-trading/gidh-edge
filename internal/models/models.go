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
	TradingDate    time.Time `json:"trading_date"`
	Symbol         string    `json:"symbol"`
	Product        string    `json:"product"`
	Side           string    `json:"side"` // LONG, SHORT, or empty "" if flat
	NetQuantity    int       `json:"net_quantity"`
	AveragePrice   float64   `json:"average_price"`
	LTP            float64   `json:"ltp"`
	RealizedPnL    float64   `json:"realized_pnl"`
	UnrealizedPnL  float64   `json:"unrealized_pnl"`  // Computed dynamically per tick on backend
	TargetPrice    float64   `json:"target_price"`    // Syncs visual chart target boundaries
	StopLossPrice  float64   `json:"stop_loss_price"` // Syncs visual chart floor boundaries
	EntryTimestamp string    `json:"entry_timestamp"`
	TimeElapsed    string    `json:"time_elapsed"` // ⚡ NEW: Tracking duration inside position

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

type InstrumentProfile struct {
	StockName       string    `json:"stock_name"`
	InstrumentToken uint32    `json:"instrument_token"`
	TradingDate     time.Time `json:"trading_date"`
	BucketSize      float64   `json:"bucket_size"`
	ATR14           float64   `json:"atr_14"`
	ADRPct          float64   `json:"adr_pct"`
	ADV30d          int64     `json:"adv_30d"`
	ADVVal30d       float64   `json:"adv_val_30d"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type VWAPDistancePercentiles struct {
	InstrumentToken uint32    `json:"instrument_token"`
	StockName       string    `json:"stock_name"`
	TradingDate     time.Time `json:"trading_date"`
	PosP50          float64   `json:"pos_p50"`
	PosP75          float64   `json:"pos_p75"`
	PosP90          float64   `json:"pos_p90"`
	PosP97          float64   `json:"pos_p97"`
	PosP99          float64   `json:"pos_p99"`
	NegP50          float64   `json:"neg_p50"`
	NegP75          float64   `json:"neg_p75"`
	NegP90          float64   `json:"neg_p90"`
	NegP97          float64   `json:"neg_p97"`
	NegP99          float64   `json:"neg_p99"`
	PosMax          float64   `json:"pos_max"`
	NegMax          float64   `json:"neg_max"`
	UpdatedAt       time.Time `json:"updated_at"`
}
