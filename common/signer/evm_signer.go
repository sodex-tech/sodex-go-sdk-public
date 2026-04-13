// Package signer provides the EVMSigner, which implements the EIP-712 signing and
// verification logic shared by both the Spark (spot) and Bolt (perps) engines.
//
// All signing is delegated to the engine-specific Signer wrappers in the spot/signer
// and perps/signer packages; this package contains only the cryptographic core.
package signer

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	"github.com/sodex-tech/sodex-go-sdk-public/common/types"
)

// EVMSigner performs EIP-712 signing and signature verification using an
// engine-specific domain. It is stateless with respect to the private key;
// callers supply the key on each Sign call.
//
// One EVMSigner instance is created per engine (spot or perps). The domain
// bakes the engine name and chain ID into every signature, so a signature
// produced for the spot engine is cryptographically invalid on the perps engine
// and vice versa.
type EVMSigner struct {
	// domain is the EIP-712 domain for this engine instance.
	// It is set once at construction and never mutated.
	domain *types.EIP712Domain
}

// NewEVMSigner creates an EVMSigner bound to the provided EIP-712 domain.
// Typically called by the engine-specific signer constructors with either
// the Spark (spot) or Bolt (perps) domain.
func NewEVMSigner(domain *types.EIP712Domain) *EVMSigner {
	return &EVMSigner{domain: domain}
}

// RecoverPublicKeyFromRequest reconstructs the Ethereum address that signed a
// wire-format signature for the given action request.
//
// The wire-format signature is 66 bytes:
//
//	byte[0]    – must equal SignatureTypeEIP712 (value 1)
//	byte[1:66] – 65-byte raw ECDSA signature (r ‖ s ‖ v)
//
// The recovery process mirrors SignAction in reverse:
//
//  1. Wrap params in ActionPayload and hash to payloadHash.
//  2. Build ExchangeAction{payloadHash, nonce} and compute the EIP-712 digest.
//  3. Call crypto.SigToPub to recover the public key from the digest and signature.
//  4. Return the corresponding Ethereum address (as a 20-byte slice).
//
// Returns ErrInvalidSignatureLength if len(signature) != 66, or
// ErrInvalidSignatureType if the leading byte is not SignatureTypeEIP712.
func (s *EVMSigner) RecoverPublicKeyFromRequest(params types.ActionPayloadParams, nonce uint64, signature []byte) ([]byte, error) {
	if len(signature) != 66 {
		return nil, types.ErrInvalidSignatureLength
	}
	if signature[0] != byte(enums.SignatureTypeEIP712) {
		return nil, types.ErrInvalidSignatureType
	}

	ap := &types.ActionPayload{
		Type:   params.ActionName(),
		Params: params,
	}
	return s.recoverPublicKey(ap, nonce, signature)
}

// recoverPublicKey is the internal helper that executes the full EIP-712
// recovery path for a generic ExchangeAction.
//
//  1. Hash the ActionPayload to a 32-byte payloadHash.
//  2. Recover the signer address from the ExchangeAction EIP-712 digest.
//  3. Return the address bytes (note: the raw ECDSA signature starts at index 1,
//     skipping the type prefix byte).
func (s *EVMSigner) recoverPublicKey(ap *types.ActionPayload, nonce uint64, signature []byte) ([]byte, error) {
	payloadHash, err := ap.Hash()
	if err != nil {
		return nil, err
	}

	// signature[1:] strips the leading SignatureType byte, leaving the 65-byte ECDSA sig.
	address, err := types.RecoverExchangeActionSigner(payloadHash, nonce, s.domain, signature[1:])
	if err != nil {
		return nil, err
	}

	return address.Bytes(), nil
}

// SignAction signs an action request and returns a 66-byte wire-format signature.
//
// The signing pipeline:
//
//  1. Serialize params as ActionPayload{type, params} and hash to payloadHash.
//  2. Build ExchangeAction{payloadHash, nonce} and compute the EIP-712 digest:
//     keccak256(0x19 ‖ 0x01 ‖ domainSeparator ‖ structHash)
//  3. ECDSA-sign the digest with privateKey (produces 65 bytes).
//  4. Prepend the SignatureType byte (0x01) to form the 66-byte wire signature.
//
// The nonce must be the caller's current valid nonce; the exchange rejects
// requests with a stale or already-consumed nonce.
func (s *EVMSigner) SignAction(params types.ActionPayloadParams, nonce uint64, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	signature, err := s.signExchangeAction(params, nonce, privateKey)
	if err != nil {
		return nil, err
	}
	// Prepend the signature type byte so the server can route to the correct verifier.
	return append([]byte{byte(enums.SignatureTypeEIP712)}, signature...), nil
}

// signExchangeAction produces the raw 65-byte ECDSA signature for a generic action.
// It does not attach the type prefix; callers are responsible for that (see SignAction).
func (s *EVMSigner) signExchangeAction(params types.ActionPayloadParams, nonce uint64, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	ap := &types.ActionPayload{
		Type:   params.ActionName(),
		Params: params,
	}
	payloadHash, err := ap.Hash()
	if err != nil {
		return nil, err
	}

	action := &types.ExchangeAction{
		PayloadHash: payloadHash,
		Nonce:       nonce,
	}

	hash := action.Hash(s.domain)
	return crypto.Sign(hash.Bytes(), privateKey)
}
