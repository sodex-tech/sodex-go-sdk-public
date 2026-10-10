package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
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

type BatchNewOrderRequest struct {
	AccountID uint64                `json:"accountID"`
	Orders    []*BatchNewOrderItem  `json:"orders"`
	Builder   *ctypes.BuilderParams `json:"builder,omitempty"`
}

func (req *BatchNewOrderRequest) ActionName() string {
	return BatchNewOrderRequestTypeName
}
