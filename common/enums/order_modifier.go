package enums

// OrderModifier represents the different ways an order can be modified
type OrderModifier int

const (
	OrderModifierUnknown      OrderModifier = iota // Unknown
	OrderModifierNormal                            // Normal order
	OrderModifierStop                              // Stop order
	OrderModifierBracket                           // Bracket order (primary order with attached TP/SL orders)
	OrderModifierAttachedStop                      // Stop order attached to a primary order
)

// String returns the string representation of OrderModifier
func (m OrderModifier) String() string {
	switch m {
	case OrderModifierNormal:
		return "NORMAL"
	case OrderModifierStop:
		return "STOP"
	case OrderModifierBracket:
		return "BRACKET"
	case OrderModifierAttachedStop:
		return "ATTACHED_STOP"
	default:
		return "UNKNOWN"
	}
}

// ParseOrderModifier parses a string into OrderModifier
func ParseOrderModifier(s string) OrderModifier {
	switch s {
	case "NORMAL":
		return OrderModifierNormal
	case "STOP":
		return OrderModifierStop
	case "BRACKET":
		return OrderModifierBracket
	case "ATTACHED_STOP":
		return OrderModifierAttachedStop
	default:
		return OrderModifierUnknown
	}
}
