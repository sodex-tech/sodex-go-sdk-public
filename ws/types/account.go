package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type WsSpotAccountState struct {
	User      common.Address `json:"user"`
	AccountID uint64         `json:"aid"`
	UserID    uint64         `json:"uid"`

	Balances []*SpotBalance `json:"B"`
	Orders   []*SpotOrder   `json:"O"`
	Twaps    []*TwapOrder   `json:"TO"`
}

type WsSpotAccountUpdate struct {
	EventTime   uint64 `json:"E"`
	BlockTime   uint64 `json:"T"`
	BlockHeight uint64 `json:"h"`

	Balances []*SpotBalance `json:"B"`
	Twaps    []*TwapOrder   `json:"TO"`
}

type WsPerpsAccountState struct {
	User      common.Address `json:"user"`
	AccountID uint64         `json:"aid"`
	UserID    uint64         `json:"uid"`

	AccountValue      decimal.Decimal `json:"av"`
	MaintenanceMargin decimal.Decimal `json:"mm"`
	MMR               decimal.Decimal `json:"mmr"`

	AvailableMargin           decimal.Decimal `json:"am"`
	AvailableMarginIsolated   decimal.Decimal `json:"ami"`
	AvailableMarginWithdrawal decimal.Decimal `json:"amw"`

	IsolatedMargin decimal.Decimal `json:"im"`
	CrossMargin    decimal.Decimal `json:"cm"`

	OpenOrderIsolatedMargin decimal.Decimal `json:"oim"`
	OpenOrderCrossMargin    decimal.Decimal `json:"ocm"`

	Balances      []*PerpsBalanceDetailed `json:"B"`
	Positions     []*PerpsPosition        `json:"P"`
	Orders        []*PerpsOrder           `json:"O"`
	Twaps         []*TwapOrder            `json:"TO"`
	SymbolConfigs []*PerpsSymbolConfig    `json:"S"`
}

type WsPerpsAccountUpdate struct {
	EventTime   uint64 `json:"E"`
	BlockTime   uint64 `json:"T"`
	BlockHeight uint64 `json:"h"`

	Balances  []*PerpsBalanceLite  `json:"B"`
	Positions []*PerpsPositionLite `json:"P"`
	Twaps     []*TwapOrder         `json:"TO"`
}
