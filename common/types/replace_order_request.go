package types

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

const ReplaceOrderRequestTypeName = "replaceOrder"

// ReplaceParams holds essential fields for replacing an order.
type ReplaceParams struct {
	// SymbolID is the unique identifier for the symbol.
	SymbolID uint64 `json:"symbolID"`
	// ClOrdID is the unique identifier for this replace request.
	ClOrdID string `json:"clOrdID"`
	// OrigOrderID is the unique identifier of the order to be replaced.
	OrigOrderID *uint64 `json:"origOrderID,omitempty"`
	// OrigClOrdID is the client-provided identifier of the order to be replaced.
	OrigClOrdID *string `json:"origClOrdID,omitempty"`
	// Price is the new price for the order (optional, only if changing price).
	Price *decimal.Decimal `json:"price,omitempty"`
	// Quantity is the new quantity for the order (optional, only if changing quantity).
	Quantity *decimal.Decimal `json:"quantity,omitempty"`
}

type ReplaceOrderRequest struct {
	AccountID uint64           `json:"accountID"`
	Orders    []*ReplaceParams `json:"orders"`
}

func (req *ReplaceOrderRequest) ActionName() string {
	return ReplaceOrderRequestTypeName
}

// ToBytes converts the ReplaceOrderRequest to bytes
func (req *ReplaceOrderRequest) ToBytes() ([]byte, error) {
	return json.Marshal(req)
}

// FromBytes populates ReplaceOrderRequest from JSON bytes
func (req *ReplaceOrderRequest) FromBytes(data []byte) error {
	return json.Unmarshal(data, req)
}
