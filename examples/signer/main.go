// Package main demonstrates how to use the sodex-go-sdk signing packages
// to sign various trading requests for both the spot and perps engines.
//
// Usage:
//
//	go run ./examples/signer
package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"

	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
	perpsSigner "github.com/sodex-tech/sodex-go-sdk-public/perps/signer"
	perpsTypes "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
	spotSigner "github.com/sodex-tech/sodex-go-sdk-public/spot/signer"
	spotTypes "github.com/sodex-tech/sodex-go-sdk-public/spot/types"
)

// chainID is the chain ID for the Sodex exchange.
const chainID = uint64(286623)

func main() {
	// ── 1. Load private key ──────────────────────────────────────────────
	// Replace with your own private key (without 0x prefix).
	privateKey, err := crypto.HexToECDSA("0123456789012345678901234567890123456789012345678901234567890123")
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("Signer address: %s\n\n", address.Hex())

	// nonce should be fetched from the exchange API before each request.
	var nonce uint64 = 1

	// ── 2. Perps engine examples ─────────────────────────────────────────
	ps := perpsSigner.NewSigner(chainID, privateKey)

	// 2a. Sign a new perps order (limit buy)
	price := decimal.NewFromFloat(50000.0)
	qty := decimal.NewFromFloat(0.1)
	newOrderReq := &perpsTypes.NewOrderRequest{
		AccountID: 1001,
		SymbolID:  1,
		Orders: []*perpsTypes.RawOrder{
			{
				ClOrdID:      "my-order-001",
				Modifier:     enums.OrderModifierNormal,
				Side:         enums.OrderSideBuy,
				Type:         enums.OrderTypeLimit,
				TimeInForce:  enums.TimeInForceGTC,
				Price:        &price,
				Quantity:     &qty,
				PositionSide: enums.PositionSideLong,
			},
		},
	}
	sig, err := ps.SignNewOrderRequest(newOrderReq, nonce)
	if err != nil {
		log.Fatalf("SignNewOrderRequest failed: %v", err)
	}
	fmt.Printf("[Perps] NewOrder signature (%d bytes): %s\n", len(sig), hex.EncodeToString(sig))
	nonce++

	// 2b. Sign a cancel order request
	orderID := uint64(12345)
	cancelReq := &perpsTypes.CancelOrderRequest{
		AccountID: 1001,
		Cancels: []*perpsTypes.CancelOrder{
			{SymbolID: 1, OrderID: &orderID},
		},
	}
	sig, err = ps.SignCancelOrderRequest(cancelReq, nonce)
	if err != nil {
		log.Fatalf("SignCancelOrderRequest failed: %v", err)
	}
	fmt.Printf("[Perps] CancelOrder signature (%d bytes): %s\n", len(sig), hex.EncodeToString(sig))
	nonce++

	// 2c. Sign a schedule cancel (cancel-all) request
	scheduleCancelReq := &ctypes.ScheduleCancelRequest{
		AccountID: 1001,
	}
	sig, err = ps.SignScheduleCancelRequest(scheduleCancelReq, nonce)
	if err != nil {
		log.Fatalf("SignScheduleCancelRequest failed: %v", err)
	}
	fmt.Printf("[Perps] ScheduleCancel signature (%d bytes): %s\n", len(sig), hex.EncodeToString(sig))
	nonce++

	// ── 3. Spot engine examples ──────────────────────────────────────────
	ss := spotSigner.NewSigner(chainID, privateKey)

	// 3a. Sign a batch new order request
	limitPrice := decimal.NewFromFloat(3500.0)
	orderQty := decimal.NewFromFloat(1.5)
	batchNewOrderReq := &spotTypes.BatchNewOrderRequest{
		AccountID: 2001,
		Orders: []*spotTypes.BatchNewOrderItem{
			{
				SymbolID:    10,
				ClOrdID:     "spot-order-001",
				Side:        enums.OrderSideBuy,
				Type:        enums.OrderTypeLimit,
				TimeInForce: enums.TimeInForceGTC,
				Price:       &limitPrice,
				Quantity:    &orderQty,
			},
		},
	}
	sig, err = ss.SignBatchNewOrderRequest(batchNewOrderReq, nonce)
	if err != nil {
		log.Fatalf("SignBatchNewOrderRequest failed: %v", err)
	}
	fmt.Printf("[Spot]  BatchNewOrder signature (%d bytes): %s\n", len(sig), hex.EncodeToString(sig))
	nonce++

	// 3b. Sign a transfer asset request
	transferReq := &ctypes.TransferAssetRequest{
		ID:            1,
		FromAccountID: 2001,
		ToAccountID:   2002,
		CoinID:        1,
		Amount:        decimal.NewFromFloat(100.0),
		Type:          enums.TransferAssetTypeInternal,
	}
	sig, err = ss.SignTransferAssetRequest(transferReq, nonce)
	if err != nil {
		log.Fatalf("SignTransferAssetRequest failed: %v", err)
	}
	fmt.Printf("[Spot]  TransferAsset signature (%d bytes): %s\n", len(sig), hex.EncodeToString(sig))

	fmt.Println("\nAll signatures generated successfully!")
}
