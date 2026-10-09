package types

import (
	"github.com/shopspring/decimal"
)

const ModifyOrderRequestTypeName = "modifyOrder"

type ModifyOrderRequest struct {
	AccountID uint64           `json:"accountID"`
	SymbolID  uint64           `json:"symbolID"`
	OrderID   *uint64          `json:"orderID,omitempty"`
	ClOrdID   *string          `json:"clOrdID,omitempty"`
	Price     *decimal.Decimal `json:"price,omitempty"`
	Quantity  *decimal.Decimal `json:"quantity,omitempty"`
	StopPrice *decimal.Decimal `json:"stopPrice,omitempty"`
}

func (req *ModifyOrderRequest) ActionName() string {
	return ModifyOrderRequestTypeName
}
