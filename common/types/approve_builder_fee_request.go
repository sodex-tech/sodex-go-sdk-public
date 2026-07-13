package types

const ApproveBuilderFeeRequestTypeName = "approveBuilderFee"

type ApproveBuilderFeeRequest struct {
	AccountID  uint64 `json:"accountID"`
	BuilderID  uint64 `json:"builderID"`
	MaxFeeRate uint64 `json:"maxFeeRate"`
}

func (req *ApproveBuilderFeeRequest) ActionName() string {
	return ApproveBuilderFeeRequestTypeName
}
