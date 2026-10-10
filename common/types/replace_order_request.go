package types

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

const ReplaceOrderRequestTypeName = "replaceOrder"

type ReplaceParams struct {
	SymbolID    uint64           `json:"symbolID"`
	ClOrdID     string           `json:"clOrdID"`
	OrigOrderID *uint64          `json:"origOrderID,omitempty"`
	OrigClOrdID *string          `json:"origClOrdID,omitempty"`
	Price       *decimal.Decimal `json:"price,omitempty"`
	Quantity    *decimal.Decimal `json:"quantity,omitempty"`
}

type ReplaceOrderRequest struct {
	AccountID uint64           `json:"accountID"`
	Orders    []*ReplaceParams `json:"orders"`
}

func (req *ReplaceOrderRequest) ActionName() string {
	return ReplaceOrderRequestTypeName
}

func (req *ReplaceOrderRequest) ToBytes() ([]byte, error) {
	return json.Marshal(req)
}

func (req *ReplaceOrderRequest) FromBytes(data []byte) error {
	return json.Unmarshal(data, req)
}
