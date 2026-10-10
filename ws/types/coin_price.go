package types

type CoinPrice struct {
	CoinID      uint64 `json:"i"`
	Coin        string `json:"a"`
	Price       string `json:"p"`
	MarginRatio string `json:"mr"`
}

type WsCoinPrice struct {
	EventTime uint64 `json:"E"`

	*CoinPrice
}
