package types

import "github.com/shopspring/decimal"

type SpotBalance struct {
	CoinID   uint64          `json:"i"`
	CoinName string          `json:"a"`
	Total    decimal.Decimal `json:"t"`
	Locked   decimal.Decimal `json:"l"`
}

type PerpsBalance struct {
	CoinID        uint64          `json:"i"`
	CoinName      string          `json:"a"`
	WalletBalance decimal.Decimal `json:"wb"`
	Collateral    decimal.Decimal `json:"co"`
	MarginRatio   decimal.Decimal `json:"mr"`
	Price         decimal.Decimal `json:"px"`
}

type PerpsBalanceLite struct {
	CoinID        uint64          `json:"i"`
	CoinName      string          `json:"a"`
	WalletBalance decimal.Decimal `json:"wb"`
	Collateral    decimal.Decimal `json:"co"`
}

type PerpsBalanceDetailed struct {
	*PerpsBalance

	IsolatedMargin    decimal.Decimal `json:"iw"`
	AvailableBalance  decimal.Decimal `json:"aw"`
	AvailableTransfer decimal.Decimal `json:"at"`
	WalletMargin      decimal.Decimal `json:"wm"`
	AvailableMargin   decimal.Decimal `json:"am"`
}
