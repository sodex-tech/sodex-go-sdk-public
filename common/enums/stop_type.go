package enums

// StopType represents the type of stop order
type StopType int

const (
	StopTypeUnknown    StopType = iota // Unknown
	StopTypeStopLoss                   // Stop loss order
	StopTypeTakeProfit                 // Take profit order
)

// String returns the string representation of StopType
func (t StopType) String() string {
	switch t {
	case StopTypeStopLoss:
		return "STOP_LOSS"
	case StopTypeTakeProfit:
		return "TAKE_PROFIT"
	default:
		return "UNKNOWN"
	}
}

// ParseStopType parses a string into StopType
func ParseStopType(s string) StopType {
	switch s {
	case "STOP_LOSS":
		return StopTypeStopLoss
	case "TAKE_PROFIT":
		return StopTypeTakeProfit
	default:
		return StopTypeUnknown
	}
}
