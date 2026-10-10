package types

type OrderBook struct {
	BlockTime   uint64 `json:"blockTime"`
	BlockHeight uint64 `json:"blockHeight"`

	UpdateID uint64     `json:"updateID"`
	Bids     [][]string `json:"bids"`
	Asks     [][]string `json:"asks"`
}
