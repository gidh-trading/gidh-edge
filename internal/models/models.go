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
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	VWAP      float64   `json:"vwap"`
	POC       float64   `json:"poc"`
	VAH       float64   `json:"vah"`
	VAL       float64   `json:"val"`
}

type Anomaly struct {
	PeriodStart     time.Time `json:"period_start"`
	LastUpdatedAt   time.Time `json:"last_updated_at"`
	Type            string    `json:"type"`
	UpgradeCount    int       `json:"upgrade_count"`
	EffortScore     float64   `json:"effort_score"`
	ResultScore     float64   `json:"result_score"`
	DivergenceScore float64   `json:"divergence_score"`
	PriceValue      float64   `json:"price_value"`
	PriceBaseline   float64   `json:"price_baseline"`
	DistPOCPct      float64   `json:"dist_poc_pct"`
	DistVAHPct      float64   `json:"dist_vah_pct"`
	DistVALPct      float64   `json:"dist_val_pct"`
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

type VolumeProfile struct {
	StockName       string    `json:"stock_name"`
	InstrumentToken uint32    `json:"instrument_token"`
	TradingDate     time.Time `json:"trading_date"`
	BucketSize      float64   `json:"bucket_size"`
	SortedPrices    []float64 `json:"sorted_prices"`
	POC             float64   `json:"poc"`
	VAH             float64   `json:"vah"`
	VAL             float64   `json:"val"`
	TotalVolume     int64     `json:"total_volume"`
	Nodes           []VPNode  `json:"nodes"`
}

type MarketDNA struct {
	InstrumentToken uint32          `json:"instrument_token"`
	Symbol          string          `json:"symbol"` // Maps to stock_name in DB
	TradingDate     time.Time       `json:"trading_date"`
	POC             float64         `json:"poc_5d"`
	VAH             float64         `json:"vah_5d"`
	VAL             float64         `json:"val_5d"`
	MacroHVNs       []VPNode        `json:"macro_hvns"`
	TimeBuckets     []TimeBucketDNA `json:"time_buckets"`
}

type TimeBucketDNA struct {
	TimeKey     string  `json:"time_key"`
	MedianVol   float64 `json:"median_vol"`
	MADVol      float64 `json:"mad_vol"`
	Surge99th   float64 `json:"surge_99th"`
	MedianRange float64 `json:"median_range"`
	MADRange    float64 `json:"mad_range"`
}

type Snapshot struct {
	HistoryBars      []Bar           `json:"history_bars"`
	HistoryAnomalies []Anomaly       `json:"history_anomalies"`
	ActiveBars       []Bar           `json:"active_bars"`
	MarketDNA        *MarketDNA      `json:"market_dna"`
	VolumeProfiles   []VolumeProfile `json:"volume_profiles"`
}

type JSONResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
