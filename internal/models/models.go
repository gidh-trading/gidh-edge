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
	Severity  float64   `json:"severity"` // gidh_score
	Message   string    `json:"message"`
}

type VolumeProfile struct {
	Buckets     map[string]int64 `json:"buckets"`
	POC         float64          `json:"poc"`
	VAH         float64          `json:"vah"`
	VAL         float64          `json:"val"`
	TotalVolume int64            `json:"total_volume"`
}

type Snapshot struct {
	HistoryBars      []Bar         `json:"history_bars"`
	HistoryAnomalies []Anomaly     `json:"history_anomalies"`
	ActiveBars       []Bar         `json:"active_bars"`
	VolumeProfile    VolumeProfile `json:"volume_profile"`
}

type TimeBucketDNA struct {
	TimeKey     string  `json:"time_key"`
	MedianVol   float64 `json:"median_vol"`
	Surge99th   float64 `json:"surge_99th"`
	MedianRange float64 `json:"median_range"`
}

type Baseline struct {
	Token       uint32          `json:"token"`
	Symbol      string          `json:"symbol"`
	Date        time.Time       `json:"date"`
	POC         float64         `json:"poc_5d"`
	VAH         float64         `json:"vah_5d"`
	VAL         float64         `json:"val_5d"`
	MacroHVNs   []float64       `json:"macro_hvns"`
	TimeBuckets []TimeBucketDNA `json:"time_buckets"`
}

type JSONResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
