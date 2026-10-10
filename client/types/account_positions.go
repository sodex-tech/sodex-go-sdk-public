package types

type PerpsPosition struct {
	ID     uint64 `json:"id"`
	Symbol string `json:"symbol"`

	MarginMode   string `json:"marginMode"`
	PositionSide string `json:"positionSide"`

	Size          string `json:"size"`
	InitialMargin string `json:"initialMargin"`

	AvgEntryPrice string `json:"avgEntryPrice"`
	CumOpenCost   string `json:"cumOpenCost"`
	CumTradingFee string `json:"cumTradingFee"`
	CumClosedSize string `json:"cumClosedSize"`
	AvgClosePrice string `json:"avgClosePrice"`
	MaxSize       string `json:"maxSize"`
	RealizedPnL   string `json:"realizedPnL"`

	Leverage uint32 `json:"leverage"`
	Active   bool   `json:"active"`

	IsTakenOver   bool   `json:"isTakenOver"`
	TakeOverPrice string `json:"takeOverPrice"`

	CreatedAt uint64 `json:"createdAt"`
	UpdatedAt uint64 `json:"updatedAt"`
}

type PerpsAccountPositions struct {
	BlockTime   uint64           `json:"blockTime"`
	BlockHeight uint64           `json:"blockHeight"`
	Positions   []*PerpsPosition `json:"positions"`
}
