package types

import "github.com/shopspring/decimal"

const (
	UserEventLiquidation = "liquidation"
)

type WsUserLiquidation struct {
	EventTime uint64 `json:"E"`

	Type string `json:"type"`

	Liquidator   uint64                     `json:"lid"`
	AccountID    uint64                     `json:"aid"`
	AccountValue decimal.Decimal            `json:"av"`
	MarginMode   string                     `json:"mm"`
	Balances     []*PerpsBalanceLite        `json:"B"`
	Positions    []*PerpsLiquidatedPosition `json:"P"`
}
