package types

import ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"

// DcaOrder is a compressed plan in the accountState/accountUpdate DO array.
type DcaOrder struct {
	Symbol        string             `json:"s"`
	OrderID       uint64             `json:"i"`
	Amount        string             `json:"a"`
	Side          string             `json:"S"`
	Schedule      ctypes.DcaSchedule `json:"sc"`
	MinPrice      *string            `json:"lp,omitempty"`
	MaxPrice      *string            `json:"hp,omitempty"`
	Randomize     bool               `json:"r"`
	Immediate     bool               `json:"im"`
	ExecutedQty   string             `json:"z"`
	ExecutedValue string             `json:"v"`
	Failures      uint32             `json:"f"`
	CreatedAt     uint64             `json:"ct"`
	NextActiveAt  uint64             `json:"nt"`
	Status        string             `json:"st"`
}
