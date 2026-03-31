package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk/common/enums"
)

const NewOrderRequestTypeName = "newOrder"

// RawOrder represents a single order in a batch request.
//
// IMPORTANT for non-Go SDK implementors:
// The server verifies signatures by parsing the request body, wrapping it in
// ActionPayload{Type, Params}, and re-marshaling via json.Marshal to recompute
// the payloadHash. Since json.Marshal serializes in struct field order, your
// signing payload JSON keys MUST appear in the exact order defined below.
//
// Additionally, the server's API struct uses string types for decimal fields
// (Price, Quantity, Funds, StopPrice), so these must be JSON strings in both
// the signing payload and HTTP body (e.g. "quantity":"0.001", not "quantity":0.001).
//
// Fields without "omitempty" (Modifier, ReduceOnly, PositionSide) must always be
// present. Fields with "omitempty" must be omitted when unset.
type RawOrder struct {
	// Basic identification
	ClOrdID string `json:"clOrdID"` // Client Order ID

	// Order parameters
	Modifier    enums.OrderModifier `json:"modifier"`
	Side        enums.OrderSide     `json:"side"`        // Buy or Sell
	Type        enums.OrderType     `json:"type"`        // Limit or Market
	TimeInForce enums.TimeInForce   `json:"timeInForce"` // GTC, FOK, IOC, GTX

	// Price and quantity
	Price    *decimal.Decimal `json:"price,omitempty"`    // Order price (required for limit orders)
	Quantity *decimal.Decimal `json:"quantity,omitempty"` // Order quantity in contracts
	Funds    *decimal.Decimal `json:"funds,omitempty"`    // Total funds to use (alternative to quantity for market orders)

	// Stop order parameters
	StopPrice   *decimal.Decimal   `json:"stopPrice,omitempty"`   // Stop price for stop orders
	StopType    *enums.StopType    `json:"stopType,omitempty"`    // Type of stop order (stop loss or take profit)
	TriggerType *enums.TriggerType `json:"triggerType,omitempty"` // Price type used to trigger stop orders

	// Perps-specific fields
	ReduceOnly   bool               `json:"reduceOnly"`   // Only reduce position size
	PositionSide enums.PositionSide `json:"positionSide"` // Position side (BOTH, LONG, SHORT)
}

type NewOrderRequest struct {
	AccountID uint64      `json:"accountID"`
	SymbolID  uint64      `json:"symbolID"`
	Orders    []*RawOrder `json:"orders"`
}

func (req *NewOrderRequest) ActionName() string {
	return NewOrderRequestTypeName
}
