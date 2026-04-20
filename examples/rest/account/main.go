// Command account queries balances, open orders, and positions for a user.
//
// Usage:
//
//	# Read-only — address from env; no signing required.
//	export SODEX_ADDRESS=0x…
//	go run ./examples/rest/account
//
//	# Or: sign with a private key and auto-derive the address.
//	export SODEX_PRIVATE_KEY=<hex, no 0x>
//	go run ./examples/rest/account
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sodex-tech/sodex-go-sdk-public/client"
)

func main() {
	cfg := client.Config{BaseURL: client.TestnetBaseURL, ChainID: client.TestnetChainID}

	var address string
	if pkHex := os.Getenv("SODEX_PRIVATE_KEY"); pkHex != "" {
		pk, err := crypto.HexToECDSA(pkHex)
		if err != nil {
			log.Fatalf("parse private key: %v", err)
		}
		cfg.PrivateKey = pk
		address = crypto.PubkeyToAddress(pk.PublicKey).Hex()
	} else {
		address = os.Getenv("SODEX_ADDRESS")
	}
	if address == "" {
		log.Fatal("either SODEX_PRIVATE_KEY or SODEX_ADDRESS must be set")
	}

	c := client.New(cfg)
	ctx := context.Background()

	fmt.Printf("Querying %s on testnet\n\n", address)

	// ── Perps ────────────────────────────────────────────────────────────────
	fmt.Println("── Perps ──────────────────────────────────────────────────────")

	balances, err := c.PerpsBalances(ctx, address)
	if err != nil {
		log.Fatalf("PerpsBalances: %v", err)
	}
	fmt.Printf("Balances (%d):\n", len(balances))
	for _, b := range balances {
		fmt.Printf("  %-8s total=%-20s locked=%s\n", b.Coin, b.Total, b.Locked)
	}

	positions, err := c.PerpsPositions(ctx, address)
	if err != nil {
		log.Fatalf("PerpsPositions: %v", err)
	}
	fmt.Printf("\nOpen positions (%d):\n", len(positions))
	for _, p := range positions {
		fmt.Printf("  %-12s side=%-5s qty=%-12s entry=%-12s mark=%-12s uPnL=%s\n",
			p.Symbol, p.PositionSide, p.Quantity, p.EntryPrice, p.MarkPrice, p.UnrealizedPnl)
	}

	orders, err := c.PerpsOrders(ctx, address)
	if err != nil {
		log.Fatalf("PerpsOrders: %v", err)
	}
	fmt.Printf("\nOpen orders (%d):\n", len(orders))
	for _, o := range orders {
		fmt.Printf("  [%d] %-12s %-4s %-6s qty=%-10s price=%-10s status=%s\n",
			o.OrderID, o.Symbol, o.Side, o.Type, o.OrigQty, o.Price, o.Status)
	}

	// ── Spot ─────────────────────────────────────────────────────────────────
	fmt.Println("\n── Spot ───────────────────────────────────────────────────────")

	info, err := c.SpotAccountInfo(ctx, address)
	if err != nil {
		log.Fatalf("SpotAccountInfo: %v", err)
	}
	fmt.Printf("Spot account: aid=%d uid=%d\n", info.AccountID, info.UserID)

	spotBalances, err := c.SpotBalances(ctx, address)
	if err != nil {
		log.Fatalf("SpotBalances: %v", err)
	}
	fmt.Printf("Balances (%d):\n", len(spotBalances))
	for _, b := range spotBalances {
		fmt.Printf("  %-8s total=%-20s locked=%s\n", b.Coin, b.Total, b.Locked)
	}
}
