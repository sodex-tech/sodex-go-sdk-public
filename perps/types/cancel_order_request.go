package types

const CancelOrderRequestTypeName = "cancelOrder"

type CancelOrder struct {
	SymbolID uint64  `json:"symbolID"`
	OrderID  *uint64 `json:"orderID,omitempty"`
	ClOrdID  *string `json:"clOrdID,omitempty"`
}

type CancelOrderRequest struct {
	AccountID uint64         `json:"accountID"`
	Cancels   []*CancelOrder `json:"cancels"`
}

func (req *CancelOrderRequest) ActionName() string {
	return CancelOrderRequestTypeName
}
