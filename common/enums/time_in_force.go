package enums

// TimeInForce represents how long an order remains active
type TimeInForce int

const (
	TimeInForceUnknown TimeInForce = iota // Unknown (0)
	TimeInForceGTC                        // Good Till Cancel (1)
	TimeInForceFOK                        // Fill or Kill (2)
	TimeInForceIOC                        // Immediate or Cancel (3)
	TimeInForceGTX                        // Good Till Crossing/Post-only (4)
)

// String returns the string representation of TimeInForce
func (t TimeInForce) String() string {
	switch t {
	case TimeInForceGTC:
		return "GTC"
	case TimeInForceIOC:
		return "IOC"
	case TimeInForceFOK:
		return "FOK"
	case TimeInForceGTX:
		return "GTX"
	default:
		return "UNKNOWN"
	}
}

// ParseTimeInForce parses a string into TimeInForce
func ParseTimeInForce(s string) TimeInForce {
	switch s {
	case "GTC":
		return TimeInForceGTC
	case "IOC":
		return TimeInForceIOC
	case "FOK":
		return TimeInForceFOK
	case "GTX":
		return TimeInForceGTX
	default:
		return TimeInForceUnknown
	}
}
