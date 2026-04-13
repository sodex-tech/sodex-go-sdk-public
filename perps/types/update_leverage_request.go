package types

import "github.com/sodex-tech/sodex-go-sdk-public/common/enums"

const UpdateLeverageRequestTypeName = "updateLeverage"

type UpdateLeverageRequest struct {
	AccountID  uint64           `json:"accountID"`
	SymbolID   uint64           `json:"symbolID"`
	Leverage   uint32           `json:"leverage"`
	MarginMode enums.MarginMode `json:"marginMode"`
}

func (req *UpdateLeverageRequest) ActionName() string {
	return UpdateLeverageRequestTypeName
}
