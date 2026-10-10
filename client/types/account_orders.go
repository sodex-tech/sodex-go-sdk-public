package types

type Builder struct {
	BuilderID uint64 `json:"builderID"`
	FeeRate   uint64 `json:"feeRate"`
}

type SpotOrder struct {
	Symbol  string `json:"symbol"`
	OrderID uint64 `json:"orderID"`
	ClOrdID string `json:"clOrdID"`

	Side        string  `json:"side"`
	Type        string  `json:"type"`
	TimeInForce string  `json:"timeInForce"`
	Price       *string `json:"price,omitempty"`
	Quantity    *string `json:"origQty,omitempty"`
	Funds       *string `json:"funds,omitempty"`

	Status        string   `json:"status"`
	ExecutedQty   string   `json:"executedQty"`
	ExecutedValue string   `json:"executedValue"`
	MarginFrozen  string   `json:"marginFrozen"`
	Builder       *Builder `json:"builder,omitempty"`

	CreatedAt uint64 `json:"createdAt"`
	UpdatedAt uint64 `json:"updatedAt"`
}

type SpotAccountOpenOrders struct {
	BlockTime   uint64       `json:"blockTime"`
	BlockHeight uint64       `json:"blockHeight"`
	Orders      []*SpotOrder `json:"orders"`
}

type PerpsOrder struct {
	*SpotOrder

	PositionSide string `json:"positionSide"`
	ReduceOnly   bool   `json:"reduceOnly"`

	StopPrice   *string `json:"stopPrice,omitempty"`
	StopType    *string `json:"stopType,omitempty"`
	TriggerType *string `json:"triggerType,omitempty"`

	PositionID       *uint64  `json:"positionID,omitempty"`
	PrimaryOrderID   *uint64  `json:"primaryOrderID,omitempty"`
	AttachedOrderIDs []uint64 `json:"attachedOrderIDs,omitempty"`
}

type PerpsAccountOpenOrders struct {
	BlockTime   uint64        `json:"blockTime"`
	BlockHeight uint64        `json:"blockHeight"`
	Orders      []*PerpsOrder `json:"orders"`
}
