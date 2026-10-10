package types

type MarkPrice struct {
	Symbol          string `json:"s"`
	OpenInterest    string `json:"oi"`
	MarkPrice       string `json:"p"`
	IndexPrice      string `json:"i"`
	FundingRate     string `json:"r"`
	NextFundingTime uint64 `json:"T"`
}

type WsMarkPrice struct {
	EventTime uint64 `json:"E"`
	*MarkPrice
}
