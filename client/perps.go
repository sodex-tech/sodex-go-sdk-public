package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/shopspring/decimal"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
	ptypes "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
)

const perpsBase = "/api/v1/perps"

// ── Market data (unauthenticated) ─────────────────────────────────────────────

// PerpsSymbols returns all available perpetuals trading pairs.
func (c *Client) PerpsSymbols(ctx context.Context) ([]Symbol, error) {
	var result []Symbol
	if err := c.get(ctx, perpsBase+"/markets/symbols", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpsTickers returns 24-hour rolling stats for all perps pairs.
func (c *Client) PerpsTickers(ctx context.Context) ([]Ticker, error) {
	var result []Ticker
	if err := c.get(ctx, perpsBase+"/markets/tickers", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpsOrderBook returns the order book snapshot for symbol.
// Pass depth <= 0 to use the API default.
func (c *Client) PerpsOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error) {
	u, _ := url.Parse(c.cfg.BaseURL + perpsBase + "/markets/" + symbol + "/orderbook")
	if depth > 0 {
		q := u.Query()
		q.Set("depth", strconv.Itoa(depth))
		u.RawQuery = q.Encode()
	}
	req, err := newGetReq(ctx, u.String())
	if err != nil {
		return nil, err
	}
	var result OrderBook
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	result.Symbol = symbol
	return &result, nil
}

// PerpsBalances returns asset balances for address.
func (c *Client) PerpsBalances(ctx context.Context, address string) ([]Balance, error) {
	var wrapper blockTimeWrapper
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/balances", perpsBase, address), &wrapper); err != nil {
		return nil, err
	}
	var result []Balance
	if len(wrapper.Balances) > 0 {
		if err := json.Unmarshal(wrapper.Balances, &result); err != nil {
			return nil, fmt.Errorf("perps: parse balances: %w", err)
		}
	}
	return result, nil
}

// PerpsOrders returns all open orders for address.
func (c *Client) PerpsOrders(ctx context.Context, address string) ([]Order, error) {
	var wrapper blockTimeWrapper
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/orders", perpsBase, address), &wrapper); err != nil {
		return nil, err
	}
	var result []Order
	if len(wrapper.Orders) > 0 {
		if err := json.Unmarshal(wrapper.Orders, &result); err != nil {
			return nil, fmt.Errorf("perps: parse orders: %w", err)
		}
	}
	return result, nil
}

// PerpsPositions returns all open positions for address.
func (c *Client) PerpsPositions(ctx context.Context, address string) ([]Position, error) {
	// Positions endpoint returns the same wrapper as orders.
	var wrapper blockTimeWrapper
	if err := c.get(ctx, fmt.Sprintf("%s/accounts/%s/positions", perpsBase, address), &wrapper); err != nil {
		return nil, err
	}
	var result []Position
	if len(wrapper.Orders) > 0 {
		if err := json.Unmarshal(wrapper.Orders, &result); err != nil {
			return nil, fmt.Errorf("perps: parse positions: %w", err)
		}
	}
	return result, nil
}

// ── Authenticated trading methods ─────────────────────────────────────────────

// PlacePerpsOrder submits a perpetuals order batch. A private key must be configured.
func (c *Client) PlacePerpsOrder(ctx context.Context, req *ptypes.NewOrderRequest) ([]PlaceOrderResult, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignNewOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign new order: %w", err)
	}
	var result []PlaceOrderResult
	if err := c.postSigned(ctx, perpsBase+"/trade/orders", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CancelPerpsOrders cancels perpetuals orders.
func (c *Client) CancelPerpsOrders(ctx context.Context, req *ptypes.CancelOrderRequest) ([]CancelOrderResult, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignCancelOrderRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign cancel order: %w", err)
	}
	var result []CancelOrderResult
	if err := c.deleteSigned(ctx, perpsBase+"/trade/orders", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateLeverage changes leverage for a perpetuals position.
func (c *Client) UpdateLeverage(ctx context.Context, req *ptypes.UpdateLeverageRequest) (*LeverageResult, error) {
	if c.perpsSgn == nil {
		return nil, ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignUpdateLeverageRequest(req, nonce)
	if err != nil {
		return nil, fmt.Errorf("perps: sign update leverage: %w", err)
	}
	var result LeverageResult
	if err := c.postSigned(ctx, perpsBase+"/trade/leverage", req, sig, nonce, &result); err != nil {
		return nil, err
	}
	return &result, nil
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

// PerpsTransfer transfers assets between perps accounts.
func (c *Client) PerpsTransfer(ctx context.Context, req *ctypes.TransferAssetRequest) error {
	if c.perpsSgn == nil {
		return ErrNotAuthenticated
	}
	nonce := c.nonce()
	sig, err := c.perpsSgn.SignTransferAssetRequest(req, nonce)
	if err != nil {
		return fmt.Errorf("perps: sign transfer: %w", err)
	}
	return c.postSigned(ctx, perpsBase+"/accounts/transfers", req, sig, nonce, nil)
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
) ([]PlaceOrderResult, error) {
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
) ([]PlaceOrderResult, error) {
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
