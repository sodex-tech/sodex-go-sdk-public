package types

import (
	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk/common/enums"
)

const TransferAssetRequestTypeName = "transferAsset"

type TransferAssetRequest struct {
	ID            uint64                  `json:"id"`
	FromAccountID uint64                  `json:"fromAccountID"`
	ToAccountID   uint64                  `json:"toAccountID"`
	CoinID        uint64                  `json:"coinID"`
	Amount        decimal.Decimal         `json:"amount"`
	Type          enums.TransferAssetType `json:"type"`
}

func (req *TransferAssetRequest) ActionName() string {
	return TransferAssetRequestTypeName
}
