package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
)

const NewOrderRequestTypeName = "newOrder"

type RawOrder struct {
	ClOrdID string `json:"clOrdID"`

	Modifier    enums.OrderModifier `json:"modifier"`
	Side        enums.OrderSide     `json:"side"`
	Type        enums.OrderType     `json:"type"`
	TimeInForce enums.TimeInForce   `json:"timeInForce"`

	Price    *decimal.Decimal `json:"price,omitempty"`
	Quantity *decimal.Decimal `json:"quantity,omitempty"`
	Funds    *decimal.Decimal `json:"funds,omitempty"`

	StopPrice   *decimal.Decimal   `json:"stopPrice,omitempty"`
	StopType    *enums.StopType    `json:"stopType,omitempty"`
	TriggerType *enums.TriggerType `json:"triggerType,omitempty"`

	ReduceOnly   bool               `json:"reduceOnly"`
	PositionSide enums.PositionSide `json:"positionSide"`
}

type NewOrderRequest struct {
	AccountID uint64                `json:"accountID"`
	SymbolID  uint64                `json:"symbolID"`
	Orders    []*RawOrder           `json:"orders"`
	Builder   *ctypes.BuilderParams `json:"builder,omitempty"`
}

func (req *NewOrderRequest) ActionName() string {
	return NewOrderRequestTypeName
}
