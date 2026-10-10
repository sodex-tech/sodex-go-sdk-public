package types

type Candle struct {
	StartTimestamp uint64 `json:"t"`

	OpenPrice  string `json:"o"`
	HighPrice  string `json:"h"`
	LowPrice   string `json:"l"`
	ClosePrice string `json:"c"`

	BaseVolume  string  `json:"v"`
	QuoteVolume string  `json:"q"`
	Trades      *uint64 `json:"n,omitempty"`
}
