package types

type SpotTicker struct {
	Symbol string `json:"symbol"`

	LastPx string  `json:"lastPx"`
	LastSz *string `json:"lastSz,omitempty"`
	OpenPx string  `json:"openPx"`
	HighPx string  `json:"highPx"`
	LowPx  string  `json:"lowPx"`

	BaseVolume  string `json:"volume"`
	QuoteVolume string `json:"quoteVolume"`

	VWAP      *string `json:"vwap,omitempty"`
	Change    string  `json:"change"`
	ChangePct float64 `json:"changePct"`

	AskPx string `json:"askPx"`
	AskSz string `json:"askSz"`
	BidPx string `json:"bidPx"`
	BidSz string `json:"bidSz"`

	OpenTime  uint64 `json:"openTime"`
	CloseTime uint64 `json:"closeTime"`
}

type PerpsTicker struct {
	*SpotTicker

	FundingRate     string `json:"fundingRate"`
	NextFundingTime uint64 `json:"nextFundingTime"`
	IndexPrice      string `json:"indexPrice"`
	MarkPrice       string `json:"markPrice"`
	OpenInterest    string `json:"openInterest"`
}
