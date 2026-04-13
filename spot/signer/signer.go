// Package signer provides the Signer for the Spark spot engine.
//
// It wraps the engine-agnostic EVMSigner with the Spark EIP-712 domain so that
// all signatures produced here are cryptographically bound to the spot engine.
// Attempting to replay a spot signature on the perps engine will fail verification
// because the domain separator encodes the engine name.
//
// # Usage
//
//	privateKey, _ := crypto.HexToECDSA("...")
//	s := signer.NewSigner(chainID, privateKey)
//
//	req := &ctypes.TransferAssetRequest{ ... }
//	sig, err := s.SignTransferAssetRequest(req, nonce)
//	// sig is a 66-byte wire-format signature ready to attach to the HTTP request.
package signer

import (
	"crypto/ecdsa"

	"github.com/sodex-tech/sodex-go-sdk-public/common/signer"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
	types "github.com/sodex-tech/sodex-go-sdk-public/spot/types"
)

// Signer signs spot-engine action requests using the Spark EIP-712 domain.
//
// A single Signer instance can be reused for the lifetime of a session.
// It holds no mutable state beyond the private key and domain, both of which
// are set at construction.
type Signer struct {
	signer     *signer.EVMSigner
	privateKey *ecdsa.PrivateKey
}

// NewSigner creates a Signer for the Spark spot engine on the given chain ID.
//
// chainID must match the chain ID used by the exchange server; mismatches will
// cause all signature verifications to fail.
func NewSigner(chainID uint64, privateKey *ecdsa.PrivateKey) *Signer {
	domain := ctypes.NewEIP712Domain(ctypes.SpotDomainName, chainID)
	return &Signer{
		signer:     signer.NewEVMSigner(&domain),
		privateKey: privateKey,
	}
}

// ── Common Actions ────────────────────────────────────────────────────────────
//
// The following actions are available on both the spot and perps engines.
// Their signatures are not interchangeable because each engine uses a different
// EIP-712 domain (SpotDomainName vs. PerpsDomainName).

// SignTransferAssetRequest signs an inter-account asset transfer request.
// The nonce must be the caller's next valid nonce for the spot engine.
func (s *Signer) SignTransferAssetRequest(request *ctypes.TransferAssetRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignReplaceOrderRequest signs a batch order-replace request.
// The nonce must be the caller's next valid nonce for the spot engine.
func (s *Signer) SignReplaceOrderRequest(request *ctypes.ReplaceOrderRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignScheduleCancelRequest signs a scheduled mass-cancel request.
// The nonce must be the caller's next valid nonce for the spot engine.
func (s *Signer) SignScheduleCancelRequest(request *ctypes.ScheduleCancelRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// ── Spot-Only Actions ─────────────────────────────────────────────────────────
//
// The following actions are exclusive to the Spark spot engine.

// SignBatchNewOrderRequest signs a batch of new-order placements in a single request.
// The nonce must be the caller's next valid nonce for the spot engine.
func (s *Signer) SignBatchNewOrderRequest(request *types.BatchNewOrderRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignBatchCancelOrderRequest signs a batch of order cancellations in a single request.
// The nonce must be the caller's next valid nonce for the spot engine.
func (s *Signer) SignBatchCancelOrderRequest(request *types.BatchCancelOrderRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}
