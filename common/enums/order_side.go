package enums

// OrderSide represents the side of an order
type OrderSide int

const (
	OrderSideUnknown OrderSide = iota // Unknown side
	OrderSideBuy                      // Buy order
	OrderSideSell                     // Sell order
)

// String returns the string representation of OrderSide
func (s OrderSide) String() string {
	switch s {
	case OrderSideBuy:
		return "BUY"
	case OrderSideSell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// Opposite returns the opposite side of the order
func (s OrderSide) Opposite() OrderSide {
	switch s {
	case OrderSideBuy:
		return OrderSideSell
	case OrderSideSell:
		return OrderSideBuy
	default:
		return OrderSideUnknown
	}
}

// ParseOrderSide parses a string into OrderSide
func ParseOrderSide(s string) OrderSide {
	switch s {
	case "BUY":
		return OrderSideBuy
	case "SELL":
		return OrderSideSell
	default:
		return OrderSideUnknown
	}
}
