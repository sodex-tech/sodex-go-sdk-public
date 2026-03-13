package types

const BatchCancelOrderRequestTypeName = "batchCancelOrder"

type BatchCancelOrderItem struct {
	SymbolID    uint64  `json:"symbolID"`
	ClOrdID     string  `json:"clOrdID"`
	OrderID     *uint64 `json:"orderID,omitempty"`
	OrigClOrdID *string `json:"origClOrdID,omitempty"`
}

type BatchCancelOrderRequest struct {
	AccountID uint64                  `json:"accountID"`
	Cancels   []*BatchCancelOrderItem `json:"cancels"`
}

func (req *BatchCancelOrderRequest) ActionName() string {
	return BatchCancelOrderRequestTypeName
}
