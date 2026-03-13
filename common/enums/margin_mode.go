package enums

// MarginMode represents the margin mode for a position
type MarginMode int

const (
	MarginModeUnknown MarginMode = iota
	MarginModeIsolated
	MarginModeCross
)

// String returns the string representation of MarginMode
func (m MarginMode) String() string {
	switch m {
	case MarginModeIsolated:
		return "ISOLATED"
	case MarginModeCross:
		return "CROSS"
	default:
		return "UNKNOWN"
	}
}

// ParseMarginMode parses a string into MarginMode
func ParseMarginMode(s string) MarginMode {
	switch s {
	case "ISOLATED":
		return MarginModeIsolated
	case "CROSS":
		return MarginModeCross
	default:
		return MarginModeUnknown
	}
}
