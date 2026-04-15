package types

import (
	"github.com/shopspring/decimal"
)

const ModifyOrderRequestTypeName = "modifyOrder"

// ModifyOrderRequest represents a single-order modify request for the Bolt
// perpetuals engine. At least one of OrderID or ClOrdID must be supplied to
// identify the resting order; providing both is rejected. At least one of
// Price, Quantity, or StopPrice must be non-nil for the modification to take
// effect.
type ModifyOrderRequest struct {
	// AccountID is the unique identifier for the account.
	AccountID uint64 `json:"accountID"`
	// SymbolID is the unique identifier for the symbol.
	SymbolID uint64 `json:"symbolID"`
	// OrderID is the exchange-assigned ID of the order to modify.
	// Exactly one of OrderID / ClOrdID must be set.
	OrderID *uint64 `json:"orderID,omitempty"`
	// ClOrdID is the client-assigned ID of the order to modify.
	// Exactly one of OrderID / ClOrdID must be set.
	ClOrdID *string `json:"clOrdID,omitempty"`
	// Price is the new limit price (optional — only include if changing).
	Price *decimal.Decimal `json:"price,omitempty"`
	// Quantity is the new order quantity (optional — only include if changing).
	Quantity *decimal.Decimal `json:"quantity,omitempty"`
	// StopPrice is the new stop price (optional — only include if changing).
	StopPrice *decimal.Decimal `json:"stopPrice,omitempty"`
}

// ActionName returns the EIP-712 action type discriminator for modify-order.
func (req *ModifyOrderRequest) ActionName() string {
	return ModifyOrderRequestTypeName
}
