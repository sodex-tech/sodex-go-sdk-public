// Package types provides the core data structures and cryptographic primitives
// used to sign and verify exchange actions on the Sodex platform.
//
// # Signing Pipeline
//
// Every authenticated request follows the same two-level EIP-712 signing pipeline:
//
//	┌─────────────────────────────────────────────────────────────┐
//	│  1. Build ActionPayload{type, params}                       │
//	│     └─▶ JSON-encode ──▶ keccak256 ──▶ payloadHash          │
//	│                                                             │
//	│  2. Build ExchangeAction{payloadHash, nonce}                │
//	│     └─▶ StructHash ──▶ EIP-712 Hash(domain) ──▶ digest     │
//	│                                                             │
//	│  3. ECDSA-sign digest ──▶ [type_byte | 65-byte sig]        │
//	└─────────────────────────────────────────────────────────────┘
//
// The engine-specific domain (Spark for spot, Bolt for perps) prevents a signature
// produced for one engine from being replayed on the other. The nonce prevents
// replay within the same engine session.
//
// # Wire Format
//
// Signatures sent to the exchange are 66 bytes:
//
//	byte[0]    – SignatureType (always SignatureTypeEIP712 = 1)
//	byte[1:66] – 65-byte ECDSA signature (r ‖ s ‖ v)
package types

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
)

// Sentinel errors returned by signature recovery functions.
var (
	// ErrInvalidSignatureLength is returned when the signature slice is not the
	// expected 65 bytes (raw ECDSA) or 66 bytes (wire format with type prefix).
	ErrInvalidSignatureLength = errors.New("invalid signature length")

	// ErrInvalidSignatureType is returned when the leading type byte of a wire-format
	// signature does not match the expected SignatureType for the operation.
	ErrInvalidSignatureType = errors.New("invalid signature type")

	// ErrInvalidPublicKey is returned when signature recovery yields the zero address,
	// indicating a malformed or incorrect signature.
	ErrInvalidPublicKey = errors.New("invalid public key")
)

// Domain name constants used as the "name" field of the EIP-712 domain separator.
// Each engine operates under its own domain so that signatures are not portable
// across engines, preventing cross-engine replay attacks.
var (
	// SpotDomainName is the EIP-712 domain name for the Spark spot engine.
	SpotDomainName = "spot"

	// PerpsDomainName is the EIP-712 domain name for the Bolt perpetuals engine.
	PerpsDomainName = "futures"
)

// EIP712Domain holds the parameters that uniquely identify a signing domain.
//
// The domain separator is derived from these fields via keccak256 and mixed into
// every signature, binding the signature to a specific application and chain.
// Changing any field produces a completely different set of valid signatures.
//
// EIP-712 type string:
//
//	EIP712Domain(string name, string version, uint256 chainId, address verifyingContract)
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract common.Address

	// separator caches the computed domain separator to avoid redundant hashing
	// on repeated calls. It is populated lazily by DomainSeparator.
	separator *common.Hash
}

// NewEIP712Domain constructs an EIP712Domain with the given name and chain ID.
// The version is always "1" and the verifying contract is the zero address,
// matching the on-chain verifier configuration.
func NewEIP712Domain(name string, chainID uint64) EIP712Domain {
	return EIP712Domain{
		Name:              name,
		Version:           "1",
		ChainID:           big.NewInt(int64(chainID)),
		VerifyingContract: common.Address{}, // Zero address
	}
}

// DefaultSparkDomain returns the production EIP-712 domain for the Spark spot engine
// on chain ID 286623.
func DefaultSparkDomain() EIP712Domain {
	return NewEIP712Domain(SpotDomainName, 286623)
}

// DefaultBoltDomain returns the production EIP-712 domain for the Bolt perpetuals engine
// on chain ID 286623.
func DefaultBoltDomain() EIP712Domain {
	return NewEIP712Domain(PerpsDomainName, 286623)
}

// DomainSeparator computes and caches the EIP-712 domain separator hash.
//
// The separator is defined as:
//
//	keccak256(
//	    keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
//	    keccak256(name),
//	    keccak256(version),
//	    uint256(chainId),       // left-padded to 32 bytes
//	    address(verifyingContract), // left-padded to 32 bytes
//	)
//
// The result is cached after the first call; subsequent calls return the cached value
// without recomputing.
func (d *EIP712Domain) DomainSeparator() common.Hash {
	if d.separator != nil {
		return *d.separator
	}

	hash := crypto.Keccak256Hash(
		crypto.Keccak256(
			[]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
		),
		crypto.Keccak256([]byte(d.Name)),
		crypto.Keccak256([]byte(d.Version)),
		math.U256Bytes(d.ChainID),
		common.LeftPadBytes(d.VerifyingContract.Bytes(), 32),
	)
	d.separator = &hash
	return hash
}

// ExchangeAction is the EIP-712 typed message that the user ultimately signs.
//
// Rather than defining a unique EIP-712 type for every possible action, the exchange
// uses a single generic type whose payload field is a keccak256 hash of the
// action-specific data. This design:
//
//   - Keeps the on-chain and off-chain type registries simple and stable.
//   - Allows arbitrary action schemas without updating the signing infrastructure.
//   - Binds the signature to both the action content (via payloadHash) and the
//     session (via nonce), preventing both content forgery and replay attacks.
//
// EIP-712 type string:
//
//	ExchangeAction(bytes32 payloadHash, uint64 nonce)
type ExchangeAction struct {
	// PayloadHash is the keccak256 hash of ActionPayload{type, params} encoded as JSON.
	// It commits the signature to the exact action type and all its parameters.
	PayloadHash common.Hash

	// Nonce is a monotonically increasing counter that prevents replay attacks.
	// The exchange rejects any request whose nonce has already been consumed.
	Nonce uint64
}

// ExchangeActionTypeHash is the EIP-712 type hash for ExchangeAction.
// It is computed once at package initialisation from the canonical type string.
var ExchangeActionTypeHash = crypto.Keccak256Hash([]byte("ExchangeAction(bytes32 payloadHash,uint64 nonce)"))

// StructHash computes the EIP-712 struct hash for ExchangeAction.
//
//	keccak256(typeHash ‖ payloadHash ‖ nonce)
//
// The nonce is ABI-encoded as a big-endian uint256 (32 bytes, right-aligned).
func (ea *ExchangeAction) StructHash() common.Hash {
	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes[24:], ea.Nonce) // uint64 occupies the last 8 bytes of the 32-byte word

	return crypto.Keccak256Hash(
		ExchangeActionTypeHash.Bytes(),
		ea.PayloadHash.Bytes(),
		nonceBytes,
	)
}

// Hash computes the final EIP-712 digest that is passed to the ECDSA signer.
//
// Following EIP-191 and EIP-712, the digest is:
//
//	keccak256(0x19 ‖ 0x01 ‖ domainSeparator ‖ structHash)
//
// The 0x19 0x01 prefix distinguishes this from a plain transaction hash and
// prevents the digest from being mistaken for a valid RLP-encoded transaction.
func (ea *ExchangeAction) Hash(domain *EIP712Domain) common.Hash {
	domainSeparator := domain.DomainSeparator()
	structHash := ea.StructHash()

	return crypto.Keccak256Hash(
		[]byte{0x19, 0x01},
		domainSeparator.Bytes(),
		structHash.Bytes(),
	)
}

// RecoverExchangeActionSigner recovers the Ethereum address that produced the given
// EIP-712 signature for an ExchangeAction.
//
// The signature must be a raw 65-byte ECDSA signature (r ‖ s ‖ v) — i.e. the
// wire-format type prefix byte must be stripped before calling this function.
//
// Returns ErrInvalidSignatureLength if len(signature) != 65,
// or ErrInvalidPublicKey if recovery yields the zero address.
func RecoverExchangeActionSigner(payloadHash common.Hash, nonce uint64, domain *EIP712Domain, signature []byte) (common.Address, error) {
	if len(signature) != 65 {
		return common.Address{}, ErrInvalidSignatureLength
	}

	ea := &ExchangeAction{
		PayloadHash: payloadHash,
		Nonce:       nonce,
	}

	hash := ea.Hash(domain)

	pubKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*pubKey)
	if address == (common.Address{}) {
		return common.Address{}, ErrInvalidPublicKey
	}
	return address, nil
}
