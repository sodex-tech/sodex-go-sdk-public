package types

import "github.com/shopspring/decimal"

type SpotOrder struct {
	Symbol        string   `json:"s"`
	ClOrdID       string   `json:"c"`
	OrderID       uint64   `json:"i"`
	Side          string   `json:"S"`
	Type          string   `json:"o"`
	TimeInForce   string   `json:"f"`
	Price         string   `json:"p"`
	Quantity      string   `json:"q"`
	Funds         *string  `json:"F,omitempty"`
	Status        string   `json:"X"`
	ExecutedQty   string   `json:"z"`
	ExecutedValue string   `json:"v"`
	MarginFrozen  string   `json:"M"`
	Builder       *Builder `json:"b,omitempty"`

	CreatedAt *uint64 `json:"ct,omitempty"`
	UpdatedAt *uint64 `json:"ut,omitempty"`
}

type SpotOrderUpdate struct {
	*SpotOrder

	UpdatedAt uint64  `json:"T"`
	TradeID   *uint64 `json:"t,omitempty"`
	LastQty   *string `json:"l,omitempty"`
	LastPx    *string `json:"L,omitempty"`
	Fee       *string `json:"n,omitempty"`
	IsMaker   *bool   `json:"m,omitempty"`
	ExecType  *string `json:"x,omitempty"`
	Reason    *string `json:"r,omitempty"`
}

type WsSpotOrderUpdate struct {
	EventTime uint64 `json:"E"`

	*SpotOrderUpdate
}

type PerpsOrder struct {
	*SpotOrder

	PositionSide string           `json:"ps"`
	ReduceOnly   bool             `json:"R"`
	StopPrice    *decimal.Decimal `json:"sp,omitempty"`
	StopType     *string          `json:"st,omitempty"`
	TriggerType  *string          `json:"tt,omitempty"`

	PositionID       *uint64  `json:"pid,omitempty"`
	PrimaryOrderID   *uint64  `json:"poid,omitempty"`
	AttachedOrderIDs []uint64 `json:"aoids,omitempty"`
}

type PerpsOrderUpdate struct {
	*PerpsOrder

	UpdatedAt uint64  `json:"T"`
	TradeID   *uint64 `json:"t,omitempty"`
	LastQty   *string `json:"l,omitempty"`
	LastPx    *string `json:"L,omitempty"`
	Fee       *string `json:"n,omitempty"`
	IsMaker   *bool   `json:"m,omitempty"`
	ExecType  *string `json:"x,omitempty"`
	Reason    *string `json:"r,omitempty"`
}

type WsPerpsOrderUpdate struct {
	EventTime uint64 `json:"E"`

	*PerpsOrderUpdate
}
