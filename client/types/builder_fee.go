package types

type BuilderFeeStats struct {
	Address     string `json:"address,omitempty"`
	Name        string `json:"name"`
	BuilderCode uint64 `json:"builderCode,string"`
	UserCount   uint64 `json:"userCount,string"`
	BuilderFee  string `json:"builderFee"`
	TradeCount  uint64 `json:"tradeCount,string"`
	TradeVolume string `json:"tradeVolume"`
}

type BuilderFeeDailyStats struct {
	Date        string `json:"date"`
	UserCount   uint64 `json:"userCount,string"`
	BuilderFee  string `json:"builderFee"`
	TradeCount  uint64 `json:"tradeCount,string"`
	TradeVolume string `json:"tradeVolume"`
}
