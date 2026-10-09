package types

import "github.com/shopspring/decimal"

type PerpsPosition struct {
	PositionID       uint64           `json:"i"`
	Symbol           string           `json:"s"`
	MarginMode       string           `json:"m"`
	PositionSide     string           `json:"ps"`
	PositionSize     decimal.Decimal  `json:"sz"`
	EntryPrice       decimal.Decimal  `json:"ep"`
	IsolatedMargin   *decimal.Decimal `json:"iw,omitempty"`
	CumOpenCost      decimal.Decimal  `json:"co"`
	CumTradingFee    decimal.Decimal  `json:"cf"`
	CumClosedSize    decimal.Decimal  `json:"cc"`
	ClosePrice       decimal.Decimal  `json:"cp"`
	MaxSize          decimal.Decimal  `json:"ms"`
	RealizedPnL      decimal.Decimal  `json:"cr"`
	UnrealizedPnL    decimal.Decimal  `json:"ur"`
	Leverage         uint32           `json:"l"`
	LiquidationPrice decimal.Decimal  `json:"lp"`

	CreatedAt *uint64 `json:"ct,omitempty"`
	UpdatedAt *uint64 `json:"ut,omitempty"`
}

type PerpsPositionLite struct {
	PositionID     uint64           `json:"i"`
	Symbol         string           `json:"s"`
	PositionSize   decimal.Decimal  `json:"sz"`
	EntryPrice     decimal.Decimal  `json:"ep"`
	IsolatedMargin *decimal.Decimal `json:"iw,omitempty"`
	PositionSide   string           `json:"ps"`
}

type PerpsLiquidatedPosition struct {
	Symbol           string           `json:"s"`
	PositionSide     string           `json:"ps"`
	PositionSize     decimal.Decimal  `json:"sz"`
	MarkPrice        decimal.Decimal  `json:"mp"`
	LiquidationPrice *decimal.Decimal `json:"lp,omitempty"`
}

type PerpsSymbolConfig struct {
	Symbol     string `json:"s"`
	Leverage   uint32 `json:"l"`
	MarginMode string `json:"m"`
}
