package types

type TwapOrder struct {
	Symbol        string `json:"symbol"`
	OrderID       uint64 `json:"orderID"`
	Quantity      string `json:"quantity"`
	Side          string `json:"side"`
	Minutes       uint64 `json:"minutes"`
	Randomize     bool   `json:"randomize"`
	ReduceOnly    bool   `json:"reduceOnly"`
	ExecutedQty   string `json:"executedQty"`
	ExecutedValue string `json:"executedValue"`
	CreatedAt     uint64 `json:"createdAt"`
	NextActiveAt  uint64 `json:"nextActiveAt"`
	Active        bool   `json:"active"`
}
