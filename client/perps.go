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
	ptypes "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
)

const perpsBase = "/api/v1/perps"

// ── Market data (unauthenticated) ─────────────────────────────────────────────

// PerpsSymbols returns all available perpetuals trading pairs.
func (c *Client) PerpsSymbols(ctx context.Context) ([]*rpctypes.PerpsSymbol, error) {
	var result []*rpctypes.PerpsSymbol
	if err := c.get(ctx, perpsBase+"/markets/symbols", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpsTickers returns 24-hour rolling stats for all perps pairs.
func (c *Client) PerpsTickers(ctx context.Context) ([]*rpctypes.PerpsTicker, error) {
	var result []*rpctypes.PerpsTicker
	if err := c.get(ctx, perpsBase+"/markets/tickers", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpsOrderBook returns the order book snapshot for symbol.
// Pass depth <= 0 to use the API default.
func (c *Client) PerpsOrderBook(ctx context.Context, symbol string, depth int) (*rpctypes.OrderBook, error) {
	u, _ := url.Parse(c.cfg.BaseURL + perpsBase + "/markets/" + symbol + "/orderbook")
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

// PerpsBalances returns asset balances for address.
func (c *Client) PerpsBalances(ctx context.Context, address string) (*rpctypes.PerpsAccountBalances, error) {
	var result rpctypes.PerpsAccountBalances
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/balances", perpsBase, address), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PerpsOrders returns all open orders for address.
func (c *Client) PerpsOrders(ctx context.Context, address string) (*rpctypes.PerpsAccountOpenOrders, error) {
	var result rpctypes.PerpsAccountOpenOrders
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/orders", perpsBase, address), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PerpsKlines returns historical OHLCV candles for a perps symbol.
//
// interval is one of: "1m","3m","5m","15m","30m","1h","2h","4h","6h","8h","12h",
// "1D","3D","1W","1M".
//
// Fields on HistoryFilter that apply: Symbol (ignored — pass via the symbol arg),
// StartTime, EndTime, Limit (default 500, max 1500).
func (c *Client) PerpsKlines(
	ctx context.Context, symbol, interval string, filter HistoryFilter,
) ([]*rpctypes.Candle, error) {
	return c.klines(ctx, perpsBase, symbol, interval, filter)
}

// PerpsPublicTrades returns recent market trades for a perps symbol.
// Only Limit on the filter applies (default 50, max 500).
func (c *Client) PerpsPublicTrades(
	ctx context.Context, symbol string, limit int,
) ([]*rpctypes.Trade, error) {
	return c.publicTrades(ctx, perpsBase, symbol, limit)
}

// PerpsOrdersHistory returns historical (non-open) orders for address on the perps engine.
// Supports filtering by symbol, time range, and limit.
func (c *Client) PerpsOrdersHistory(
	ctx context.Context, address string, filter HistoryFilter,
) ([]*rpctypes.PerpsOrder, error) {
	var result []*rpctypes.PerpsOrder
	if err := c.ordersHistory(ctx, perpsBase, address, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpsUserTrades returns the authenticated user's trade (fill) history on the perps engine.
// Supports filtering by symbol, orderID, time range, and limit.
func (c *Client) PerpsUserTrades(
	ctx context.Context, address string, filter HistoryFilter,
) ([]*rpctypes.AccountTrade, error) {
	return c.userTrades(ctx, perpsBase, address, filter)
}

// PerpsFundingHistory returns historical funding payments for the user's perps positions.
// Filter supports Symbol, StartTime, EndTime, Limit.
func (c *Client) PerpsFundingHistory(
	ctx context.Context, address string, filter HistoryFilter,
) ([]*rpctypes.PerpsFunding, error) {
	var result []*rpctypes.PerpsFunding
	path := fmt.Sprintf("%s/accounts/%s/fundings", perpsBase, address)
	if err := c.getHistory(ctx, path, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpsPositions returns all open positions for address.
func (c *Client) PerpsPositions(ctx context.Context, address string) (*rpctypes.PerpsAccountPositions, error) {
	var result rpctypes.PerpsAccountPositions
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/positions", perpsBase, address), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Authenticated trading methods ─────────────────────────────────────────────

// PlacePerpsOrder submits a perpetuals order batch. A private key must be configured.
func (c *Client) PlacePerpsOrder(ctx context.Context, req *ptypes.NewOrderRequest) ([]*rpctypes.NewOrderResult, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignNewOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign new order: %w", err)
	}
	var result []*rpctypes.NewOrderResult
	if err := c.postSigned(ctx, perpsBase+"/trade/orders", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CancelPerpsOrders cancels perpetuals orders.
func (c *Client) CancelPerpsOrders(ctx context.Context, req *ptypes.CancelOrderRequest) ([]*rpctypes.CancelOrderResult, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignCancelOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign cancel order: %w", err)
	}
	var result []*rpctypes.CancelOrderResult
	if err := c.deleteSigned(ctx, perpsBase+"/trade/orders", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ModifyPerpsOrder modifies a single resting perpetuals order's price, quantity,
// or stop price without cancelling and re-placing it. Identify the target order
// via OrderID or ClOrdID (exactly one).
func (c *Client) ModifyPerpsOrder(
	ctx context.Context, req *ptypes.ModifyOrderRequest,
) (*rpctypes.ModifyOrderResult, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignModifyOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign modify order: %w", err)
	}
	var result rpctypes.ModifyOrderResult
	if err := c.postSigned(ctx, perpsBase+"/trade/orders/modify", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReplacePerpsOrders atomically replaces a batch of resting perpetuals orders.
// Each replacement cancels the original and places a new one; if the new order
// is rejected (e.g. invalid price), the original is also cancelled.
func (c *Client) ReplacePerpsOrders(
	ctx context.Context, req *ctypes.ReplaceOrderRequest,
) ([]*rpctypes.ReplaceOrderResult, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignReplaceOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign replace order: %w", err)
	}
	var result []*rpctypes.ReplaceOrderResult
	if err := c.postSigned(ctx, perpsBase+"/trade/orders/replace", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateLeverage changes leverage for a perpetuals position.
// The endpoint returns no data on success.
func (c *Client) UpdateLeverage(ctx context.Context, req *ptypes.UpdateLeverageRequest) error {
	if c.perpsSgn == nil {
		return ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignUpdateLeverageRequest(req, nonce)
	if err != nil {
		return fmt.Errorf("perps: sign update leverage: %w", err)
	}
	return c.postSigned(ctx, perpsBase+"/trade/leverage", req, sig, nonce, nil)
}

// UpdateMargin adjusts margin for a perpetuals position.
func (c *Client) UpdateMargin(ctx context.Context, req *ptypes.UpdateMarginRequest) error {
	if c.perpsSgn == nil {
		return ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignUpdateMarginRequest(req, nonce)
	if err != nil {
		return fmt.Errorf("perps: sign update margin: %w", err)
	}
	return c.postSigned(ctx, perpsBase+"/trade/margin", req, sig, nonce, nil)
}

// PerpsTransfer transfers assets between perps accounts and returns the transfer ID.
func (c *Client) PerpsTransfer(ctx context.Context, req *ctypes.TransferAssetRequest) (*rpctypes.TransferAssetResponse, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignTransferAssetRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign transfer: %w", err)
	}
	var result rpctypes.TransferAssetResponse
	if err := c.postSigned(ctx, perpsBase+"/accounts/transfers", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SchedulePerpsCancel arms (or clears) a "dead-man's switch" that automatically
// cancels all of the user's perps orders after scheduledTimestamp (unix ms).
//
// Pass a non-nil ScheduledTimestamp on req to arm the schedule, or nil to clear
// an existing schedule. Re-sending with a future timestamp extends the deadline.
func (c *Client) SchedulePerpsCancel(ctx context.Context, req *ctypes.ScheduleCancelRequest) error {
	if c.perpsSgn == nil {
		return ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignScheduleCancelRequest(req, nonce)
	if err != nil {
		return fmt.Errorf("perps: sign schedule cancel: %w", err)
	}
	return c.postSigned(ctx, perpsBase+"/trade/orders/schedule-cancel", req, sig, nonce, nil)
}

// ── Convenience helpers ───────────────────────────────────────────────────────

// PlacePerpsLimitOrder is a one-call helper for a single perps limit order.
// symbolID is the numeric ID from PerpsSymbols().
func (c *Client) PlacePerpsLimitOrder(
	ctx context.Context,
	accountID, symbolID uint64,
	clOrdID string,
	side enums.OrderSide,
	posSide enums.PositionSide,
	tif enums.TimeInForce,
	price, qty decimal.Decimal,
	reduceOnly bool,
) ([]*rpctypes.NewOrderResult, error) {
	return c.PlacePerpsOrder(ctx, &ptypes.NewOrderRequest{
		AccountID: accountID,
		SymbolID:  symbolID,
		Orders: []*ptypes.RawOrder{{
			ClOrdID:      clOrdID,
			Modifier:     enums.OrderModifierNormal,
			Side:         side,
			Type:         enums.OrderTypeLimit,
			TimeInForce:  tif,
			Price:        &price,
			Quantity:     &qty,
			PositionSide: posSide,
			ReduceOnly:   reduceOnly,
		}},
	})
}

// PlacePerpsMarketOrder is a one-call helper for a single perps market order.
func (c *Client) PlacePerpsMarketOrder(
	ctx context.Context,
	accountID, symbolID uint64,
	clOrdID string,
	side enums.OrderSide,
	posSide enums.PositionSide,
	qty decimal.Decimal,
	reduceOnly bool,
) ([]*rpctypes.NewOrderResult, error) {
	return c.PlacePerpsOrder(ctx, &ptypes.NewOrderRequest{
		AccountID: accountID,
		SymbolID:  symbolID,
		Orders: []*ptypes.RawOrder{{
			ClOrdID:      clOrdID,
			Modifier:     enums.OrderModifierNormal,
			Side:         side,
			Type:         enums.OrderTypeMarket,
			TimeInForce:  enums.TimeInForceIOC,
			Quantity:     &qty,
			PositionSide: posSide,
			ReduceOnly:   reduceOnly,
		}},
	})
}
