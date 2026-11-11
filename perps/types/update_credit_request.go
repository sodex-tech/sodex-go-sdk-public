package types

import "github.com/shopspring/decimal"

const UpdateCreditRequestTypeName = "updateCredit"

// UpdateCreditRequest adjusts a credit account. A positive amount mints credit; a negative amount burns it.
// Amount must be non-zero and respect the asset's precision.
type UpdateCreditRequest struct {
	AccountID uint64          `json:"accountID"`
	Amount    decimal.Decimal `json:"amount"`
}

func (req *UpdateCreditRequest) ActionName() string {
	return UpdateCreditRequestTypeName
}
