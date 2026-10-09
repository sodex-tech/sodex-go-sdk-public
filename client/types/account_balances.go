package types

type SpotBalance struct {
	ID     uint64 `json:"id"`
	Coin   string `json:"coin"`
	Total  string `json:"total"`
	Locked string `json:"locked"`
}

type SpotAccountBalances struct {
	BlockTime   uint64         `json:"blockTime"`
	BlockHeight uint64         `json:"blockHeight"`
	Balances    []*SpotBalance `json:"balances"`
}

type PerpsBalance struct {
	ID          uint64  `json:"id"`
	Coin        string  `json:"coin"`
	Total       string  `json:"total"`
	Collateral  string  `json:"collateral"`
	MarginRatio string  `json:"marginRatio"`
	Price       *string `json:"price,omitempty"`
}

type PerpsAccountBalances struct {
	BlockTime   uint64          `json:"blockTime"`
	BlockHeight uint64          `json:"blockHeight"`
	Balances    []*PerpsBalance `json:"balances"`
}
