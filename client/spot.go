package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/shopspring/decimal"
	rpctypes "github.com/sodex-tech/sodex-go-sdk-public/client/types"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
	stypes "github.com/sodex-tech/sodex-go-sdk-public/spot/types"
	wstypes "github.com/sodex-tech/sodex-go-sdk-public/ws/types"
)

const spotBase = "/api/v1/spot"

// ── Market data (unauthenticated) ─────────────────────────────────────────────

// SpotSymbols returns all available spot trading pairs.
func (c *Client) SpotSymbols(ctx context.Context) ([]*rpctypes.SpotSymbol, error) {
	var result []*rpctypes.SpotSymbol
	if err := c.get(ctx, spotBase+"/markets/symbols", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotTickers returns 24-hour rolling stats for all spot pairs.
func (c *Client) SpotTickers(ctx context.Context) ([]*rpctypes.SpotTicker, error) {
	var result []*rpctypes.SpotTicker
	if err := c.get(ctx, spotBase+"/markets/tickers", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotOrderBook returns the order book snapshot for symbol.
// symbol is the internal name (e.g. vBTC_vUSDC). Pass depth <= 0 to use the API default.
func (c *Client) SpotOrderBook(ctx context.Context, symbol string, depth int) (*rpctypes.OrderBook, error) {
	u, _ := url.Parse(c.cfg.BaseURL + spotBase + "/markets/" + symbol + "/orderbook")
	if depth > 0 {
		q := u.Query()
		q.Set("limit", strconv.Itoa(depth))
		u.RawQuery = q.Encode()
	}
	req, err := newGetReq(ctx, u.String())
	if err != nil {
		return nil, err
	}
	var result rpctypes.OrderBook
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotKlines returns historical OHLCV candles for a spot symbol.
//
// symbol is the internal name (e.g. "vBTC_vUSDC").
// interval is one of: "1m","3m","5m","15m","30m","1h","2h","4h","6h","8h","12h",
// "1D","3D","1W","1M".
// Only filter.StartTime / filter.EndTime / filter.Limit apply (default 500, max 1500).
func (c *Client) SpotKlines(
	ctx context.Context, symbol, interval string, filter HistoryFilter,
) ([]*rpctypes.Candle, error) {
	return c.klines(ctx, spotBase, symbol, interval, filter)
}

// SpotPublicTrades returns recent market trades for a spot symbol.
// Only limit applies (default 50, max 500).
func (c *Client) SpotPublicTrades(
	ctx context.Context, symbol string, limit int,
) ([]*rpctypes.Trade, error) {
	return c.publicTrades(ctx, spotBase, symbol, limit)
}

// SpotOrdersHistory returns historical (non-open) orders for address on the spot engine.
// Supports filtering by symbol, time range, and limit.
func (c *Client) SpotOrdersHistory(
	ctx context.Context, address string, filter HistoryFilter,
) ([]*rpctypes.SpotOrder, error) {
	var result []*rpctypes.SpotOrder
	if err := c.ordersHistory(ctx, spotBase, address, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotUserTrades returns the user's trade (fill) history on the spot engine.
// Supports filtering by symbol, orderID, time range, and limit.
func (c *Client) SpotUserTrades(
	ctx context.Context, address string, filter HistoryFilter,
) ([]*rpctypes.AccountTrade, error) {
	return c.userTrades(ctx, spotBase, address, filter)
}

// SpotAccountInfo returns the account ID and user ID for the given address.
func (c *Client) SpotAccountInfo(ctx context.Context, address string) (*wstypes.WsSpotAccountState, error) {
	var result wstypes.WsSpotAccountState
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/state", spotBase, address), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotBalances returns asset balances with snapshot block metadata.
func (c *Client) SpotBalances(ctx context.Context, address string) (*rpctypes.SpotAccountBalances, error) {
	var result rpctypes.SpotAccountBalances
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/balances", spotBase, address), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotOrders returns all open orders for address.
func (c *Client) SpotOrders(ctx context.Context, address string) (*rpctypes.SpotAccountOpenOrders, error) {
	var result rpctypes.SpotAccountOpenOrders
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/orders", spotBase, address), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Authenticated trading methods ─────────────────────────────────────────────

// PlaceSpotOrders submits a batch of spot orders. A private key must be configured.
func (c *Client) PlaceSpotOrders(ctx context.Context, req *stypes.BatchNewOrderRequest) ([]*rpctypes.BatchNewOrderResult, error) {
	if c.spotSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.spotSgn.SignBatchNewOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("spot: sign new order: %w", err)
	}
	var result []*rpctypes.BatchNewOrderResult
	if err := c.postSigned(ctx, spotBase+"/trade/orders/batch", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CancelSpotOrders submits a batch of spot order cancellations.
func (c *Client) CancelSpotOrders(ctx context.Context, req *stypes.BatchCancelOrderRequest) ([]*rpctypes.BatchCancelOrderResult, error) {
	if c.spotSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.spotSgn.SignBatchCancelOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("spot: sign cancel order: %w", err)
	}
	var result []*rpctypes.BatchCancelOrderResult
	if err := c.deleteSigned(ctx, spotBase+"/trade/orders/batch", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ReplaceSpotOrders replaces a batch of existing spot orders.
func (c *Client) ReplaceSpotOrders(ctx context.Context, req *ctypes.ReplaceOrderRequest) ([]*rpctypes.ReplaceOrderResult, error) {
	if c.spotSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.spotSgn.SignReplaceOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("spot: sign replace order: %w", err)
	}
	var result []*rpctypes.ReplaceOrderResult
	if err := c.postSigned(ctx, spotBase+"/trade/orders/replace", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotTransfer transfers assets between spot accounts and returns the transfer ID.
func (c *Client) SpotTransfer(ctx context.Context, req *ctypes.TransferAssetRequest) (*rpctypes.TransferAssetResponse, error) {
	if c.spotSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.spotSgn.SignTransferAssetRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("spot: sign transfer: %w", err)
	}
	var result rpctypes.TransferAssetResponse
	if err := c.postSigned(ctx, spotBase+"/accounts/transfers", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ScheduleSpotCancel arms (or clears) a "dead-man's switch" that automatically
// cancels all of the user's spot orders after scheduledTimestamp (unix ms).
//
// Pass a non-nil ScheduledTimestamp on req to arm the schedule, or nil to clear
// an existing schedule. Re-sending with a future timestamp extends the deadline.
func (c *Client) ScheduleSpotCancel(ctx context.Context, req *ctypes.ScheduleCancelRequest) error {
	if c.spotSgn == nil {
		return ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.spotSgn.SignScheduleCancelRequest(req, nonce)
	if err != nil {
		return fmt.Errorf("spot: sign schedule cancel: %w", err)
	}
	return c.postSigned(ctx, spotBase+"/trade/orders/schedule-cancel", req, sig, nonce, nil)
}

// ── Convenience helpers ───────────────────────────────────────────────────────

// PlaceSpotLimitOrder is a one-call helper for a single spot limit order.
// symbolID is the numeric ID from SpotSymbols().
func (c *Client) PlaceSpotLimitOrder(
	ctx context.Context,
	accountID, symbolID uint64,
	clOrdID string,
	side enums.OrderSide,
	tif enums.TimeInForce,
	price, qty decimal.Decimal,
) ([]*rpctypes.BatchNewOrderResult, error) {
	return c.PlaceSpotOrders(ctx, &stypes.BatchNewOrderRequest{
		AccountID: accountID,
		Orders: []*stypes.BatchNewOrderItem{{
			SymbolID:    symbolID,
			ClOrdID:     clOrdID,
			Side:        side,
			Type:        enums.OrderTypeLimit,
			TimeInForce: tif,
			Price:       &price,
			Quantity:    &qty,
		}},
	})
}

// PlaceSpotMarketOrder is a one-call helper for a single spot market order.
func (c *Client) PlaceSpotMarketOrder(
	ctx context.Context,
	accountID, symbolID uint64,
	clOrdID string,
	side enums.OrderSide,
	qty decimal.Decimal,
) ([]*rpctypes.BatchNewOrderResult, error) {
	return c.PlaceSpotOrders(ctx, &stypes.BatchNewOrderRequest{
		AccountID: accountID,
		Orders: []*stypes.BatchNewOrderItem{{
			SymbolID:    symbolID,
			ClOrdID:     clOrdID,
			Side:        side,
			Type:        enums.OrderTypeMarket,
			TimeInForce: enums.TimeInForceIOC,
			Quantity:    &qty,
		}},
	})
}
