// Command approve-builder-fee approves a builder's maximum fee rate on both
// Spot and Perps with one master-wallet signature.
//
// Usage:
//
//	export SODEX_PRIVATE_KEY=<hex, no 0x>
//	export SODEX_ACCOUNT_ID=<primary account ID>
//	export SODEX_BUILDER_ID=<builder account ID>
//	export SODEX_BUILDER_FEE_RATE=<maximum fee rate>
//	go run ./examples/rest/approve-builder-fee
package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sodex-tech/sodex-go-sdk-public/client"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
)

func requiredUint64(name string) uint64 {
	value, err := strconv.ParseUint(os.Getenv(name), 10, 64)
	if err != nil {
		log.Fatalf("%s must be a uint64: %v", name, err)
	}
	return value
}

func main() {
	privateKey, err := crypto.HexToECDSA(os.Getenv("SODEX_PRIVATE_KEY"))
	if err != nil {
		log.Fatalf("SODEX_PRIVATE_KEY must be valid hex without a 0x prefix: %v", err)
	}

	c := client.New(client.Config{
		BaseURL:    client.TestnetBaseURL,
		ChainID:    client.TestnetChainID,
		PrivateKey: privateKey,
	})
	request := &ctypes.ApproveBuilderFeeRequest{
		AccountID:  requiredUint64("SODEX_ACCOUNT_ID"),
		BuilderID:  requiredUint64("SODEX_BUILDER_ID"),
		MaxFeeRate: requiredUint64("SODEX_BUILDER_FEE_RATE"),
	}
	if err := c.ApproveBuilderFee(context.Background(), c.Address(), request); err != nil {
		log.Fatalf("ApproveBuilderFee: %v", err)
	}
	log.Printf(
		"approved builder %d with max fee rate %d for account %d on Spot and Perps",
		request.BuilderID,
		request.MaxFeeRate,
		request.AccountID,
	)
}
