package types

import "github.com/shopspring/decimal"

const UpdateCollateralRequestTypeName = "updateCollateral"

// UpdateCollateralRequest adjusts cross-margin collateral. A positive amount adds collateral; a negative amount removes it.
// Amount must be non-zero and respect the asset's precision.
type UpdateCollateralRequest struct {
	AccountID uint64          `json:"accountID"`
	CoinID    uint64          `json:"coinID"`
	Amount    decimal.Decimal `json:"amount"`
}

func (req *UpdateCollateralRequest) ActionName() string {
	return UpdateCollateralRequestTypeName
}
