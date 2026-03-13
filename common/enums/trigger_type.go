package enums

// TriggerType represents the price type used to trigger a stop order
type TriggerType int

const (
	TriggerTypeUnknown    TriggerType = iota // Unknown (0)
	TriggerTypeLastPrice                     // Last trade price (1)
	TriggerTypeMarkPrice                     // Mark price (2)
	TriggerTypeIndexPrice                    // Index price (3)
)

// String returns the string representation of TriggerType
func (t TriggerType) String() string {
	switch t {
	case TriggerTypeLastPrice:
		return "LAST_PRICE"
	case TriggerTypeMarkPrice:
		return "MARK_PRICE"
	case TriggerTypeIndexPrice:
		return "INDEX_PRICE"
	default:
		return "UNKNOWN"
	}
}

// ParseTriggerType parses a string into TriggerType
func ParseTriggerType(s string) TriggerType {
	switch s {
	case "LAST_PRICE":
		return TriggerTypeLastPrice
	case "MARK_PRICE":
		return TriggerTypeMarkPrice
	case "INDEX_PRICE":
		return TriggerTypeIndexPrice
	default:
		return TriggerTypeUnknown
	}
}
