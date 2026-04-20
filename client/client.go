package client

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	perpssgn "github.com/sodex-tech/sodex-go-sdk-public/perps/signer"
	spotsgn "github.com/sodex-tech/sodex-go-sdk-public/spot/signer"
)

const (
	// DefaultBaseURL is the default Sodex mainnet API base URL.
	DefaultBaseURL = "https://mainnet-gw.sodex.dev"
	// TestnetBaseURL is the Sodex testnet API base URL.
	TestnetBaseURL = "https://testnet-gw.sodex.dev"
	// DefaultChainID is the Sodex mainnet chain ID.
	DefaultChainID = uint64(286623)
	// TestnetChainID is the Sodex testnet chain ID.
	TestnetChainID = uint64(138565)
)

// Config holds all configuration for a Client instance.
type Config struct {
	// BaseURL is the API root (without trailing slash).
	// Defaults to DefaultBaseURL if empty.
	BaseURL string

	// ChainID is the EVM chain ID used for EIP-712 domain separation.
	// Defaults to DefaultChainID if zero.
	ChainID uint64

	// PrivateKey enables authenticated (trading) methods.
	// Leave nil for read-only (market-data) access.
	PrivateKey *ecdsa.PrivateKey

	// APIKeyName is the name of the API key to use for authenticated requests.
	// When set, the X-API-Key header is included in signed requests.
	// Leave empty to authenticate directly with the master private key.
	APIKeyName string

	// HTTPClient is an optional custom HTTP client.
	// A 30-second-timeout client is used if nil.
	HTTPClient *http.Client
}

// Client is an HTTP client for the Sodex REST API.
// It is safe to use concurrently.
type Client struct {
	cfg      Config
	http     *http.Client
	spotSgn  *spotsgn.Signer
	perpsSgn *perpssgn.Signer
	// lastNonce ensures strict nonce monotonicity even under concurrent calls.
	lastNonce uint64
}

// New creates a Client from cfg.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.ChainID == 0 {
		cfg.ChainID = DefaultChainID
	}
	h := cfg.HTTPClient
	if h == nil {
		h = &http.Client{Timeout: 30 * time.Second}
	}
	c := &Client{cfg: cfg, http: h}
	if cfg.PrivateKey != nil {
		c.spotSgn = spotsgn.NewSigner(cfg.ChainID, cfg.PrivateKey)
		c.perpsSgn = perpssgn.NewSigner(cfg.ChainID, cfg.PrivateKey)
	}
	return c
}

// Address returns the Ethereum address derived from the configured private key,
// or an empty string when no key is configured.
func (c *Client) Address() string {
	if c.cfg.PrivateKey == nil {
		return ""
	}
	return crypto.PubkeyToAddress(c.cfg.PrivateKey.PublicKey).Hex()
}

// nonce returns a strictly-monotonic uint64 nonce close to the current Unix
// millisecond timestamp. The Sodex API expects the nonce to be a millisecond
// timestamp and accepts values within (now-2days, now+1day).
func (c *Client) nonce() uint64 {
	ts := uint64(time.Now().UnixMilli())
	for {
		last := atomic.LoadUint64(&c.lastNonce)
		next := ts
		if next <= last {
			next = last + 1
		}
		if atomic.CompareAndSwapUint64(&c.lastNonce, last, next) {
			return next
		}
	}
}

// ErrNotAuthenticated is returned when a trading method is called on a
// Client that was created without a private key.
var ErrNotAuthenticated = fmt.Errorf("client: not authenticated — set Config.PrivateKey or SODEX_PRIVATE_KEY")

// ErrAPI wraps an application-level error returned by the Sodex API.
type ErrAPI struct {
	Code    int
	Message string
}

func (e *ErrAPI) Error() string {
	return fmt.Sprintf("sodex API error (code %d): %s", e.Code, e.Message)
}

// ── internal HTTP helpers ─────────────────────────────────────────────────────

// newGetReq builds a GET request for an arbitrary absolute URL.
// Used by methods that need to set query parameters before calling do().
func newGetReq(ctx context.Context, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("client: build GET %s: %w", rawURL, err)
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// get issues an unauthenticated GET request and decodes the JSON response into result.
func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := newGetReq(ctx, c.cfg.BaseURL+path)
	if err != nil {
		return err
	}
	return c.do(req, result)
}

// postSigned issues an authenticated POST request carrying the EIP-712 signature headers.
func (c *Client) postSigned(ctx context.Context, path string, body any, sig []byte, nonce uint64, result any) error {
	bz, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("client: marshal POST body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(bz))
	if err != nil {
		return fmt.Errorf("client: build POST %s: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Sign", "0x"+hex.EncodeToString(sig))
	req.Header.Set("X-API-Nonce", strconv.FormatUint(nonce, 10))
	req.Header.Set("X-API-Chain", strconv.FormatUint(c.cfg.ChainID, 10))
	if c.cfg.APIKeyName != "" {
		req.Header.Set("X-API-Key", c.cfg.APIKeyName)
	}
	return c.do(req, result)
}

// deleteSigned issues an authenticated DELETE request carrying a JSON body and signature headers.
func (c *Client) deleteSigned(ctx context.Context, path string, body any, sig []byte, nonce uint64, result any) error {
	bz, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("client: marshal DELETE body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.cfg.BaseURL+path, bytes.NewReader(bz))
	if err != nil {
		return fmt.Errorf("client: build DELETE %s: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Sign", "0x"+hex.EncodeToString(sig))
	req.Header.Set("X-API-Nonce", strconv.FormatUint(nonce, 10))
	req.Header.Set("X-API-Chain", strconv.FormatUint(c.cfg.ChainID, 10))
	if c.cfg.APIKeyName != "" {
		req.Header.Set("X-API-Key", c.cfg.APIKeyName)
	}
	return c.do(req, result)
}

// ── shared history helpers ────────────────────────────────────────────────────
//
// The klines/trades/history endpoints follow a common query-param shape. These
// helpers avoid repeating the same URL-building and decoding code across spot.go
// and perps.go.

// getHistory issues a GET with url.Values built from a HistoryFilter and decodes
// the JSON response into result. Only fields the caller has set (non-zero) are
// included on the wire; Symbol/OrderID are engine-independent so callers pass
// them in the filter, but note: some endpoints (klines) ignore Symbol because
// it is already encoded in the path.
func (c *Client) getHistory(ctx context.Context, path string, filter HistoryFilter, result any) error {
	u, err := url.Parse(c.cfg.BaseURL + path)
	if err != nil {
		return fmt.Errorf("client: parse history URL: %w", err)
	}
	q := u.Query()
	if filter.Symbol != "" {
		q.Set("symbol", filter.Symbol)
	}
	if filter.StartTime > 0 {
		q.Set("startTime", strconv.FormatInt(filter.StartTime, 10))
	}
	if filter.EndTime > 0 {
		q.Set("endTime", strconv.FormatInt(filter.EndTime, 10))
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}
	u.RawQuery = q.Encode()
	req, err := newGetReq(ctx, u.String())
	if err != nil {
		return err
	}
	return c.do(req, result)
}

// klines issues GET /<base>/markets/<symbol>/klines.
func (c *Client) klines(
	ctx context.Context, base, symbol, interval string, filter HistoryFilter,
) ([]Candle, error) {
	if interval == "" {
		return nil, fmt.Errorf("client: klines: interval is required")
	}
	u, err := url.Parse(fmt.Sprintf("%s%s/markets/%s/klines", c.cfg.BaseURL, base, symbol))
	if err != nil {
		return nil, fmt.Errorf("client: parse klines URL: %w", err)
	}
	q := u.Query()
	q.Set("interval", interval)
	if filter.StartTime > 0 {
		q.Set("startTime", strconv.FormatInt(filter.StartTime, 10))
	}
	if filter.EndTime > 0 {
		q.Set("endTime", strconv.FormatInt(filter.EndTime, 10))
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}
	u.RawQuery = q.Encode()
	req, err := newGetReq(ctx, u.String())
	if err != nil {
		return nil, err
	}
	var result []Candle
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// publicTrades issues GET /<base>/markets/<symbol>/trades.
func (c *Client) publicTrades(
	ctx context.Context, base, symbol string, limit int,
) ([]PublicTrade, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s/markets/%s/trades", c.cfg.BaseURL, base, symbol))
	if err != nil {
		return nil, fmt.Errorf("client: parse public trades URL: %w", err)
	}
	if limit > 0 {
		q := u.Query()
		q.Set("limit", strconv.Itoa(limit))
		u.RawQuery = q.Encode()
	}
	req, err := newGetReq(ctx, u.String())
	if err != nil {
		return nil, err
	}
	var result []PublicTrade
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ordersHistory issues GET /<base>/accounts/<address>/orders/history.
func (c *Client) ordersHistory(
	ctx context.Context, base, address string, filter HistoryFilter,
) ([]Order, error) {
	var result []Order
	path := fmt.Sprintf("%s/accounts/%s/orders/history", base, address)
	if err := c.getHistory(ctx, path, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// userTrades issues GET /<base>/accounts/<address>/trades.
func (c *Client) userTrades(
	ctx context.Context, base, address string, filter HistoryFilter,
) ([]UserTrade, error) {
	var result []UserTrade
	path := fmt.Sprintf("%s/accounts/%s/trades", base, address)
	if err := c.getHistory(ctx, path, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// do executes req and decodes the response JSON into result.
// It first attempts to parse the response as an APIResponse envelope.
// If code != 0 it returns an ErrAPI regardless of the HTTP status code.
// If result is nil, the body is discarded.
func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("client: %s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("client: read response body: %w", err)
	}

	// Check for application-level error (code != 0).
	var envelope struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Msg     string          `json:"msg"`
		Error   string          `json:"error"`
		Data    json.RawMessage `json:"data"`
	}
	if jsonErr := json.Unmarshal(body, &envelope); jsonErr == nil && envelope.Code != 0 {
		msg := envelope.Message
		if msg == "" {
			msg = envelope.Msg
		}
		if msg == "" {
			msg = envelope.Error
		}
		if msg == "" {
			msg = string(body)
		}
		return &ErrAPI{Code: envelope.Code, Message: msg}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("client: HTTP %d from %s %s: %s",
			resp.StatusCode, req.Method, req.URL.Path, string(body))
	}

	if result == nil {
		return nil
	}

	// Prefer decoding envelope.Data when available; fall back to raw body.
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		return json.Unmarshal(envelope.Data, result)
	}
	return json.Unmarshal(body, result)
}
