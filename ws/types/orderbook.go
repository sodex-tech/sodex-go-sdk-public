package types

type BookTicker struct {
	Symbol   string `json:"s"`
	UpdateID uint64 `json:"u"`
	AskPx    string `json:"a"`
	AskSz    string `json:"A"`
	BidPx    string `json:"b"`
	BidSz    string `json:"B"`
}

type WsBookTicker struct {
	EventTime uint64 `json:"E"`
	*BookTicker
}

type DepthSnapshot struct {
	Symbol   string     `json:"s"`
	UpdateID uint64     `json:"u"`
	Asks     [][]string `json:"a"`
	Bids     [][]string `json:"b"`
}

type DepthUpdate struct {
	Symbol        string     `json:"s"`
	FirstUpdateID uint64     `json:"U"`
	LastUpdateID  uint64     `json:"u"`
	Asks          [][]string `json:"a"`
	Bids          [][]string `json:"b"`
}

type WsDepthSnapshot struct {
	EventTime uint64 `json:"E"`
	*DepthSnapshot
}

type WsDepthUpdate struct {
	EventTime uint64 `json:"E"`
	*DepthUpdate
}
