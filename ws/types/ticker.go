package types

import "github.com/shopspring/decimal"

type Ticker struct {
	Symbol      string          `json:"s"`
	ClosePx     decimal.Decimal `json:"c"`
	LastSz      decimal.Decimal `json:"Q"`
	VWAP        decimal.Decimal `json:"w"`
	AskPx       decimal.Decimal `json:"a"`
	AskSz       decimal.Decimal `json:"A"`
	BidPx       decimal.Decimal `json:"b"`
	BidSz       decimal.Decimal `json:"B"`
	Change      decimal.Decimal `json:"p"`
	ChangePct   float64         `json:"P"`
	OpenPx      decimal.Decimal `json:"o"`
	HighPx      decimal.Decimal `json:"h"`
	LowPx       decimal.Decimal `json:"l"`
	BaseVolume  decimal.Decimal `json:"v"`
	QuoteVolume decimal.Decimal `json:"q"`
	OpenTime    uint64          `json:"O"`
	CloseTime   uint64          `json:"C"`
}

type WsTicker struct {
	EventTime uint64 `json:"E"`

	*Ticker
}

type MiniTicker struct {
	Symbol      string          `json:"s"`
	ClosePx     decimal.Decimal `json:"c"`
	OpenPx      decimal.Decimal `json:"o"`
	HighPx      decimal.Decimal `json:"h"`
	LowPx       decimal.Decimal `json:"l"`
	BaseVolume  decimal.Decimal `json:"v"`
	QuoteVolume decimal.Decimal `json:"q"`
}

type WsMiniTicker struct {
	EventTime uint64 `json:"E"`

	*MiniTicker
}
