// Package signer provides the Signer for the Bolt perpetuals engine.
//
// It wraps the engine-agnostic EVMSigner with the Bolt EIP-712 domain so that
// all signatures produced here are cryptographically bound to the perps engine.
// Attempting to replay a perps signature on the spot engine will fail verification
// because the domain separator encodes the engine name.
//
// # Usage
//
//	privateKey, _ := crypto.HexToECDSA("...")
//	s := signer.NewSigner(chainID, privateKey)
//
//	req := &types.NewOrderRequest{ ... }
//	sig, err := s.SignNewOrderRequest(req, nonce)
//	// sig is a 66-byte wire-format signature ready to attach to the HTTP request.
package signer

import (
	"crypto/ecdsa"

	"github.com/sodex-tech/sodex-go-sdk-public/common/signer"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
	types "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
)

// Signer signs perps-engine action requests using the Bolt EIP-712 domain.
//
// A single Signer instance can be reused for the lifetime of a session.
// It holds no mutable state beyond the private key and domain, both of which
// are set at construction.
type Signer struct {
	signer     *signer.EVMSigner
	privateKey *ecdsa.PrivateKey
}

// NewSigner creates a Signer for the Bolt perpetuals engine on the given chain ID.
//
// chainID must match the chain ID used by the exchange server; mismatches will
// cause all signature verifications to fail.
func NewSigner(chainID uint64, privateKey *ecdsa.PrivateKey) *Signer {
	domain := ctypes.NewEIP712Domain(ctypes.PerpsDomainName, chainID)
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
// The nonce must be the caller's next valid nonce for the perps engine.
func (s *Signer) SignTransferAssetRequest(request *ctypes.TransferAssetRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignReplaceOrderRequest signs a batch order-replace request.
// The nonce must be the caller's next valid nonce for the perps engine.
func (s *Signer) SignReplaceOrderRequest(request *ctypes.ReplaceOrderRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignScheduleCancelRequest signs a scheduled mass-cancel request.
// The nonce must be the caller's next valid nonce for the perps engine.
func (s *Signer) SignScheduleCancelRequest(request *ctypes.ScheduleCancelRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// ── Perps-Only Actions ────────────────────────────────────────────────────────
//
// The following actions are exclusive to the Bolt perpetuals engine.

// SignNewOrderRequest signs a new perpetuals order placement request.
// The nonce must be the caller's next valid nonce for the perps engine.
func (s *Signer) SignNewOrderRequest(request *types.NewOrderRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignCancelOrderRequest signs a single order-cancellation request.
// The nonce must be the caller's next valid nonce for the perps engine.
func (s *Signer) SignCancelOrderRequest(request *types.CancelOrderRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignUpdateLeverageRequest signs a position-leverage update request.
// The nonce must be the caller's next valid nonce for the perps engine.
func (s *Signer) SignUpdateLeverageRequest(request *types.UpdateLeverageRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}

// SignUpdateMarginRequest signs a position-margin adjustment request.
// The nonce must be the caller's next valid nonce for the perps engine.
func (s *Signer) SignUpdateMarginRequest(request *types.UpdateMarginRequest, nonce uint64) ([]byte, error) {
	return s.signer.SignAction(request, nonce, s.privateKey)
}
