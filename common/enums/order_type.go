package enums

// OrderType represents the type of an order
type OrderType int

const (
	OrderTypeUnknown OrderType = iota // Unknown type
	OrderTypeLimit                    // Limit order
	OrderTypeMarket                   // Market order
)

// String returns the string representation of OrderType
func (t OrderType) String() string {
	switch t {
	case OrderTypeLimit:
		return "LIMIT"
	case OrderTypeMarket:
		return "MARKET"
	default:
		return "UNKNOWN"
	}
}

// ParseOrderType parses a string into OrderType
func ParseOrderType(s string) OrderType {
	switch s {
	case "LIMIT":
		return OrderTypeLimit
	case "MARKET":
		return OrderTypeMarket
	default:
		return OrderTypeUnknown
	}
}
