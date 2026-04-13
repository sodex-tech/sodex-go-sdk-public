// Package ws provides a WebSocket client for real-time Sodex market data and account updates.
package ws

import "encoding/json"

// ── Request types ────────────────────────────────────────────────────────────

// Request is the client-to-server WebSocket message.
type Request struct {
	Op     string          `json:"op"`
	ID     int64           `json:"id,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
}

// SubscribeParams holds the parameters for a subscribe/unsubscribe request.
type SubscribeParams struct {
	Channel  string   `json:"channel"`
	Symbol   string   `json:"symbol,omitempty"`
	Symbols  []string `json:"symbols,omitempty"`
	User     string   `json:"user,omitempty"`
	TickSize string   `json:"tickSize,omitempty"`
	Level    int      `json:"level,omitempty"`
	Interval string   `json:"interval,omitempty"`
}

// ── Response types ───────────────────────────────────────────────────────────

// Response is the server acknowledgment for subscribe/unsubscribe.
type Response struct {
	Op      string          `json:"op"`
	ID      int64           `json:"id,omitempty"`
	Success *bool           `json:"success,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
	ConnID  string          `json:"connID,omitempty"`
	Code    string          `json:"code,omitempty"`
}

// Push is a server push message for subscribed channels.
type Push struct {
	Channel string          `json:"channel"`
	Type    string          `json:"type"` // "snapshot" or "update"
	Data    json.RawMessage `json:"data"`
}

// ── Channel data types ───────────────────────────────────────────────────────

// Ticker is a full 24h ticker snapshot/update.
type Ticker struct {
	EventTime          int64   `json:"E"`
	Symbol             string  `json:"s"`
	LastPrice          string  `json:"c"`
	LastQty            string  `json:"Q"`
	WeightedAvgPrice   string  `json:"w"`
	AskPrice           string  `json:"a"`
	AskQty             string  `json:"A"`
	BidPrice           string  `json:"b"`
	BidQty             string  `json:"B"`
	PriceChange        string  `json:"p"`
	PriceChangePercent float64 `json:"P"`
	OpenPrice          string  `json:"o"`
	HighPrice          string  `json:"h"`
	LowPrice           string  `json:"l"`
	Volume             string  `json:"v"`
	QuoteVolume        string  `json:"q"`
	OpenTime           int64   `json:"O"`
	CloseTime          int64   `json:"C"`
}

// MiniTicker is a simplified ticker.
type MiniTicker struct {
	EventTime   int64  `json:"E"`
	Symbol      string `json:"s"`
	LastPrice   string `json:"c"`
	OpenPrice   string `json:"o"`
	HighPrice   string `json:"h"`
	LowPrice    string `json:"l"`
	Volume      string `json:"v"`
	QuoteVolume string `json:"q"`
}

// BookTicker is a best bid/ask update.
type BookTicker struct {
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	UpdateID  int64  `json:"u"`
	AskPrice  string `json:"a"`
	AskQty    string `json:"A"`
	BidPrice  string `json:"b"`
	BidQty    string `json:"B"`
}

// Trade is a market trade event.
type Trade struct {
	EventTime int64  `json:"E"`
	TradeTime int64  `json:"T"`
	TradeID   int64  `json:"t"`
	Symbol    string `json:"s"`
	Side      string `json:"S"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	BuyerID   int64  `json:"bi"`
	SellerID  int64  `json:"si"`
}

// L2Book is an L2/L4 order book snapshot.
type L2Book struct {
	EventTime int64      `json:"E"`
	Symbol    string     `json:"s"`
	UpdateID  int64      `json:"u"`
	Asks      [][]string `json:"a"` // [[price, qty], ...]
	Bids      [][]string `json:"b"` // [[price, qty], ...]
}

// Candle is an OHLCV candlestick update.
type Candle struct {
	OpenTime    int64  `json:"t"`
	CloseTime   int64  `json:"T"`
	Symbol      string `json:"s"`
	Interval    string `json:"i"`
	Open        string `json:"o"`
	High        string `json:"h"`
	Low         string `json:"l"`
	Close       string `json:"c"`
	Volume      string `json:"v"`
	QuoteVolume string `json:"q"`
	NumTrades   int64  `json:"n"`
	Closed      bool   `json:"x"`
}

// MarkPrice is a perps mark price update.
type MarkPrice struct {
	EventTime       int64  `json:"E"`
	Symbol          string `json:"s"`
	OpenInterest    string `json:"oi"`
	MarkPx          string `json:"p"`
	IndexPx         string `json:"i"`
	FundingRate     string `json:"r"`
	NextFundingTime int64  `json:"T"`
}

// ── Account channel types ────────────────────────────────────────────────────

// AccountOrderUpdate is an order status change event.
type AccountOrderUpdate struct {
	EventTime   int64  `json:"E"`
	TradeTime   int64  `json:"T"`
	Symbol      string `json:"s"`
	ClOrdID     string `json:"c"`
	OrderID     int64  `json:"i"`
	Side        string `json:"S"`
	OrderType   string `json:"o"`
	Price       string `json:"p"`
	OrigQty     string `json:"q"`
	Status      string `json:"X"`
	FilledQty   string `json:"z"`
	FilledValue string `json:"v"`
	TradeID     int64  `json:"t"`
	LastQty     string `json:"l"`
	LastPrice   string `json:"L"`
	Fee         string `json:"n"`
	IsMaker     bool   `json:"m"`
	ExecType    string `json:"x"`
	Reason      string `json:"r"`
}

// AccountTrade is a user trade/fill event.
type AccountTrade struct {
	EventTime int64  `json:"E"`
	TradeTime int64  `json:"T"`
	TradeID   int64  `json:"t"`
	Symbol    string `json:"s"`
	OrderID   int64  `json:"i"`
	ClOrdID   string `json:"c"`
	Side      string `json:"S"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	Fee       string `json:"f"`
	IsMaker   bool   `json:"m"`
	Direction string `json:"d,omitempty"` // perps only: LONG/SHORT
}

// Channel name constants.
const (
	ChannelTicker          = "ticker"
	ChannelAllTicker       = "allTicker"
	ChannelMiniTicker      = "miniTicker"
	ChannelAllMiniTicker   = "allMiniTicker"
	ChannelBookTicker      = "bookTicker"
	ChannelAllBookTicker   = "allBookTicker"
	ChannelTrade           = "trade"
	ChannelL2Book          = "l2Book"
	ChannelL4Book          = "l4Book"
	ChannelCandle          = "candle"
	ChannelMarkPrice       = "markPrice"
	ChannelAllMarkPrice    = "allMarkPrice"
	ChannelAccountState    = "accountState"
	ChannelAccountUpdate   = "accountUpdate"
	ChannelAccountOrderUpd = "accountOrderUpdate"
	ChannelAccountTrade    = "accountTrade"
	ChannelAccountEvent    = "accountEvent"
)
