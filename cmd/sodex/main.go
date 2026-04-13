// Command sodex is a CLI for the Sodex decentralized exchange.
//
// Run `sodex agent-help` for a comprehensive guide designed for AI agents.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/sodex-tech/sodex-go-sdk-public/client"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
	ptypes "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
	stypes "github.com/sodex-tech/sodex-go-sdk-public/spot/types"
	"github.com/sodex-tech/sodex-go-sdk-public/ws"
)

// ── Global state ──────────────────────────────────────────────────────────────

type globalFlags struct {
	baseURL    string
	privateKey string
	address    string
	chainID    uint64
	format     string
	testnet    bool
	apiKey     string
	accountID  uint64
}

type outputFormat string

const (
	formatPretty outputFormat = "pretty"
	formatTable  outputFormat = "table"
	formatJSON   outputFormat = "json"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildClient(g *globalFlags) (*client.Client, error) {
	// Resolve testnet mode: flag > env
	testnet := g.testnet
	if !testnet {
		if v := os.Getenv("SODEX_TESTNET"); v == "1" || v == "true" {
			testnet = true
		}
	}

	cfg := client.Config{
		BaseURL:    firstNonEmpty(g.baseURL, os.Getenv("SODEX_BASE_URL")),
		ChainID:    g.chainID,
		APIKeyName: firstNonEmpty(g.apiKey, os.Getenv("SODEX_API_KEY")),
	}

	// When testnet is enabled, override defaults (only if not explicitly set by user)
	if testnet {
		if cfg.BaseURL == "" || cfg.BaseURL == client.DefaultBaseURL {
			cfg.BaseURL = client.TestnetBaseURL
		}
		if cfg.ChainID == client.DefaultChainID {
			cfg.ChainID = client.TestnetChainID
		}
	}

	pkHex := firstNonEmpty(g.privateKey, os.Getenv("SODEX_PRIVATE_KEY"))
	if pkHex != "" {
		pkHex = strings.TrimPrefix(pkHex, "0x")
		pk, err := crypto.HexToECDSA(pkHex)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
		cfg.PrivateKey = pk
	}
	return client.New(cfg), nil
}

// resolveAccountID returns the account ID from flag > env, or 0 if not set.
func resolveAccountID(g *globalFlags) (uint64, error) {
	if g.accountID != 0 {
		return g.accountID, nil
	}
	if v := os.Getenv("SODEX_ACCOUNT_ID"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid SODEX_ACCOUNT_ID %q: %w", v, err)
		}
		return id, nil
	}
	return 0, fmt.Errorf("--account-id is required (or set SODEX_ACCOUNT_ID)")
}

// resolveAddress returns the address to query, in priority order:
// explicit --address flag > SODEX_ADDRESS env > derived from private key.
func resolveAddress(g *globalFlags) (string, error) {
	if addr := firstNonEmpty(g.address, os.Getenv("SODEX_ADDRESS")); addr != "" {
		return addr, nil
	}
	pkHex := firstNonEmpty(g.privateKey, os.Getenv("SODEX_PRIVATE_KEY"))
	if pkHex == "" {
		return "", fmt.Errorf("provide --address, SODEX_ADDRESS, or --private-key to derive the address")
	}
	pkHex = strings.TrimPrefix(pkHex, "0x")
	pk, err := crypto.HexToECDSA(pkHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	return crypto.PubkeyToAddress(pk.PublicKey).Hex(), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func resolveEngine(args []string) string {
	if len(args) > 0 {
		return strings.ToLower(args[0])
	}
	return "spot"
}

func resolveFormat(s string) (outputFormat, error) {
	switch strings.ToLower(s) {
	case "pretty", "":
		return formatPretty, nil
	case "table":
		return formatTable, nil
	case "json":
		return formatJSON, nil
	default:
		return "", fmt.Errorf("unknown format %q: must be one of pretty, table, json", s)
	}
}

func parseOrderSide(s string) (enums.OrderSide, error) {
	switch strings.ToLower(s) {
	case "buy", "b":
		return enums.OrderSideBuy, nil
	case "sell", "s":
		return enums.OrderSideSell, nil
	default:
		return enums.OrderSideUnknown, fmt.Errorf("invalid side %q: must be buy or sell", s)
	}
}

func parseOrderType(s string) (enums.OrderType, error) {
	switch strings.ToLower(s) {
	case "limit", "l":
		return enums.OrderTypeLimit, nil
	case "market", "m":
		return enums.OrderTypeMarket, nil
	default:
		return enums.OrderTypeUnknown, fmt.Errorf("invalid type %q: must be limit or market", s)
	}
}

func parseTIF(s string) enums.TimeInForce {
	switch strings.ToLower(s) {
	case "gtc":
		return enums.TimeInForceGTC
	case "ioc":
		return enums.TimeInForceIOC
	case "fok":
		return enums.TimeInForceFOK
	case "gtx":
		return enums.TimeInForceGTX
	default:
		return enums.TimeInForceGTC
	}
}

func parseMarginMode(s string) enums.MarginMode {
	switch strings.ToLower(s) {
	case "cross":
		return enums.MarginModeCross
	default:
		return enums.MarginModeIsolated
	}
}

// resolveSymbol fetches the symbol list and matches a user-provided symbol
// name (case-insensitive) against internal name or display name.
// Returns (symbolID, internalName).
func resolveSymbol(ctx context.Context, c *client.Client, engine, symbol string) (uint64, string, error) {
	var symbols []client.Symbol
	var err error
	switch engine {
	case "spot":
		symbols, err = c.SpotSymbols(ctx)
	case "perps":
		symbols, err = c.PerpsSymbols(ctx)
	default:
		return 0, "", fmt.Errorf("unknown engine %q: must be spot or perps", engine)
	}
	if err != nil {
		return 0, "", fmt.Errorf("fetch %s symbols: %w", engine, err)
	}
	upper := strings.ToUpper(symbol)
	for _, s := range symbols {
		if strings.EqualFold(s.Symbol, symbol) ||
			strings.EqualFold(s.DisplayName, symbol) ||
			strings.EqualFold(s.DisplayName, strings.ReplaceAll(upper, "-", "/")) {
			return s.SymbolID, s.Symbol, nil
		}
	}
	return 0, "", fmt.Errorf("symbol %q not found in %s markets; run `sodex markets %s` to see available symbols", symbol, engine, engine)
}

// resolveSymbolID is a convenience wrapper that only returns the ID.
func resolveSymbolID(ctx context.Context, c *client.Client, engine, symbol string) (uint64, error) {
	id, _, err := resolveSymbol(ctx, c, engine, symbol)
	return id, err
}

// resolveSymbolName resolves user input to the internal symbol name for API paths.
func resolveSymbolName(ctx context.Context, c *client.Client, engine, symbol string) (string, error) {
	_, name, err := resolveSymbol(ctx, c, engine, symbol)
	return name, err
}

// printJSON marshals v to indented JSON on stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ── Print helpers ─────────────────────────────────────────────────────────────

func printSymbols(symbols []client.Symbol, f outputFormat) error {
	switch f {
	case formatJSON:
		return printJSON(symbols)
	case formatTable:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tDISPLAY\tBASE\tQUOTE\tSTATUS\tPRICE_PREC\tQTY_PREC")
		for _, s := range symbols {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%d\t%d\n",
				s.SymbolID, s.Symbol, s.DisplayName, s.BaseAsset, s.QuoteAsset,
				s.Status, s.PricePrecision, s.QuantityPrecision)
		}
		return w.Flush()
	default:
		for _, s := range symbols {
			fmt.Printf("%-20s  id=%-6d  name=%-20s  base=%-10s  quote=%-10s  status=%s\n",
				s.DisplayName, s.SymbolID, s.Symbol, s.BaseAsset, s.QuoteAsset, s.Status)
		}
		return nil
	}
}

func printTickers(tickers []client.Ticker, f outputFormat) error {
	switch f {
	case formatJSON:
		return printJSON(tickers)
	case formatTable:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SYMBOL\tLAST\tBID\tASK\tCHANGE%\tVOLUME")
		for _, t := range tickers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.2f%%\t%s\n",
				t.Symbol, t.LastPrice, t.BidPrice, t.AskPrice,
				t.PriceChangePercent, t.Volume)
		}
		return w.Flush()
	default:
		for _, t := range tickers {
			fmt.Printf("%-20s  last=%-14s  bid=%-14s  ask=%-14s  change=%-10s  vol=%s\n",
				t.Symbol, t.LastPrice, t.BidPrice, t.AskPrice,
				fmt.Sprintf("%.2f%%", t.PriceChangePercent), t.Volume)
		}
		return nil
	}
}

func printOrderBook(ob *client.OrderBook, depth int, f outputFormat) error {
	if f == formatJSON {
		return printJSON(ob)
	}
	maxRows := depth
	if maxRows <= 0 || maxRows > len(ob.Asks) {
		maxRows = len(ob.Asks)
	}
	if maxRows > len(ob.Bids) {
		maxRows = len(ob.Bids)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "Order Book: %s\n\n", ob.Symbol)
	fmt.Fprintln(w, "ASKS (price / qty)")
	for i := maxRows - 1; i >= 0; i-- {
		fmt.Fprintf(w, "  %s\t%s\n", ob.Asks[i].Price, ob.Asks[i].Quantity)
	}
	fmt.Fprintln(w, "  ─────────────────")
	fmt.Fprintln(w, "BIDS (price / qty)")
	for i := 0; i < maxRows; i++ {
		fmt.Fprintf(w, "  %s\t%s\n", ob.Bids[i].Price, ob.Bids[i].Quantity)
	}
	return w.Flush()
}

func printBalances(balances []client.Balance, f outputFormat) error {
	switch f {
	case formatJSON:
		return printJSON(balances)
	case formatTable:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COIN_ID\tCOIN\tTOTAL\tLOCKED")
		for _, b := range balances {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", b.CoinID, b.Coin, b.Total, b.Locked)
		}
		return w.Flush()
	default:
		for _, b := range balances {
			fmt.Printf("%-12s  id=%-4d  total=%-18s  locked=%-18s\n",
				b.Coin, b.CoinID, b.Total, b.Locked)
		}
		return nil
	}
}

func printOrders(orders []client.Order, f outputFormat) error {
	switch f {
	case formatJSON:
		return printJSON(orders)
	case formatTable:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ORDER_ID\tCL_ORD_ID\tSYMBOL\tSIDE\tTYPE\tPRICE\tQTY\tFILLED\tSTATUS")
		for _, o := range orders {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				o.OrderID, o.ClOrdID, o.Symbol, o.Side, o.Type,
				o.Price, o.OrigQty, o.ExecutedQty, o.Status)
		}
		return w.Flush()
	default:
		for _, o := range orders {
			fmt.Printf("%-12d  cl=%s  %s  %s/%s  price=%-14s  qty=%-12s  filled=%-12s  status=%s\n",
				o.OrderID, o.ClOrdID, o.Symbol, o.Side, o.Type,
				o.Price, o.OrigQty, o.ExecutedQty, o.Status)
		}
		return nil
	}
}

func printPositions(positions []client.Position, f outputFormat) error {
	switch f {
	case formatJSON:
		return printJSON(positions)
	case formatTable:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SYMBOL\tSIDE\tQTY\tENTRY\tMARK\tLIQ\tPNL\tLEVERAGE\tMARGIN")
		for _, p := range positions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%dx\t%s\n",
				p.Symbol, p.PositionSide, p.Quantity, p.EntryPrice,
				p.MarkPrice, p.LiqPrice, p.UnrealizedPnl, p.Leverage, p.Margin)
		}
		return w.Flush()
	default:
		for _, p := range positions {
			fmt.Printf("%-20s  side=%-6s  qty=%-12s  entry=%-12s  mark=%-12s  pnl=%-12s  lev=%dx\n",
				p.Symbol, p.PositionSide, p.Quantity, p.EntryPrice,
				p.MarkPrice, p.UnrealizedPnl, p.Leverage)
		}
		return nil
	}
}

// ── Commands ──────────────────────────────────────────────────────────────────

func newMarketsCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "markets [spot|perps]",
		Short: "List available trading pairs",
		Long:  "List all trading pairs for the spot (Spark) or perps (Bolt) engine. Defaults to spot.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := resolveEngine(args)
			f, err := resolveFormat(firstNonEmpty(format, g.format))
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()
			var symbols []client.Symbol
			switch engine {
			case "spot":
				symbols, err = c.SpotSymbols(ctx)
			case "perps":
				symbols, err = c.PerpsSymbols(ctx)
			default:
				return fmt.Errorf("engine must be spot or perps")
			}
			if err != nil {
				return err
			}
			return printSymbols(symbols, f)
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newTickersCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "tickers [spot|perps]",
		Short: "Show 24-hour ticker stats",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := resolveEngine(args)
			f, err := resolveFormat(firstNonEmpty(format, g.format))
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()
			var tickers []client.Ticker
			switch engine {
			case "spot":
				tickers, err = c.SpotTickers(ctx)
			case "perps":
				tickers, err = c.PerpsTickers(ctx)
			default:
				return fmt.Errorf("engine must be spot or perps")
			}
			if err != nil {
				return err
			}
			return printTickers(tickers, f)
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newOrderBookCmd(g *globalFlags) *cobra.Command {
	var (
		depth  int
		format string
	)
	cmd := &cobra.Command{
		Use:   "orderbook [spot|perps] SYMBOL",
		Short: "Show order book for a symbol",
		Long: `Show the order book for a trading pair.
The engine argument defaults to spot. SYMBOL is the human-readable pair (e.g. BTC-USDT).

Examples:
  sodex orderbook BTC-USDT
  sodex orderbook perps BTC-USDT --depth 10`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var engine, symbol string
			if len(args) == 1 {
				engine = "spot"
				symbol = args[0]
			} else {
				engine = strings.ToLower(args[0])
				if engine != "spot" && engine != "perps" {
					// treat both as symbol + default engine
					engine = "spot"
					symbol = args[0]
				} else {
					symbol = args[1]
				}
			}
			f, err := resolveFormat(firstNonEmpty(format, g.format))
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()

			// Resolve to internal symbol name for the API path.
			internalName, err := resolveSymbolName(ctx, c, engine, symbol)
			if err != nil {
				return err
			}

			var ob *client.OrderBook
			switch engine {
			case "spot":
				ob, err = c.SpotOrderBook(ctx, internalName, depth)
			case "perps":
				ob, err = c.PerpsOrderBook(ctx, internalName, depth)
			}
			if err != nil {
				return err
			}
			return printOrderBook(ob, depth, f)
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 20, "Number of levels to display")
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newAccountIDCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "account-id [ADDRESS]",
		Short: "Look up account ID for a wallet address",
		Long: `Look up the account ID and user ID for a wallet address.

ADDRESS can be a positional argument, --address, or derived from --private-key.

Examples:
  sodex account-id 0xABC...
  sodex account-id --address 0xABC...
  sodex account-id   # derives from SODEX_PRIVATE_KEY`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g2 := *g
			if len(args) >= 1 {
				g2.address = args[0]
			}
			addr, err := resolveAddress(&g2)
			if err != nil {
				return err
			}
			f, err := resolveFormat(firstNonEmpty(format, g.format))
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			info, err := c.SpotAccountInfo(context.Background(), addr)
			if err != nil {
				return err
			}
			if info.AccountID == 0 && info.UserID == 0 {
				return fmt.Errorf("no account found for address %s", addr)
			}
			switch f {
			case formatJSON:
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			case formatTable:
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "ADDRESS\tACCOUNT_ID\tUSER_ID")
				fmt.Fprintf(w, "%s\t%d\t%d\n", info.Address, info.AccountID, info.UserID)
				return w.Flush()
			default:
				fmt.Printf("Address:    %s\n", info.Address)
				fmt.Printf("Account ID: %d\n", info.AccountID)
				fmt.Printf("User ID:    %d\n", info.UserID)
				return nil
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newBalanceCmd(g *globalFlags) *cobra.Command {
	var (
		format  string
		addrArg string
	)
	cmd := &cobra.Command{
		Use:   "balance [spot|perps] [ADDRESS]",
		Short: "Show account balances",
		Long: `Show asset balances for a wallet address.

ADDRESS can be provided as a positional argument, via --address, or derived
from --private-key. Engine defaults to spot.

Examples:
  sodex balance 0xABC...
  sodex balance perps --address 0xABC...
  sodex balance spot --private-key 0x...`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := "spot"
			if len(args) >= 1 && (args[0] == "spot" || args[0] == "perps") {
				engine = args[0]
				args = args[1:]
			}
			if len(args) >= 1 {
				addrArg = args[0]
			}
			g2 := *g
			if addrArg != "" {
				g2.address = addrArg
			}
			addr, err := resolveAddress(&g2)
			if err != nil {
				return err
			}
			f, err := resolveFormat(firstNonEmpty(format, g.format))
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()
			var balances []client.Balance
			switch engine {
			case "spot":
				balances, err = c.SpotBalances(ctx, addr)
			case "perps":
				balances, err = c.PerpsBalances(ctx, addr)
			default:
				return fmt.Errorf("engine must be spot or perps")
			}
			if err != nil {
				return err
			}
			return printBalances(balances, f)
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newOrdersCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orders",
		Short: "Order management (list, place, cancel)",
	}
	cmd.AddCommand(
		newOrdersListCmd(g),
		newOrdersPlaceCmd(g),
		newOrdersCancelCmd(g),
	)
	return cmd
}

func newOrdersListCmd(g *globalFlags) *cobra.Command {
	var (
		format  string
		addrArg string
	)
	cmd := &cobra.Command{
		Use:   "list [spot|perps] [ADDRESS]",
		Short: "List open orders",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := "spot"
			if len(args) >= 1 && (args[0] == "spot" || args[0] == "perps") {
				engine = args[0]
				args = args[1:]
			}
			if len(args) >= 1 {
				addrArg = args[0]
			}
			g2 := *g
			if addrArg != "" {
				g2.address = addrArg
			}
			addr, err := resolveAddress(&g2)
			if err != nil {
				return err
			}
			f, err := resolveFormat(firstNonEmpty(format, g.format))
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()
			var orders []client.Order
			switch engine {
			case "spot":
				orders, err = c.SpotOrders(ctx, addr)
			case "perps":
				orders, err = c.PerpsOrders(ctx, addr)
			default:
				return fmt.Errorf("engine must be spot or perps")
			}
			if err != nil {
				return err
			}
			return printOrders(orders, f)
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newOrdersPlaceCmd(g *globalFlags) *cobra.Command {
	var (
		symbol     string
		sideStr    string
		typeStr    string
		priceStr   string
		qtyStr     string
		tifStr     string
		reduceOnly bool
		format     string
	)
	cmd := &cobra.Command{
		Use:   "place [spot|perps]",
		Short: "Place a new order",
		Long: `Place a new order on the spot or perps engine.

The symbol is resolved by name (e.g. BTC-USDT) against the live market list.
A private key is required (--private-key or SODEX_PRIVATE_KEY).

Examples:
  sodex orders place spot --symbol BTC-USDT --side buy --type limit --price 95000 --qty 0.01
  sodex orders place perps --symbol BTC-USDT --side sell --type market --qty 0.1 --reduce-only
  sodex orders place perps --symbol ETH-USDT --side buy --price 3500 --qty 1 --tif gtx`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := resolveEngine(args)
			if symbol == "" {
				return fmt.Errorf("--symbol is required (e.g. --symbol BTC-USDT)")
			}
			accountID, err := resolveAccountID(g)
			if err != nil {
				return err
			}
			if sideStr == "" {
				return fmt.Errorf("--side is required (buy or sell)")
			}
			if qtyStr == "" {
				return fmt.Errorf("--qty is required")
			}
			side, err := parseOrderSide(sideStr)
			if err != nil {
				return err
			}
			orderType, err := parseOrderType(typeStr)
			if err != nil {
				return err
			}
			tif := parseTIF(tifStr)
			qty, err := decimal.NewFromString(qtyStr)
			if err != nil || qty.IsZero() || qty.IsNegative() {
				return fmt.Errorf("invalid --qty %q: must be a positive decimal", qtyStr)
			}
			var price decimal.Decimal
			if priceStr != "" {
				price, err = decimal.NewFromString(priceStr)
				if err != nil || price.IsZero() || price.IsNegative() {
					return fmt.Errorf("invalid --price %q: must be a positive decimal", priceStr)
				}
			} else if orderType == enums.OrderTypeLimit {
				return fmt.Errorf("--price is required for limit orders")
			}

			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()

			symbolID, err := resolveSymbolID(ctx, c, engine, symbol)
			if err != nil {
				return err
			}

			clOrdID := fmt.Sprintf("sodex-cli-%d", time.Now().UnixMilli())
			f, _ := resolveFormat(firstNonEmpty(format, g.format))

			var results []client.PlaceOrderResult
			switch engine {
			case "spot":
				item := &stypes.BatchNewOrderItem{
					SymbolID:    symbolID,
					ClOrdID:     clOrdID,
					Side:        side,
					Type:        orderType,
					TimeInForce: tif,
					Quantity:    &qty,
				}
				if !price.IsZero() {
					item.Price = &price
				}
				results, err = c.PlaceSpotOrders(ctx, &stypes.BatchNewOrderRequest{
					AccountID: accountID,
					Orders:    []*stypes.BatchNewOrderItem{item},
				})
			case "perps":
				rawOrder := &ptypes.RawOrder{
					ClOrdID:      clOrdID,
					Modifier:     enums.OrderModifierNormal,
					Side:         side,
					Type:         orderType,
					TimeInForce:  tif,
					Quantity:     &qty,
					PositionSide: enums.PositionSideBoth,
					ReduceOnly:   reduceOnly,
				}
				if !price.IsZero() {
					rawOrder.Price = &price
				}
				results, err = c.PlacePerpsOrder(ctx, &ptypes.NewOrderRequest{
					AccountID: accountID,
					SymbolID:  symbolID,
					Orders:    []*ptypes.RawOrder{rawOrder},
				})
			default:
				return fmt.Errorf("engine must be spot or perps")
			}
			if err != nil {
				return err
			}
			if f == formatJSON {
				return printJSON(results)
			}
			for _, r := range results {
				fmt.Printf("placed  clOrdID=%-30s  orderID=%-12d  status=%s\n",
					r.ClOrdID, r.OrderID, r.Status)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&symbol, "symbol", "", "Trading pair (e.g. BTC-USDT) [required]")
	cmd.Flags().StringVar(&sideStr, "side", "", "Order side: buy or sell [required]")
	cmd.Flags().StringVar(&typeStr, "type", "limit", "Order type: limit or market")
	cmd.Flags().StringVar(&priceStr, "price", "", "Limit price (required for limit orders)")
	cmd.Flags().StringVar(&qtyStr, "qty", "", "Order quantity [required]")
	cmd.Flags().StringVar(&tifStr, "tif", "gtc", "Time-in-force: gtc|ioc|fok|gtx")
	cmd.Flags().BoolVar(&reduceOnly, "reduce-only", false, "Reduce-only flag (perps only)")
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newOrdersCancelCmd(g *globalFlags) *cobra.Command {
	var (
		symbolID uint64
		orderID  uint64
		clOrdID  string
		format   string
	)
	cmd := &cobra.Command{
		Use:   "cancel [spot|perps]",
		Short: "Cancel an existing order",
		Long: `Cancel a resting order by order ID or client order ID.
Either --order-id or --cl-ord-id must be provided.

Examples:
  sodex orders cancel spot --symbol-id 42 --order-id 99999
  sodex orders cancel perps --symbol-id 1 --cl-ord-id sodex-cli-1711900800000`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := resolveEngine(args)
			accountID, err := resolveAccountID(g)
			if err != nil {
				return err
			}
			if symbolID == 0 {
				return fmt.Errorf("--symbol-id is required")
			}
			if orderID == 0 && clOrdID == "" {
				return fmt.Errorf("provide --order-id or --cl-ord-id")
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()
			f, _ := resolveFormat(firstNonEmpty(format, g.format))

			var results []client.CancelOrderResult
			switch engine {
			case "spot":
				item := &stypes.BatchCancelOrderItem{
					SymbolID: symbolID,
					ClOrdID:  clOrdID,
				}
				if orderID != 0 {
					item.OrderID = &orderID
				}
				results, err = c.CancelSpotOrders(ctx, &stypes.BatchCancelOrderRequest{
					AccountID: accountID,
					Cancels:   []*stypes.BatchCancelOrderItem{item},
				})
			case "perps":
				cancel := &ptypes.CancelOrder{
					SymbolID: symbolID,
				}
				if orderID != 0 {
					cancel.OrderID = &orderID
				}
				if clOrdID != "" {
					cancel.ClOrdID = &clOrdID
				}
				results, err = c.CancelPerpsOrders(ctx, &ptypes.CancelOrderRequest{
					AccountID: accountID,
					Cancels:   []*ptypes.CancelOrder{cancel},
				})
			default:
				return fmt.Errorf("engine must be spot or perps")
			}
			if err != nil {
				return err
			}
			if f == formatJSON {
				return printJSON(results)
			}
			for _, r := range results {
				oid := "—"
				if r.OrderID != nil {
					oid = strconv.FormatUint(*r.OrderID, 10)
				}
				fmt.Printf("cancelled  orderID=%-12s  clOrdID=%-30s  status=%s\n",
					oid, r.ClOrdID, r.Status)
			}
			return nil
		},
	}
	cmd.Flags().Uint64Var(&symbolID, "symbol-id", 0, "Numeric symbol ID [required]")
	cmd.Flags().Uint64Var(&orderID, "order-id", 0, "Order ID to cancel")
	cmd.Flags().StringVar(&clOrdID, "cl-ord-id", "", "Client order ID to cancel")
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newPositionsCmd(g *globalFlags) *cobra.Command {
	var (
		format  string
		addrArg string
	)
	cmd := &cobra.Command{
		Use:   "positions [ADDRESS]",
		Short: "Show open perpetuals positions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) >= 1 {
				addrArg = args[0]
			}
			g2 := *g
			if addrArg != "" {
				g2.address = addrArg
			}
			addr, err := resolveAddress(&g2)
			if err != nil {
				return err
			}
			f, err := resolveFormat(firstNonEmpty(format, g.format))
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			positions, err := c.PerpsPositions(context.Background(), addr)
			if err != nil {
				return err
			}
			return printPositions(positions, f)
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|table|json")
	return cmd
}

func newLeverageCmd(g *globalFlags) *cobra.Command {
	var (
		modeStr string
	)
	cmd := &cobra.Command{
		Use:   "leverage SYMBOL VALUE",
		Short: "Update leverage for a perps position",
		Long: `Update the leverage multiplier for a perpetuals position.

Examples:
  sodex leverage BTC-USDT 10
  sodex leverage ETH-USDT 5 --mode cross`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := args[0]
			lev, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil || lev == 0 {
				return fmt.Errorf("invalid leverage value %q: must be a positive integer", args[1])
			}
			accountID, err := resolveAccountID(g)
			if err != nil {
				return err
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			ctx := context.Background()
			symbolID, err := resolveSymbolID(ctx, c, "perps", symbol)
			if err != nil {
				return err
			}
			result, err := c.UpdateLeverage(ctx, &ptypes.UpdateLeverageRequest{
				AccountID:  accountID,
				SymbolID:   symbolID,
				Leverage:   uint32(lev),
				MarginMode: parseMarginMode(modeStr),
			})
			if err != nil {
				return err
			}
			if result != nil {
				fmt.Printf("leverage updated: %s  leverage=%dx  mode=%s\n",
					result.Symbol, result.Leverage, result.MarginMode)
			} else {
				fmt.Printf("leverage updated: %s  leverage=%dx\n", symbol, lev)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&modeStr, "mode", "isolated", "Margin mode: isolated or cross")
	return cmd
}

func newTransferCmd(g *globalFlags) *cobra.Command {
	var (
		fromAccountID uint64
		toAccountID   uint64
		coinID        uint64
		amountStr     string
		transferType  int
	)
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer assets between accounts",
		Long: `Transfer assets between accounts using the spot engine.

Examples:
  sodex transfer --from 1001 --to 1002 --coin 1 --amount 100
  sodex transfer --from 1001 --to 1002 --coin 1 --amount 50 --type 4`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromAccountID == 0 {
				return fmt.Errorf("--from is required")
			}
			if toAccountID == 0 {
				return fmt.Errorf("--to is required")
			}
			if coinID == 0 {
				return fmt.Errorf("--coin is required")
			}
			amount, err := decimal.NewFromString(amountStr)
			if err != nil || amount.IsZero() || amount.IsNegative() {
				return fmt.Errorf("invalid --amount %q: must be a positive decimal", amountStr)
			}
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			req := &ctypes.TransferAssetRequest{
				ID:            uint64(time.Now().UnixMilli()),
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				CoinID:        coinID,
				Amount:        amount,
				Type:          enums.TransferAssetType(transferType),
			}
			if err := c.SpotTransfer(context.Background(), req); err != nil {
				return err
			}
			fmt.Printf("transfer submitted: from=%d to=%d coin=%d amount=%s\n",
				fromAccountID, toAccountID, coinID, amount.String())
			return nil
		},
	}
	cmd.Flags().Uint64Var(&fromAccountID, "from", 0, "Source account ID [required]")
	cmd.Flags().Uint64Var(&toAccountID, "to", 0, "Destination account ID [required]")
	cmd.Flags().Uint64Var(&coinID, "coin", 0, "Coin ID [required]")
	cmd.Flags().StringVar(&amountStr, "amount", "", "Amount to transfer [required]")
	cmd.Flags().IntVar(&transferType, "type", 4, "Transfer type (0=EVM deposit, 1=perps deposit, 4=internal, etc.)")
	return cmd
}

func newAgentHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "agent-help",
		Short: "Print comprehensive usage guide for AI agents",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(agentHelpText)
		},
	}
}

// ── Subscribe (WebSocket) ────────────────────────────────────────────────────

func resolveBaseURL(g *globalFlags) string {
	testnet := g.testnet
	if !testnet {
		if v := os.Getenv("SODEX_TESTNET"); v == "1" || v == "true" {
			testnet = true
		}
	}
	if u := firstNonEmpty(g.baseURL, os.Getenv("SODEX_BASE_URL")); u != "" {
		return u
	}
	if testnet {
		return client.TestnetBaseURL
	}
	return client.DefaultBaseURL
}

func newSubscribeCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to real-time WebSocket feeds",
		Long: `Subscribe to real-time data from the Sodex WebSocket API.

Channels: trades, orderbook, ticker, candle, bbo, mark-price,
          order-updates, fills

Examples:
  sodex --testnet subscribe trades spot BTC-USD
  sodex subscribe orderbook perps ETH-USD --depth 10
  sodex subscribe candle spot BTC/USDC --interval 1m
  sodex subscribe order-updates spot --user 0xYourAddress
  sodex subscribe ticker perps --format json`,
	}
	cmd.AddCommand(
		newSubTradesCmd(g),
		newSubOrderbookCmd(g),
		newSubTickerCmd(g),
		newSubCandleCmd(g),
		newSubBboCmd(g),
		newSubMarkPriceCmd(g),
		newSubOrderUpdatesCmd(g),
		newSubFillsCmd(g),
	)
	return cmd
}

// runSubscription is the shared logic for all subscribe subcommands.
func runSubscription(g *globalFlags, engine string, params ws.SubscribeParams, printer func(ws.Push)) error {
	baseURL := resolveBaseURL(g)
	wsc, err := ws.NewClient(baseURL, engine)
	if err != nil {
		return err
	}
	wsc.OnError(func(err error) {
		fmt.Fprintf(os.Stderr, "ws error: %v\n", err)
	})

	if _, err := wsc.Subscribe(params, printer); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	go func() {
		sig := make(chan os.Signal, 1)
		// signal.Notify not needed — Connect blocks until context cancelled
		_ = sig
	}()

	return wsc.Connect(ctx)
}

func newSubTradesCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "trades [spot|perps] SYMBOL",
		Short: "Subscribe to real-time trades",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, symbol := resolveEngineAndSymbol(args)
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			name, err := resolveSymbolName(context.Background(), c, engine, symbol)
			if err != nil {
				return err
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, engine, ws.SubscribeParams{
				Channel: ws.ChannelTrade,
				Symbols: []string{name},
			}, tradesPrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	return cmd
}

func newSubOrderbookCmd(g *globalFlags) *cobra.Command {
	var (
		format string
		depth  int
	)
	cmd := &cobra.Command{
		Use:   "orderbook [spot|perps] SYMBOL",
		Short: "Subscribe to order book updates",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, symbol := resolveEngineAndSymbol(args)
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			name, err := resolveSymbolName(context.Background(), c, engine, symbol)
			if err != nil {
				return err
			}
			level := 10
			if depth > 0 {
				level = depth
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, engine, ws.SubscribeParams{
				Channel: ws.ChannelL4Book,
				Symbol:  name,
				Level:   level,
			}, orderbookPrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	cmd.Flags().IntVar(&depth, "depth", 10, "Order book depth (10, 20, 100, 500, 1000)")
	return cmd
}

func newSubTickerCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "ticker [spot|perps] [SYMBOL...]",
		Short: "Subscribe to ticker updates",
		Long:  "Subscribe to ticker updates. Without symbols, subscribes to all tickers.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := resolveEngine(args)
			symbols := args
			if len(args) > 0 && (args[0] == "spot" || args[0] == "perps") {
				symbols = args[1:]
			}
			ch := ws.ChannelAllTicker
			params := ws.SubscribeParams{Channel: ch}
			if len(symbols) > 0 {
				c, err := buildClient(g)
				if err != nil {
					return err
				}
				var names []string
				for _, s := range symbols {
					name, err := resolveSymbolName(context.Background(), c, engine, s)
					if err != nil {
						return err
					}
					names = append(names, name)
				}
				params.Channel = ws.ChannelTicker
				params.Symbols = names
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, engine, params, tickerPrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	return cmd
}

func newSubCandleCmd(g *globalFlags) *cobra.Command {
	var (
		format   string
		interval string
	)
	cmd := &cobra.Command{
		Use:   "candle [spot|perps] SYMBOL",
		Short: "Subscribe to candlestick updates",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, symbol := resolveEngineAndSymbol(args)
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			name, err := resolveSymbolName(context.Background(), c, engine, symbol)
			if err != nil {
				return err
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, engine, ws.SubscribeParams{
				Channel:  ws.ChannelCandle,
				Symbol:   name,
				Interval: interval,
			}, candlePrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	cmd.Flags().StringVar(&interval, "interval", "1m", "Candle interval: 1m|5m|15m|30m|1h|4h|1D|1W")
	return cmd
}

func newSubBboCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "bbo [spot|perps] SYMBOL",
		Short: "Subscribe to best bid/offer updates",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, symbol := resolveEngineAndSymbol(args)
			c, err := buildClient(g)
			if err != nil {
				return err
			}
			name, err := resolveSymbolName(context.Background(), c, engine, symbol)
			if err != nil {
				return err
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, engine, ws.SubscribeParams{
				Channel: ws.ChannelBookTicker,
				Symbols: []string{name},
			}, bboPrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	return cmd
}

func newSubMarkPriceCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "mark-price [SYMBOL...]",
		Short: "Subscribe to mark price updates (perps only)",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ch := ws.ChannelAllMarkPrice
			params := ws.SubscribeParams{Channel: ch}
			if len(args) > 0 {
				c, err := buildClient(g)
				if err != nil {
					return err
				}
				var names []string
				for _, s := range args {
					name, err := resolveSymbolName(context.Background(), c, "perps", s)
					if err != nil {
						return err
					}
					names = append(names, name)
				}
				params.Channel = ws.ChannelMarkPrice
				params.Symbols = names
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, "perps", params, markPricePrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	return cmd
}

func newSubOrderUpdatesCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "order-updates [spot|perps]",
		Short: "Subscribe to real-time order status changes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := resolveEngine(args)
			addr, err := resolveAddress(g)
			if err != nil {
				return err
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, engine, ws.SubscribeParams{
				Channel: ws.ChannelAccountOrderUpd,
				User:    addr,
			}, orderUpdatesPrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	return cmd
}

func newSubFillsCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "fills [spot|perps]",
		Short: "Subscribe to real-time trade fills",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := resolveEngine(args)
			addr, err := resolveAddress(g)
			if err != nil {
				return err
			}
			f, _ := resolveFormat(firstNonEmpty(format, g.format))
			return runSubscription(g, engine, ws.SubscribeParams{
				Channel: ws.ChannelAccountTrade,
				User:    addr,
			}, fillsPrinter(f))
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format: pretty|json")
	return cmd
}

// ── Arg parsing helper ───────────────────────────────────────────────────────

func resolveEngineAndSymbol(args []string) (string, string) {
	if len(args) == 1 {
		return "spot", args[0]
	}
	engine := strings.ToLower(args[0])
	if engine == "spot" || engine == "perps" {
		return engine, args[1]
	}
	return "spot", args[0]
}

// ── WS print helpers ─────────────────────────────────────────────────────────

func tradesPrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var trades []ws.Trade
		if err := json.Unmarshal(p.Data, &trades); err != nil {
			return
		}
		for _, t := range trades {
			ts := time.UnixMilli(t.TradeTime).Format("15:04:05.000")
			fmt.Printf("%s  %-20s  %-4s  price=%-14s  qty=%-12s\n",
				ts, t.Symbol, t.Side, t.Price, t.Quantity)
		}
	}
}

func orderbookPrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var book ws.L2Book
		if err := json.Unmarshal(p.Data, &book); err != nil {
			return
		}
		// Clear screen and print
		fmt.Print("\033[2J\033[H")
		fmt.Printf("Order Book: %s  (update=%d)\n\n", book.Symbol, book.UpdateID)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ASKS (price / qty)")
		maxRows := len(book.Asks)
		if maxRows > 20 {
			maxRows = 20
		}
		for i := maxRows - 1; i >= 0; i-- {
			if len(book.Asks[i]) >= 2 {
				fmt.Fprintf(w, "  %s\t%s\n", book.Asks[i][0], book.Asks[i][1])
			}
		}
		fmt.Fprintln(w, "  ─────────────────")
		fmt.Fprintln(w, "BIDS (price / qty)")
		maxRows = len(book.Bids)
		if maxRows > 20 {
			maxRows = 20
		}
		for i := 0; i < maxRows; i++ {
			if len(book.Bids[i]) >= 2 {
				fmt.Fprintf(w, "  %s\t%s\n", book.Bids[i][0], book.Bids[i][1])
			}
		}
		w.Flush()
	}
}

func tickerPrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var tickers []ws.Ticker
		if err := json.Unmarshal(p.Data, &tickers); err != nil {
			return
		}
		for _, t := range tickers {
			fmt.Printf("%-20s  last=%-14s  bid=%-14s  ask=%-14s  change=%.2f%%  vol=%s\n",
				t.Symbol, t.LastPrice, t.BidPrice, t.AskPrice,
				t.PriceChangePercent, t.Volume)
		}
	}
}

func candlePrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var candles []ws.Candle
		if err := json.Unmarshal(p.Data, &candles); err != nil {
			return
		}
		for _, c := range candles {
			ts := time.UnixMilli(c.OpenTime).Format("2006-01-02 15:04")
			closed := ""
			if c.Closed {
				closed = " [CLOSED]"
			}
			fmt.Printf("%s  %-20s  %s  O=%-12s H=%-12s L=%-12s C=%-12s V=%-12s%s\n",
				ts, c.Symbol, c.Interval, c.Open, c.High, c.Low, c.Close, c.Volume, closed)
		}
	}
}

func bboPrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var tickers []ws.BookTicker
		if err := json.Unmarshal(p.Data, &tickers); err != nil {
			return
		}
		for _, t := range tickers {
			fmt.Printf("%-20s  bid=%-14s(%s)  ask=%-14s(%s)\n",
				t.Symbol, t.BidPrice, t.BidQty, t.AskPrice, t.AskQty)
		}
	}
}

func markPricePrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var prices []ws.MarkPrice
		if err := json.Unmarshal(p.Data, &prices); err != nil {
			return
		}
		for _, m := range prices {
			fmt.Printf("%-20s  mark=%-14s  index=%-14s  funding=%-10s  OI=%s\n",
				m.Symbol, m.MarkPx, m.IndexPx, m.FundingRate, m.OpenInterest)
		}
	}
}

func orderUpdatesPrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var upd ws.AccountOrderUpdate
		if err := json.Unmarshal(p.Data, &upd); err != nil {
			return
		}
		ts := time.UnixMilli(upd.EventTime).Format("15:04:05.000")
		fmt.Printf("%s  %-20s  oid=%-12d  %s/%s  status=%-18s  price=%-14s  qty=%-12s  filled=%-12s  exec=%s\n",
			ts, upd.Symbol, upd.OrderID, upd.Side, upd.OrderType,
			upd.Status, upd.Price, upd.OrigQty, upd.FilledQty, upd.ExecType)
	}
}

func fillsPrinter(f outputFormat) func(ws.Push) {
	return func(p ws.Push) {
		if f == formatJSON {
			printJSON(p)
			return
		}
		var fill ws.AccountTrade
		if err := json.Unmarshal(p.Data, &fill); err != nil {
			return
		}
		ts := time.UnixMilli(fill.TradeTime).Format("15:04:05.000")
		maker := "taker"
		if fill.IsMaker {
			maker = "maker"
		}
		fmt.Printf("%s  %-20s  %s  price=%-14s  qty=%-12s  fee=%-10s  %s\n",
			ts, fill.Symbol, fill.Side, fill.Price, fill.Quantity, fill.Fee, maker)
	}
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	g := &globalFlags{}

	root := &cobra.Command{
		Use:   "sodex",
		Short: "CLI for the Sodex decentralized exchange",
		Long: `sodex — command-line interface for the Sodex DEX

Supports both the Spark (spot) and Bolt (perpetuals) trading engines.
All trading commands require a private key. Market data commands are public.

Use --testnet for testnet or set SODEX_TESTNET=1.
Configure API key auth with --api-key and --private-key (API key's private key).
Set --account-id globally or via SODEX_ACCOUNT_ID for trading commands.

Run 'sodex agent-help' for a comprehensive guide for AI agent usage.`,
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&g.baseURL, "base-url", "",
		"API base URL (default: mainnet) [$SODEX_BASE_URL]")
	root.PersistentFlags().StringVar(&g.privateKey, "private-key", "",
		"Private key hex (0x-prefixed or raw) [$SODEX_PRIVATE_KEY]")
	root.PersistentFlags().StringVar(&g.address, "address", "",
		"Ethereum address (derived from private key if not set) [$SODEX_ADDRESS]")
	root.PersistentFlags().Uint64Var(&g.chainID, "chain-id", client.DefaultChainID,
		"Chain ID [$SODEX_CHAIN_ID]")
	root.PersistentFlags().StringVar(&g.format, "format", "pretty",
		"Default output format: pretty|table|json [$SODEX_FORMAT]")
	root.PersistentFlags().BoolVar(&g.testnet, "testnet", false,
		"Use testnet (overrides base-url and chain-id) [$SODEX_TESTNET]")
	root.PersistentFlags().StringVar(&g.apiKey, "api-key", "",
		"API key name for authenticated requests [$SODEX_API_KEY]")
	root.PersistentFlags().Uint64Var(&g.accountID, "account-id", 0,
		"Default account ID for trading commands [$SODEX_ACCOUNT_ID]")

	root.AddCommand(
		newAccountIDCmd(g),
		newMarketsCmd(g),
		newTickersCmd(g),
		newOrderBookCmd(g),
		newBalanceCmd(g),
		newOrdersCmd(g),
		newPositionsCmd(g),
		newLeverageCmd(g),
		newTransferCmd(g),
		newSubscribeCmd(g),
		newAgentHelpCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// ── Agent help text ───────────────────────────────────────────────────────────

const agentHelpText = `SODEX CLI — AI Agent Guide
==========================

sodex is a command-line interface for the Sodex decentralized exchange.
It supports the Spark (spot) and Bolt (perpetuals) trading engines.

AUTHENTICATION
--------------
Trading commands require a private key. Set it via:
  • Environment variable: SODEX_PRIVATE_KEY=0x<hex>
  • Flag: --private-key 0x<hex>

For API key authentication (recommended for production):
  1. Generate an API key pair and register it with the exchange
  2. Use the API key's private key for signing: --private-key 0x<api-key-private>
  3. Set the API key name: --api-key <name> or SODEX_API_KEY=<name>

The wallet address is derived automatically from the private key.
You can also pass it explicitly: --address 0x<addr> or SODEX_ADDRESS=0x<addr>

WARNING: Never log or commit private keys. In production, prefer environment
variables loaded from a secrets manager.

CONFIGURATION
-------------
  SODEX_BASE_URL     API base URL (default: https://mainnet-gw.sodex.dev)
  SODEX_PRIVATE_KEY  Private key hex (enables trading)
  SODEX_API_KEY      API key name (for API key authentication)
  SODEX_ACCOUNT_ID   Default account ID for trading commands
  SODEX_ADDRESS      Wallet address (overrides key derivation for queries)
  SODEX_CHAIN_ID     Chain ID (default: 286623)
  SODEX_TESTNET      Set to "1" or "true" for testnet mode
  SODEX_FORMAT       Default output format: pretty|table|json

NETWORK SELECTION
-----------------
  --testnet          Use testnet (base-url=https://testnet-gw.sodex.dev, chain-id=138565)
  Default is mainnet (base-url=https://mainnet-gw.sodex.dev, chain-id=286623).
  Explicit --base-url and --chain-id override testnet defaults.

OUTPUT FORMATS
--------------
  pretty   Human-readable text (default)
  table    Tab-aligned columns (good for terminal parsing)
  json     Machine-readable JSON (ideal for programmatic use)

Use --format json for structured output suitable for further processing.

ENGINES
-------
  spot    Spark engine — spot trading (default)
  perps   Bolt engine  — perpetuals trading

Most commands accept an optional first argument to select the engine:
  sodex markets spot
  sodex markets perps

COMMANDS
--------

  account-id [ADDRESS] [--format]
    Look up the account ID and user ID for a wallet address.
    ADDRESS can be positional, --address, or derived from --private-key.
    Returns: address, accountID, userID.

  markets [spot|perps] [--format]
    List all available trading pairs.
    Key fields: symbolID (required for cancel), symbol, baseAsset, quoteAsset, status

  tickers [spot|perps] [--format]
    Show 24-hour stats: lastPrice, bidPrice, askPrice, priceChangePercent, volume.
    For perps: also markPrice, indexPrice, fundingRate, openInterest.

  orderbook [spot|perps] SYMBOL [--depth N] [--format]
    Show the live order book for a trading pair.
    Default depth: 20 levels. SYMBOL is the human-readable name (e.g. BTC-USDT).

  balance [spot|perps] [ADDRESS] [--format]
    Show asset balances. ADDRESS can be positional, --address, or derived from key.
    Fields: asset, free, locked, accountID.

  orders list [spot|perps] [ADDRESS] [--format]
    List open orders for an address.
    Fields: orderID, clOrdID, symbol, side, type, price, quantity, filledQuantity, status.

  orders place [spot|perps] --symbol SYMBOL --side buy|sell \
               [--type limit|market] [--price P] --qty Q [--tif gtc|ioc|fok|gtx] \
               [--reduce-only] [--format]
    Place a new order. The symbol name is resolved to a symbolID automatically.
    --type defaults to limit. --tif defaults to gtc.
    --reduce-only is only meaningful for perps.
    --account-id is resolved from the global flag or SODEX_ACCOUNT_ID.
    Returns: clOrdID, orderID, status for each submitted order.

  orders cancel [spot|perps] --symbol-id N \
                [--order-id N | --cl-ord-id S] [--format]
    Cancel a resting order. Use --symbol-id (numeric) from ` + "`" + `sodex markets` + "`" + `.
    Provide either --order-id or --cl-ord-id (not both).
    --account-id is resolved from the global flag or SODEX_ACCOUNT_ID.

  positions [ADDRESS] [--format]
    Show open perpetuals positions.
    Fields: symbol, positionSide, quantity, entryPrice, markPrice, liquidationPrice,
            unrealizedPnl, leverage, marginMode, margin.

  leverage SYMBOL VALUE [--mode isolated|cross]
    Update leverage for a perps position. VALUE is a positive integer (e.g. 10).
    --mode defaults to isolated. --account-id from global flag or SODEX_ACCOUNT_ID.

  transfer --from ACCOUNT_ID --to ACCOUNT_ID --coin COIN_ID --amount AMOUNT [--type N]
    Transfer assets between accounts. Requires private key.
    --type: 0=EVM deposit, 1=perps deposit, 2=EVM withdraw, 3=perps withdraw,
            4=internal, 5=spot withdraw, 6=spot deposit.

  agent-help
    Print this guide.

COMMON WORKFLOWS
----------------

1. Query all perps markets and find the symbolID for ETH-USDT:
     sodex markets perps --format json | jq '.[] | select(.symbol=="ETH-USDT")'

2. Check current ticker for BTC-USDT on spot:
     sodex tickers spot --format json | jq '.[] | select(.symbol=="BTC-USDT")'

3. Get your spot balances:
     export SODEX_PRIVATE_KEY=0x<key>
     sodex balance spot

4. Configure for testnet with API key:
     export SODEX_PRIVATE_KEY=0x<api-key-private>
     export SODEX_API_KEY=mykey
     export SODEX_ACCOUNT_ID=1001
     sodex --testnet balance spot

5. Place a spot limit buy for 0.01 BTC at $95,000:
     sodex orders place spot \
       --symbol BTC-USDT --side buy --price 95000 --qty 0.01

6. Place a perps market sell (close position) for 0.1 BTC:
     sodex orders place perps \
       --symbol BTC-USDT --side sell --type market --qty 0.1 --reduce-only

7. List open perps orders:
     sodex orders list perps --format table

8. Cancel a spot order by client order ID:
     sodex orders cancel spot --symbol-id 42 \
       --cl-ord-id sodex-cli-1711900800000

9. Check perps positions:
     sodex positions --format json

10. Set 10x leverage on ETH-USDT perps:
     sodex leverage ETH-USDT 10

11. Transfer 100 USDT internally between sub-accounts:
     sodex transfer --from 1001 --to 1002 --coin 1 --amount 100

ERROR MESSAGES
--------------
  "not authenticated — set Config.PrivateKey or SODEX_PRIVATE_KEY"
    → Set SODEX_PRIVATE_KEY environment variable.

  "sodex API error (code N): <message>"
    → Application-level error from the exchange. Check the message for details.
    Common codes: 401 = auth failure, 400 = bad request, 429 = rate limit exceeded.

  "symbol X not found in Y markets"
    → Run ` + "`" + `sodex markets <engine>` + "`" + ` to see valid symbol names.

  "invalid private key"
    → Ensure the key is 32 bytes (64 hex chars), optionally 0x-prefixed.

NONCE MANAGEMENT
----------------
The CLI automatically generates nonces as millisecond timestamps.
The exchange accepts nonces within (now - 2 days, now + 1 day).
No manual nonce management is needed when using the CLI.

SIGNING
-------
All trading commands use EIP-712 signatures with engine-specific domains:
  Spot (Spark):  domain name="spot",    chainId=286623
  Perps (Bolt):  domain name="futures", chainId=286623

Signatures are automatically computed and attached to the X-API-Sign header.
`
