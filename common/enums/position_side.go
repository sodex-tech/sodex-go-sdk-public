package enums

// PositionSide represents the side of a position
type PositionSide int

const (
	PositionSideUnknown PositionSide = iota
	PositionSideBoth
	PositionSideLong
	PositionSideShort
)

// String returns the string representation of PositionSide
func (s PositionSide) String() string {
	switch s {
	case PositionSideBoth:
		return "BOTH"
	case PositionSideLong:
		return "LONG"
	case PositionSideShort:
		return "SHORT"
	default:
		return "UNKNOWN"
	}
}

// ParsePositionSide parses a string into PositionSide
func ParsePositionSide(s string) PositionSide {
	switch s {
	case "BOTH":
		return PositionSideBoth
	case "LONG":
		return PositionSideLong
	case "SHORT":
		return PositionSideShort
	default:
		return PositionSideUnknown
	}
}
