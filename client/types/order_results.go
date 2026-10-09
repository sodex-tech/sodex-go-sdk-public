package types

type BatchNewOrderResult struct {
	Code    int    `json:"code"`
	ClOrdID string `json:"clOrdID"`

	Error *string `json:"error,omitempty"`

	OrderID *uint64 `json:"orderID,omitempty"`
}

type BatchCancelOrderResult struct {
	Code    int    `json:"code"`
	ClOrdID string `json:"clOrdID"`

	Error *string `json:"error,omitempty"`

	OrderID     *uint64 `json:"orderID,omitempty"`
	OrigClOrdID *string `json:"origClOrdID,omitempty"`
}

type NewOrderResult struct {
	Code    int    `json:"code"`
	ClOrdID string `json:"clOrdID"`

	Error *string `json:"error,omitempty"`

	OrderID *uint64 `json:"orderID,omitempty"`
}

type CancelOrderResult struct {
	Code int `json:"code"`

	Error *string `json:"error,omitempty"`

	OrderID *uint64 `json:"orderID,omitempty"`
	ClOrdID *string `json:"clOrdID,omitempty"`
}

type ModifyOrderResult struct {
	Code int `json:"code"`

	Error *string `json:"error,omitempty"`
}

type ReplaceOrderResult struct {
	Code    int    `json:"code"`
	ClOrdID string `json:"clOrdID"`

	Error *string `json:"error,omitempty"`

	OrderID *uint64 `json:"orderID,omitempty"`
}

type TransferAssetResponse struct {
	ID uint64 `json:"id"`
}
