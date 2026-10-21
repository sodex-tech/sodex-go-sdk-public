package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
)

const NewDcaOrderRequestTypeName = "newDcaOrder"
const CancelDcaOrderRequestTypeName = "cancelDcaOrder"

// DcaSchedule uses UTC hours. Day is omitted for interval and daily plans.
type DcaSchedule struct {
	Kind string  `json:"kind"`
	Hour uint32  `json:"hour"`
	Day  *uint32 `json:"day,omitempty"`
}

// NewDcaOrderRequest creates a plan with Amount in USDC notional per execution.
// Field order and explicit false flags are part of the signed payload.
type NewDcaOrderRequest struct {
	AccountID uint64           `json:"accountID"`
	SymbolID  uint64           `json:"symbolID"`
	Side      enums.OrderSide  `json:"side"`
	Amount    decimal.Decimal  `json:"amount"`
	Schedule  DcaSchedule      `json:"schedule"`
	MinPrice  *decimal.Decimal `json:"minPrice,omitempty"`
	MaxPrice  *decimal.Decimal `json:"maxPrice,omitempty"`
	Randomize bool             `json:"randomize"`
	Immediate bool             `json:"immediate"`
}

func (req *NewDcaOrderRequest) ActionName() string {
	return NewDcaOrderRequestTypeName
}

// CancelDcaOrderRequest identifies the parent plan, not a child execution order.
type CancelDcaOrderRequest struct {
	AccountID uint64 `json:"accountID"`
	SymbolID  uint64 `json:"symbolID"`
	OrderID   uint64 `json:"orderID"`
}

func (req *CancelDcaOrderRequest) ActionName() string {
	return CancelDcaOrderRequestTypeName
}
