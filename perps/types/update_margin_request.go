package types

import (
	"github.com/shopspring/decimal"
)

const UpdateMarginRequestTypeName = "updateMargin"

type UpdateMarginRequest struct {
	AccountID uint64          `json:"accountID"`
	SymbolID  uint64          `json:"symbolID"`
	Amount    decimal.Decimal `json:"amount"`
}

func (req *UpdateMarginRequest) ActionName() string {
	return UpdateMarginRequestTypeName
}
