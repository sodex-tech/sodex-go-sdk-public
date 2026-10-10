package types

type AccountTrade struct {
	Symbol     string  `json:"symbol"`
	TradeID    uint64  `json:"tradeID"`
	OrderID    uint64  `json:"orderID"`
	ClOrdID    string  `json:"clOrdID"`
	Side       string  `json:"side"`
	Price      string  `json:"price"`
	Quantity   string  `json:"quantity"`
	Fee        string  `json:"fee"`
	BuilderFee *string `json:"builderFee,omitempty"`
	FeeCoin    string  `json:"feeCoin"`
	Timestamp  uint64  `json:"time"`
	IsMaker    bool    `json:"isMaker"`
}
