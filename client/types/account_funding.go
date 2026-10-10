package types

type PerpsFunding struct {
	Symbol       string `json:"symbol"`
	PositionID   uint64 `json:"positionID"`
	PositionSide string `json:"positionSide"`
	FundingFee   string `json:"fundingFee"`
	FeeCoin      string `json:"feeCoin"`
	Timestamp    uint64 `json:"timestamp"`
}
