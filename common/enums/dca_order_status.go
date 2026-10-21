package enums

type DcaOrderStatus int

const (
	DcaOrderStatusUnknown DcaOrderStatus = iota
	DcaOrderStatusActive
	DcaOrderStatusCancelled
	DcaOrderStatusTerminated
	DcaOrderStatusPruned
)

func (s DcaOrderStatus) String() string {
	switch s {
	case DcaOrderStatusActive:
		return "ACTIVE"
	case DcaOrderStatusCancelled:
		return "CANCELLED"
	case DcaOrderStatusTerminated:
		return "TERMINATED"
	case DcaOrderStatusPruned:
		return "PRUNED"
	default:
		return "UNKNOWN"
	}
}

func ParseDcaOrderStatus(s string) DcaOrderStatus {
	switch s {
	case "ACTIVE":
		return DcaOrderStatusActive
	case "CANCELLED":
		return DcaOrderStatusCancelled
	case "TERMINATED":
		return DcaOrderStatusTerminated
	case "PRUNED":
		return DcaOrderStatusPruned
	default:
		return DcaOrderStatusUnknown
	}
}
