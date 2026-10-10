package types

import "github.com/shopspring/decimal"

type Candle struct {
	StartTimestamp  uint64 `json:"t"`
	FinishTimestamp uint64 `json:"T"`

	Symbol   string `json:"s"`
	Interval string `json:"i"`

	OpenPrice  string `json:"o"`
	HighPrice  string `json:"h"`
	LowPrice   string `json:"l"`
	ClosePrice string `json:"c"`

	BaseVolume  decimal.Decimal `json:"v"`
	QuoteVolume decimal.Decimal `json:"q"`
	Trades      uint64          `json:"n"`

	Closed bool `json:"x"`
}
