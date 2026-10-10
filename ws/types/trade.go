package types

type Trade struct {
	TradeTime uint64 `json:"T"`
	TradeID   uint64 `json:"t"`
	Symbol    string `json:"s"`
	Side      string `json:"S"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	Buyer     uint64 `json:"bi"`
	Seller    uint64 `json:"si"`
}

type WsTrade struct {
	EventTime uint64 `json:"E"`
	*Trade
}

type UserTrade struct {
	TradeTime  uint64  `json:"T"`
	TradeID    uint64  `json:"t"`
	Symbol     string  `json:"s"`
	OrderID    uint64  `json:"i"`
	ClOrdID    string  `json:"c"`
	Side       string  `json:"S"`
	Price      string  `json:"p"`
	Quantity   string  `json:"q"`
	Fee        string  `json:"f"`
	BuilderFee *string `json:"bf,omitempty"`
	IsMaker    bool    `json:"m"`
}

type UserPerpsTrade struct {
	*UserTrade

	Dir string `json:"d"`
}

type WsUserSpotTrade struct {
	EventTime uint64 `json:"E"`
	*UserTrade
}

type WsUserPerpsTrade struct {
	EventTime uint64 `json:"E"`
	*UserPerpsTrade
}
