package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
)

const NewTwapOrderRequestTypeName = "newTwapOrder"
const CancelTwapOrderRequestTypeName = "cancelTwapOrder"

type NewTwapOrderRequest struct {
	AccountID  uint64          `json:"accountID"`
	SymbolID   uint64          `json:"symbolID"`
	Side       enums.OrderSide `json:"side"`
	Quantity   decimal.Decimal `json:"quantity"`
	Minutes    uint64          `json:"minutes"`
	Randomize  bool            `json:"randomize"`
	ReduceOnly *bool           `json:"reduceOnly,omitempty"`
}

func (req *NewTwapOrderRequest) ActionName() string {
	return NewTwapOrderRequestTypeName
}

type CancelTwapOrderRequest struct {
	AccountID uint64 `json:"accountID"`
	SymbolID  uint64 `json:"symbolID"`
	OrderID   uint64 `json:"orderID"`
}

func (req *CancelTwapOrderRequest) ActionName() string {
	return CancelTwapOrderRequestTypeName
}
