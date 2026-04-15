// Package client provides an HTTP client for the Sodex REST API.
// It wraps the EIP-712 signing logic from the spot and perps signer packages
// to produce authenticated requests automatically.
package client

import (
	"encoding/json"
	"fmt"
)

// APIResponse is the standard JSON envelope returned by all Sodex REST endpoints.
// code == 0 means success; any non-zero value is an application-level error.
type APIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// Symbol describes a tradeable market (shared by spot and perps).
type Symbol struct {
	SymbolID          uint64 `json:"id"`
	Symbol            string `json:"name"`
	DisplayName       string `json:"displayName"`
	BaseAsset         string `json:"baseCoin"`
	QuoteAsset        string `json:"quoteCoin"`
	Status            string `json:"status"`
	PricePrecision    int    `json:"pricePrecision"`
	QuantityPrecision int    `json:"quantityPrecision"`
	MinQuantity       string `json:"minQuantity"`
	MaxQuantity       string `json:"maxQuantity"`
	MinPrice          string `json:"minPrice"`
	MaxPrice          string `json:"maxPrice"`
	TickSize          string `json:"tickSize"`
	StepSize          string `json:"stepSize"`
	MinNotional       string `json:"minNotional"`
	MakerFee          string `json:"makerFee,omitempty"`
	TakerFee          string `json:"takerFee,omitempty"`
	// Perps-only:
	MaxLeverage  *int    `json:"maxLeverage,omitempty"`
	ContractSize *string `json:"contractSize,omitempty"`
}

// Ticker holds 24-hour rolling statistics for a symbol.
type Ticker struct {
	Symbol             string  `json:"symbol"`
	LastPrice          string  `json:"lastPx"`
	OpenPrice          string  `json:"openPx"`
	HighPrice          string  `json:"highPx"`
	LowPrice           string  `json:"lowPx"`
	BidPrice           string  `json:"bidPx"`
	BidSize            string  `json:"bidSz"`
	AskPrice           string  `json:"askPx"`
	AskSize            string  `json:"askSz"`
	Volume             string  `json:"volume"`
	QuoteVolume        string  `json:"quoteVolume"`
	PriceChange        string  `json:"change"`
	PriceChangePercent float64 `json:"changePct"`
	// Perps-only:
	MarkPrice    *string `json:"markPrice,omitempty"`
	IndexPrice   *string `json:"indexPrice,omitempty"`
	FundingRate  *string `json:"fundingRate,omitempty"`
	OpenInterest *string `json:"openInterest,omitempty"`
}

// OrderBookLevel is a single price level in an order book snapshot.
type OrderBookLevel struct {
	Price    string
	Quantity string
}

// UnmarshalJSON parses an order book level from the API's [price, qty] array format.
func (l *OrderBookLevel) UnmarshalJSON(data []byte) error {
	var arr [2]string
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("orderbook level: expected [price, qty] array: %w", err)
	}
	l.Price = arr[0]
	l.Quantity = arr[1]
	return nil
}

// MarshalJSON encodes an order book level back to a JSON object for display.
func (l OrderBookLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Price    string `json:"price"`
		Quantity string `json:"quantity"`
	}{l.Price, l.Quantity})
}

// OrderBook is a full depth snapshot for a symbol.
type OrderBook struct {
	Symbol   string           `json:"-"` // set by caller, not in API response
	Bids     []OrderBookLevel `json:"bids"`
	Asks     []OrderBookLevel `json:"asks"`
	UpdateID uint64           `json:"updateID"`
}

// blockTimeWrapper is used to unwrap nested API responses that include
// blockTime/blockHeight metadata alongside the actual data.
type blockTimeWrapper struct {
	BlockTime   int64           `json:"blockTime"`
	BlockHeight int64           `json:"blockHeight"`
	Balances    json.RawMessage `json:"balances,omitempty"`
	Orders      json.RawMessage `json:"orders,omitempty"`
}

// AccountInfo holds the account ID and user ID returned by the /state endpoint.
type AccountInfo struct {
	Address   string `json:"user"`
	AccountID uint64 `json:"aid"`
	UserID    uint64 `json:"uid"`
}

// Balance represents a single asset balance in an account.
type Balance struct {
	CoinID uint64 `json:"id"`
	Coin   string `json:"coin"`
	Total  string `json:"total"`
	Locked string `json:"locked"`
}

// Order represents a resting or historical order record.
type Order struct {
	OrderID       uint64 `json:"orderID"`
	ClOrdID       string `json:"clOrdID"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Type          string `json:"type"`
	TimeInForce   string `json:"timeInForce"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	ExecutedValue string `json:"executedValue"`
	Status        string `json:"status"`
	MarginFrozen  string `json:"marginFrozen,omitempty"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

// Position represents an open perpetuals position.
type Position struct {
	Symbol        string `json:"symbol"`
	SymbolID      uint64 `json:"symbolID"`
	AccountID     uint64 `json:"accountID"`
	PositionSide  string `json:"positionSide"`
	Quantity      string `json:"quantity"`
	EntryPrice    string `json:"entryPrice"`
	MarkPrice     string `json:"markPrice"`
	LiqPrice      string `json:"liquidationPrice"`
	UnrealizedPnl string `json:"unrealizedPnl"`
	Leverage      int    `json:"leverage"`
	MarginMode    string `json:"marginMode"`
	Margin        string `json:"margin"`
}

// PlaceOrderResult is a single entry in the response from order-placement endpoints.
type PlaceOrderResult struct {
	OrderID uint64 `json:"orderID"`
	ClOrdID string `json:"clOrdID"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// CancelOrderResult is a single entry in the response from cancel endpoints.
type CancelOrderResult struct {
	OrderID *uint64 `json:"orderID,omitempty"`
	ClOrdID string  `json:"clOrdID"`
	Status  string  `json:"status"`
	Message string  `json:"message,omitempty"`
}

// LeverageResult is the response from the update-leverage endpoint.
type LeverageResult struct {
	Symbol     string `json:"symbol"`
	Leverage   int    `json:"leverage"`
	MarginMode string `json:"marginMode"`
}

// ModifyOrderResult is the response from the perps modify-order endpoint.
// Code == 0 indicates the modification was accepted by the engine; a non-zero
// Code means the engine rejected the modification and Error explains why.
type ModifyOrderResult struct {
	Code  int    `json:"code"`
	Error string `json:"error,omitempty"`
}

// Candle is a single OHLCV bar returned by the klines endpoint.
// Field JSON tags are single letters for wire-size reasons (matches API output).
type Candle struct {
	StartTime   uint64  `json:"t"` // Bar start time (unix milliseconds)
	Open        string  `json:"o"`
	High        string  `json:"h"`
	Low         string  `json:"l"`
	Close       string  `json:"c"`
	BaseVolume  string  `json:"v"`           // Total volume in base currency
	QuoteVolume string  `json:"q"`           // Total volume in quote currency
	Trades      *uint64 `json:"n,omitempty"` // Number of trades in the bar, when reported
}

// PublicTrade is a single recent market trade returned by the public trades endpoint.
type PublicTrade struct {
	TradeID   uint64  `json:"t"` // The trade ID
	TradeTime uint64  `json:"T"` // Trade time in unix milliseconds
	Symbol    string  `json:"s"`
	Side      string  `json:"S"` // "BUY" or "SELL" — denotes the taker side
	Price     string  `json:"p"`
	Quantity  string  `json:"q"`
	Buyer     *uint64 `json:"bi,omitempty"` // Buyer account ID (when returned)
	Seller    *uint64 `json:"si,omitempty"` // Seller account ID (when returned)
}

// UserTrade is a single filled-order record for an account (private per-user trade history).
// Distinct from PublicTrade which is market-wide.
type UserTrade struct {
	Symbol    string `json:"symbol"`
	TradeID   uint64 `json:"tradeID"`
	OrderID   uint64 `json:"orderID"`
	ClOrdID   string `json:"clOrdID"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	Fee       string `json:"fee"`
	FeeCoin   string `json:"feeCoin"`
	Timestamp uint64 `json:"time"` // Trade time in unix milliseconds
	IsMaker   bool   `json:"isMaker"`
}

// FundingPayment is a single funding payment debit/credit on a perps position.
type FundingPayment struct {
	Symbol       string `json:"symbol"`
	PositionID   uint64 `json:"positionID"`
	PositionSide string `json:"positionSide"`
	FundingFee   string `json:"fundingFee"` // Positive = user paid; negative = user received
	FeeCoin      string `json:"feeCoin"`
	Timestamp    uint64 `json:"timestamp"`
}

// HistoryFilter captures the shared set of pagination/filter query parameters for
// the history endpoints (orders/history, trades, fundings).
// All fields are optional — zero values mean "omit from the query".
type HistoryFilter struct {
	Symbol    string // filter by trading pair (engine-native format, e.g. "BTC-USD" or "vBTC_vUSDC")
	StartTime int64  // inclusive lower bound, unix milliseconds
	EndTime   int64  // inclusive upper bound, unix milliseconds
	Limit     int    // max number of rows to return (API caps at 1000)
}
