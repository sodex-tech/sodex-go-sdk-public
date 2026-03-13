package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk/common/enums"
)

const NewOrderRequestTypeName = "newOrder"

// RawOrder represents a single order in a batch request
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
