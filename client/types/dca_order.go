package types

import ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"

type DcaOrder struct {
	Symbol        string             `json:"symbol"`
	OrderID       uint64             `json:"orderID"`
	Amount        string             `json:"amount"`
	Side          string             `json:"side"`
	Schedule      ctypes.DcaSchedule `json:"schedule"`
	MinPrice      *string            `json:"minPrice,omitempty"`
	MaxPrice      *string            `json:"maxPrice,omitempty"`
	Randomize     bool               `json:"randomize"`
	Immediate     bool               `json:"immediate"`
	ExecutedQty   string             `json:"executedQty"`
	ExecutedValue string             `json:"executedValue"`
	Failures      uint32             `json:"failures"`
	CreatedAt     uint64             `json:"createdAt"`
	NextActiveAt  uint64             `json:"nextActiveAt"`
	Status        string             `json:"status"`
}

type AccountDcaOrders struct {
	BlockTime   uint64      `json:"blockTime"`
	BlockHeight uint64      `json:"blockHeight"`
	Dcas        []*DcaOrder `json:"dcas"`
}

type DcaOrderResult struct {
	Code    int     `json:"code"`
	Error   *string `json:"error,omitempty"`
	OrderID *uint64 `json:"orderID,omitempty"`
}
