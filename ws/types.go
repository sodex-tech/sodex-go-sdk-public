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
	Coins    []string `json:"coins,omitempty"`
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
	ChannelCoinPrice       = "coinPrice"
	ChannelAllCoinPrice    = "allCoinPrice"
	ChannelAccountState    = "accountState"
	ChannelAccountUpdate   = "accountUpdate"
	ChannelAccountOrderUpd = "accountOrderUpdate"
	ChannelAccountTrade    = "accountTrade"
	ChannelAccountEvent    = "accountEvent"
)
