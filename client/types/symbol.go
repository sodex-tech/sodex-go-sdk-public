package types

type SymbolStatus string

const (
	SymbolStatusTrading SymbolStatus = "TRADING"
	SymbolStatusHalt    SymbolStatus = "HALT"
)

type SpotSymbol struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`

	BaseCoinID        uint64 `json:"baseCoinID"`
	BaseCoin          string `json:"baseCoin"`
	BaseCoinPrecision uint8  `json:"baseCoinPrecision"`

	QuoteCoinID        uint64 `json:"quoteCoinID"`
	QuoteCoin          string `json:"quoteCoin"`
	QuoteCoinPrecision uint8  `json:"quoteCoinPrecision"`

	PricePrecision int32  `json:"pricePrecision"`
	TickSize       string `json:"tickSize"`
	MinPrice       string `json:"minPrice"`
	MaxPrice       string `json:"maxPrice"`

	QuantityPrecision int32  `json:"quantityPrecision"`
	StepSize          string `json:"stepSize"`
	MinQty            string `json:"minQuantity"`
	MaxQty            string `json:"maxQuantity"`

	MarketMinQty string `json:"marketMinQuantity"`
	MarketMaxQty string `json:"marketMaxQuantity"`

	MinNotional string `json:"minNotional"`
	MaxNotional string `json:"maxNotional"`

	BuyLimitUpRatio      string `json:"buyLimitUpRatio"`
	SellLimitDownRatio   string `json:"sellLimitDownRatio"`
	MarketDeviationRatio string `json:"marketDeviationRatio"`

	MakerFee    string  `json:"makerFee"`
	TakerFee    string  `json:"takerFee"`
	FeeDiscount *string `json:"feeDiscount,omitempty"`

	Status SymbolStatus `json:"status"`
}

type PerpsMarginTier struct {
	MaxNotionalValue      string `json:"maxNotionalValue"`
	MaintenanceMarginRate string `json:"maintenanceMarginRate"`
	MaxLeverage           uint32 `json:"maxLeverage"`
	MaintenanceDeduction  string `json:"maintenanceDeduction"`
}

type PerpsSymbol struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`

	BaseCoin string `json:"baseCoin"`

	QuoteCoinID        uint64 `json:"quoteCoinID"`
	QuoteCoin          string `json:"quoteCoin"`
	QuoteCoinPrecision uint8  `json:"quoteCoinPrecision"`

	OpenInterestCap    string `json:"openInterestCap"`
	OpenInterestCapUSD string `json:"openInterestCapUSD"`

	PricePrecision int32  `json:"pricePrecision"`
	TickSize       string `json:"tickSize"`
	MinPrice       string `json:"minPrice"`
	MaxPrice       string `json:"maxPrice"`

	QuantityPrecision int32  `json:"quantityPrecision"`
	StepSize          string `json:"stepSize"`
	MinQty            string `json:"minQuantity"`
	MaxQty            string `json:"maxQuantity"`

	MarketMinQty string `json:"marketMinQuantity"`
	MarketMaxQty string `json:"marketMaxQuantity"`

	MinNotional string `json:"minNotional"`
	MaxNotional string `json:"maxNotional"`

	BuyLimitUpRatio      string `json:"buyLimitUpRatio"`
	SellLimitDownRatio   string `json:"sellLimitDownRatio"`
	MarketDeviationRatio string `json:"marketDeviationRatio"`

	MaxLeverage  uint32             `json:"maxLeverage"`
	InitLeverage uint32             `json:"initLeverage"`
	MarginTiers  []*PerpsMarginTier `json:"marginTiers"`

	FundingInterval uint32  `json:"fundingInterval"`
	InterestRate    string  `json:"interestRate"`
	MaxFundingRate  string  `json:"maxFundingRate"`
	MinFundingRate  string  `json:"minFundingRate"`
	FundingDiscount *string `json:"fundingDiscount,omitempty"`

	MakerFee    string  `json:"makerFee"`
	TakerFee    string  `json:"takerFee"`
	FeeDiscount *string `json:"feeDiscount,omitempty"`

	Status SymbolStatus `json:"status"`
}
