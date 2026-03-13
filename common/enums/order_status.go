package enums

type OrderStatus int

const (
	OrderStatusUnknown OrderStatus = iota
	OrderStatusNew
	OrderStatusPartiallyFilled
	OrderStatusFilled
	OrderStatusCanceled
	OrderStatusRejected
	OrderStatusExpired
	OrderStatusPendingNew
	OrderStatusPendingCancel
	OrderStatusPendingModify
	OrderStatusTriggered
	OrderStatusReplaced
	OrderStatusPendingReplace
)

// String returns the string representation of OrderStatus
func (o OrderStatus) String() string {
	switch o {
	case OrderStatusNew:
		return "NEW"
	case OrderStatusPartiallyFilled:
		return "PARTIALLY_FILLED"
	case OrderStatusFilled:
		return "FILLED"
	case OrderStatusCanceled:
		return "CANCELED"
	case OrderStatusRejected:
		return "REJECTED"
	case OrderStatusExpired:
		return "EXPIRED"
	case OrderStatusPendingNew:
		return "PENDING_NEW"
	case OrderStatusPendingCancel:
		return "PENDING_CANCEL"
	case OrderStatusPendingModify:
		return "PENDING_MODIFY"
	case OrderStatusTriggered:
		return "TRIGGERED"
	case OrderStatusReplaced:
		return "REPLACED"
	case OrderStatusPendingReplace:
		return "PENDING_REPLACE"
	default:
		return "UNKNOWN"
	}
}

// ParseOrderStatus parses a string into OrderStatus
func ParseOrderStatus(s string) OrderStatus {
	switch s {
	case "NEW":
		return OrderStatusNew
	case "PARTIALLY_FILLED":
		return OrderStatusPartiallyFilled
	case "FILLED":
		return OrderStatusFilled
	case "CANCELED":
		return OrderStatusCanceled
	case "REJECTED":
		return OrderStatusRejected
	case "EXPIRED":
		return OrderStatusExpired
	case "PENDING_NEW":
		return OrderStatusPendingNew
	case "PENDING_CANCEL":
		return OrderStatusPendingCancel
	case "PENDING_MODIFY":
		return OrderStatusPendingModify
	case "TRIGGERED":
		return OrderStatusTriggered
	case "REPLACED":
		return OrderStatusReplaced
	case "PENDING_REPLACE":
		return OrderStatusPendingReplace
	default:
		return OrderStatusUnknown
	}
}
