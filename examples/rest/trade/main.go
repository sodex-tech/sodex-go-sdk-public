// Command trade places and then cancels a single perps limit order on testnet.
//
// Usage:
//
//	export SODEX_PRIVATE_KEY=<hex, no 0x>
//	export SODEX_ACCOUNT_ID=<your account ID>
//	go run ./examples/rest/trade
//
// The order is placed far below market (GTC @ $1000) so it rests safely in the
// book; the example then cancels it. No fill should occur.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"

	"github.com/sodex-tech/sodex-go-sdk-public/client"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ptypes "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
)

func main() {
	pkHex := os.Getenv("SODEX_PRIVATE_KEY")
	if pkHex == "" {
		log.Fatal("SODEX_PRIVATE_KEY is required (hex, no 0x prefix)")
	}
	accountID, err := strconv.ParseUint(os.Getenv("SODEX_ACCOUNT_ID"), 10, 64)
	if err != nil {
		log.Fatalf("SODEX_ACCOUNT_ID must be a uint64: %v", err)
	}

	pk, err := crypto.HexToECDSA(pkHex)
	if err != nil {
		log.Fatalf("parse private key: %v", err)
	}

	c := client.New(client.Config{
		BaseURL:    client.TestnetBaseURL,
		ChainID:    client.TestnetChainID,
		PrivateKey: pk,
	})
	fmt.Printf("Signer address: %s\n\n", c.Address())

	ctx := context.Background()

	// ── 1. Discover the symbol ID for BTC-USD ────────────────────────────────
	symbols, err := c.PerpsSymbols(ctx)
	if err != nil {
		log.Fatalf("PerpsSymbols: %v", err)
	}
	var btc *client.Symbol
	for i, s := range symbols {
		if s.Symbol == "BTC-USD" {
			btc = &symbols[i]
			break
		}
	}
	if btc == nil {
		log.Fatal("BTC-USD not found in PerpsSymbols")
	}
	fmt.Printf("BTC-USD symbolID=%d tickSize=%s stepSize=%s\n\n", btc.SymbolID, btc.TickSize, btc.StepSize)

	// ── 2. Place a GTC limit buy far below market (won't fill) ───────────────
	clOrdID := fmt.Sprintf("demo-%d", time.Now().UnixMilli())
	placed, err := c.PlacePerpsLimitOrder(
		ctx,
		accountID,
		btc.SymbolID,
		clOrdID,
		enums.OrderSideBuy,
		enums.PositionSideLong,
		enums.TimeInForceGTC,
		decimal.NewFromInt(1000), // intentionally far below market
		decimal.NewFromFloat(0.001),
		false, // reduceOnly
	)
	if err != nil {
		log.Fatalf("PlacePerpsLimitOrder: %v", err)
	}
	if len(placed) == 0 {
		log.Fatal("no results returned from place")
	}
	order := placed[0]
	fmt.Printf("Placed: orderID=%d clOrdID=%s status=%s\n\n", order.OrderID, order.ClOrdID, order.Status)

	// ── 3. Cancel it ─────────────────────────────────────────────────────────
	orderID := order.OrderID
	cancelled, err := c.CancelPerpsOrders(ctx, &ptypes.CancelOrderRequest{
		AccountID: accountID,
		Cancels: []*ptypes.CancelOrder{
			{SymbolID: btc.SymbolID, OrderID: &orderID},
		},
	})
	if err != nil {
		log.Fatalf("CancelPerpsOrders: %v", err)
	}
	for _, r := range cancelled {
		fmt.Printf("Cancelled: clOrdID=%s status=%s\n", r.ClOrdID, r.Status)
	}
}
