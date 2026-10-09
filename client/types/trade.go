package types

type Trade struct {
	TradeID   uint64  `json:"t"`
	TradeTime uint64  `json:"T"`
	Symbol    string  `json:"s"`
	Side      string  `json:"S"`
	Price     string  `json:"p"`
	Quantity  string  `json:"q"`
	Buyer     *uint64 `json:"bi,omitempty"`
	Seller    *uint64 `json:"si,omitempty"`
}
