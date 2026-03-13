package enums

type ExecType int

const (
	ExecTypeUnknown         ExecType = iota // Unknown type
	ExecTypeNew                             // New order
	ExecTypePartiallyFilled                 // Partially filled
	ExecTypeFilled                          // Filled
	ExecTypeCanceled                        // Canceled
	ExecTypeRejected                        // Rejected
	ExecTypeModified                        // Modified
	ExecTypeExpired                         // Expired
	ExecTypeReplaced                        // Replaced
)

// String returns the string representation of ExecType
func (e ExecType) String() string {
	switch e {
	case ExecTypeNew:
		return "NEW"
	case ExecTypePartiallyFilled:
		return "PARTIALLY_FILLED"
	case ExecTypeFilled:
		return "FILLED"
	case ExecTypeCanceled:
		return "CANCELED"
	case ExecTypeRejected:
		return "REJECTED"
	case ExecTypeModified:
		return "MODIFIED"
	case ExecTypeExpired:
		return "EXPIRED"
	case ExecTypeReplaced:
		return "REPLACED"
	default:
		return "UNKNOWN"
	}
}

// ParseExecType parses a string into ExecType
func ParseExecType(s string) ExecType {
	switch s {
	case "NEW":
		return ExecTypeNew
	case "PARTIALLY_FILLED":
		return ExecTypePartiallyFilled
	case "FILLED":
		return ExecTypeFilled
	case "CANCELED":
		return ExecTypeCanceled
	case "REJECTED":
		return ExecTypeRejected
	case "MODIFIED":
		return ExecTypeModified
	case "EXPIRED":
		return ExecTypeExpired
	case "REPLACED":
		return ExecTypeReplaced
	default:
		return ExecTypeUnknown
	}
}
