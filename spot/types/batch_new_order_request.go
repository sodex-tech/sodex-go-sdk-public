package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk/common/enums"
)

const BatchNewOrderRequestTypeName = "batchNewOrder"

type BatchNewOrderItem struct {
	SymbolID    uint64            `json:"symbolID"`
	ClOrdID     string            `json:"clOrdID"`
	Side        enums.OrderSide   `json:"side"`
	Type        enums.OrderType   `json:"type"`
	TimeInForce enums.TimeInForce `json:"timeInForce"`
	Price       *decimal.Decimal  `json:"price,omitempty"`
	Quantity    *decimal.Decimal  `json:"quantity,omitempty"`
	Funds       *decimal.Decimal  `json:"funds,omitempty"`
}

// BatchNewOrderRequest represents a batch new order request
type BatchNewOrderRequest struct {
	AccountID uint64               `json:"accountID"`
	Orders    []*BatchNewOrderItem `json:"orders"`
}

// ActionName returns the action name for the batch new order request
func (req *BatchNewOrderRequest) ActionName() string {
	return BatchNewOrderRequestTypeName
}
