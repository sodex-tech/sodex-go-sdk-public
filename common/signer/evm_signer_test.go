// Package signer_test verifies the correctness of the EIP-712 signing pipeline.
//
// Tests are organised around the properties that must hold for any production
// signing SDK:
//
//  1. Wire format     — every signature is exactly 66 bytes with the right type prefix.
//  2. Round-trip      — Sign followed by Recover returns the original signer address.
//  3. Cross-engine    — a spot signature cannot be recovered as valid on the perps domain.
//  4. Determinism     — identical inputs always produce the same signature (RFC 6979).
//  5. Sensitivity     — changing nonce, action type, or domain changes the signature.
//  6. Error handling  — malformed inputs return typed sentinel errors.
//  7. Domain caching  — DomainSeparator is idempotent and chain-ID-aware.
package signer_test

import (
	"bytes"
	"crypto/ecdsa"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	"github.com/sodex-tech/sodex-go-sdk-public/common/signer"
	"github.com/sodex-tech/sodex-go-sdk-public/common/types"
)

// ── Test fixtures ─────────────────────────────────────────────────────────────

// testPrivateKeyHex is a well-known deterministic test key taken from the
// standard Ethereum signing test vectors. Do NOT use with real funds.
const testPrivateKeyHex = "0123456789012345678901234567890123456789012345678901234567890123"

// testChainID is the production chain ID used throughout these tests.
const testChainID = uint64(286623)

// mustLoadKey parses testPrivateKeyHex and returns the private key together
// with its Ethereum address. It is a test helper that fails the test on error.
func mustLoadKey(t *testing.T) (*ecdsa.PrivateKey, [20]byte) {
	t.Helper()
	key, err := gethcrypto.HexToECDSA(testPrivateKeyHex)
	require.NoError(t, err, "failed to parse test private key")
	return key, gethcrypto.PubkeyToAddress(key.PublicKey)
}

// newSigner constructs an EVMSigner for the given EIP-712 domain name and the
// canonical test chain ID.
func newSigner(domainName string) *signer.EVMSigner {
	domain := types.NewEIP712Domain(domainName, testChainID)
	return signer.NewEVMSigner(&domain)
}

// ── 1. Wire format ────────────────────────────────────────────────────────────

// TestSignAction_WireFormat verifies the 66-byte wire format contract:
//
//	byte[0]    = SignatureTypeEIP712 (0x01)
//	byte[1:66] = 65-byte ECDSA signature (r ‖ s ‖ v)
func TestSignAction_WireFormat(t *testing.T) {
	key, _ := mustLoadKey(t)
	s := newSigner(types.SpotDomainName)

	sig, err := s.SignAction(&types.ScheduleCancelRequest{AccountID: 1001}, 1, key)

	require.NoError(t, err)
	assert.Len(t, sig, 66, "signature must be exactly 66 bytes")
	assert.Equal(t, byte(enums.SignatureTypeEIP712), sig[0],
		"leading byte must be SignatureTypeEIP712 (0x01)")
}

// ── 2. Round-trip ─────────────────────────────────────────────────────────────

// TestSignAndRecover_Spot signs a request under the Spark domain and asserts
// that RecoverPublicKeyFromRequest returns the expected signer address.
func TestSignAndRecover_Spot(t *testing.T) {
	key, wantAddr := mustLoadKey(t)
	s := newSigner(types.SpotDomainName)

	req := &types.ScheduleCancelRequest{AccountID: 1001}
	sig, err := s.SignAction(req, 1, key)
	require.NoError(t, err)

	got, err := s.RecoverPublicKeyFromRequest(req, 1, sig)
	require.NoError(t, err)
	assert.Equal(t, wantAddr[:], got, "recovered address must match signer")
}

// TestSignAndRecover_Perps signs a request under the Bolt domain and asserts
// that RecoverPublicKeyFromRequest returns the expected signer address.
func TestSignAndRecover_Perps(t *testing.T) {
	key, wantAddr := mustLoadKey(t)
	s := newSigner(types.PerpsDomainName)

	req := &types.ScheduleCancelRequest{AccountID: 2002}
	sig, err := s.SignAction(req, 7, key)
	require.NoError(t, err)

	got, err := s.RecoverPublicKeyFromRequest(req, 7, sig)
	require.NoError(t, err)
	assert.Equal(t, wantAddr[:], got, "recovered address must match signer")
}

// TestSignAndRecover_MultipleRequests verifies that round-trip recovery works
// for every concrete request type shared between both engines.
func TestSignAndRecover_MultipleRequests(t *testing.T) {
	key, wantAddr := mustLoadKey(t)
	s := newSigner(types.SpotDomainName)

	reqs := []types.ActionPayloadParams{
		&types.ScheduleCancelRequest{AccountID: 1001},
		&types.ScheduleCancelRequest{AccountID: 1001, ScheduledTimestamp: ptr(uint64(9999999))},
		&types.ReplaceOrderRequest{AccountID: 1001},
		&types.TransferAssetRequest{ID: 1, FromAccountID: 1001, ToAccountID: 1002, CoinID: 3},
	}

	for _, req := range reqs {
		t.Run(req.ActionName(), func(t *testing.T) {
			sig, err := s.SignAction(req, 42, key)
			require.NoError(t, err)

			got, err := s.RecoverPublicKeyFromRequest(req, 42, sig)
			require.NoError(t, err)
			assert.Equal(t, wantAddr[:], got, "recovered address must match signer")
		})
	}
}

// ── 3. Cross-engine isolation ─────────────────────────────────────────────────

// TestCrossEngineIsolation is the primary defence-in-depth test against cross-engine
// replay attacks. A signature produced for the Spark (spot) domain must NOT recover
// the correct signer address when verified under the Bolt (perps) domain.
func TestCrossEngineIsolation(t *testing.T) {
	key, wantAddr := mustLoadKey(t)
	spotSigner := newSigner(types.SpotDomainName)
	perpsSigner := newSigner(types.PerpsDomainName)

	req := &types.ScheduleCancelRequest{AccountID: 1001}
	sig, err := spotSigner.SignAction(req, 5, key)
	require.NoError(t, err)

	// Recovery under the wrong domain succeeds cryptographically but yields
	// a different (wrong) address — the signature is useless for the perps engine.
	recovered, err := perpsSigner.RecoverPublicKeyFromRequest(req, 5, sig)
	require.NoError(t, err)
	assert.NotEqual(t, wantAddr[:], recovered,
		"spot signature must not recover the correct address under the perps domain")
}

// ── 4. Determinism ────────────────────────────────────────────────────────────

// TestSignAction_Determinism verifies that signing the same (key, request, nonce)
// triple always produces an identical signature. This relies on go-ethereum's
// deterministic ECDSA implementation (RFC 6979).
func TestSignAction_Determinism(t *testing.T) {
	key, _ := mustLoadKey(t)
	s := newSigner(types.SpotDomainName)
	req := &types.ScheduleCancelRequest{AccountID: 1001}

	sig1, err := s.SignAction(req, 3, key)
	require.NoError(t, err)
	sig2, err := s.SignAction(req, 3, key)
	require.NoError(t, err)

	assert.True(t, bytes.Equal(sig1, sig2),
		"repeated signing of identical inputs must produce identical signatures (RFC 6979)")
}

// ── 5. Sensitivity ────────────────────────────────────────────────────────────

// TestSignAction_NonceSensitivity verifies that the nonce is baked into the
// EIP-712 digest: the same request signed with different nonces must produce
// different signatures, preventing replay of an old signature at a future nonce.
func TestSignAction_NonceSensitivity(t *testing.T) {
	key, _ := mustLoadKey(t)
	s := newSigner(types.SpotDomainName)
	req := &types.ScheduleCancelRequest{AccountID: 1001}

	sig0, err := s.SignAction(req, 0, key)
	require.NoError(t, err)
	sig1, err := s.SignAction(req, 1, key)
	require.NoError(t, err)

	assert.False(t, bytes.Equal(sig0, sig1),
		"signatures with different nonces must not be equal")
}

// TestSignAction_ActionTypeSensitivity verifies that the action type string is
// baked into the payload hash: two requests with the same field content but
// different action names must produce different signatures.
func TestSignAction_ActionTypeSensitivity(t *testing.T) {
	key, _ := mustLoadKey(t)
	s := newSigner(types.SpotDomainName)
	const nonce = uint64(0)

	sigCancel, err := s.SignAction(&types.ScheduleCancelRequest{AccountID: 1001}, nonce, key)
	require.NoError(t, err)

	// TransferAssetRequest has a different ActionName ("transferAsset") which
	// changes the JSON envelope and therefore the payload hash.
	sigTransfer, err := s.SignAction(&types.TransferAssetRequest{ID: 1, FromAccountID: 1001}, nonce, key)
	require.NoError(t, err)

	assert.False(t, bytes.Equal(sigCancel, sigTransfer),
		"different action types must produce different signatures")
}

// TestSignAction_ParamSensitivity verifies that modifying a request parameter
// changes the resulting signature. This confirms the payload hash commits to all
// field values, not just the action name.
func TestSignAction_ParamSensitivity(t *testing.T) {
	key, _ := mustLoadKey(t)
	s := newSigner(types.SpotDomainName)
	const nonce = uint64(0)

	sig1, err := s.SignAction(&types.ScheduleCancelRequest{AccountID: 1001}, nonce, key)
	require.NoError(t, err)
	sig2, err := s.SignAction(&types.ScheduleCancelRequest{AccountID: 9999}, nonce, key)
	require.NoError(t, err)

	assert.False(t, bytes.Equal(sig1, sig2),
		"changing a request parameter must change the signature")
}

// ── 6. Error handling ─────────────────────────────────────────────────────────

// TestRecoverPublicKeyFromRequest_ErrInvalidSignatureLength verifies that
// RecoverPublicKeyFromRequest returns ErrInvalidSignatureLength for any signature
// whose length is not exactly 66 bytes.
func TestRecoverPublicKeyFromRequest_ErrInvalidSignatureLength(t *testing.T) {
	s := newSigner(types.SpotDomainName)
	req := &types.ScheduleCancelRequest{AccountID: 1}

	cases := []struct {
		desc string
		sig  []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"65 bytes (raw ECDSA, missing type prefix)", make([]byte, 65)},
		{"67 bytes (one byte too long)", make([]byte, 67)},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := s.RecoverPublicKeyFromRequest(req, 0, tc.sig)
			assert.ErrorIs(t, err, types.ErrInvalidSignatureLength)
		})
	}
}

// TestRecoverPublicKeyFromRequest_ErrInvalidSignatureType verifies that
// RecoverPublicKeyFromRequest returns ErrInvalidSignatureType when the leading
// byte is not SignatureTypeEIP712 (0x01).
func TestRecoverPublicKeyFromRequest_ErrInvalidSignatureType(t *testing.T) {
	s := newSigner(types.SpotDomainName)
	req := &types.ScheduleCancelRequest{AccountID: 1}

	cases := []struct {
		desc     string
		typeByte byte
	}{
		{"type=0 (Unknown)", 0x00},
		{"type=2 (hypothetical future type)", 0x02},
		{"type=0xFF", 0xFF},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sig := make([]byte, 66)
			sig[0] = tc.typeByte

			_, err := s.RecoverPublicKeyFromRequest(req, 0, sig)
			assert.ErrorIs(t, err, types.ErrInvalidSignatureType)
		})
	}
}

// ── 7. Domain separator ───────────────────────────────────────────────────────

// TestDomainSeparator_Idempotent verifies that calling DomainSeparator multiple
// times returns the same hash (the result is cached after the first call).
func TestDomainSeparator_Idempotent(t *testing.T) {
	domain := types.NewEIP712Domain(types.SpotDomainName, testChainID)

	sep1 := domain.DomainSeparator()
	sep2 := domain.DomainSeparator()

	assert.Equal(t, sep1, sep2, "DomainSeparator must be idempotent")
}

// TestDomainSeparator_EngineDistinct verifies that the spot and perps engines
// have different domain separators. This is the root of cross-engine isolation.
func TestDomainSeparator_EngineDistinct(t *testing.T) {
	spotDomain := types.NewEIP712Domain(types.SpotDomainName, testChainID)
	perpsDomain := types.NewEIP712Domain(types.PerpsDomainName, testChainID)

	assert.NotEqual(t, spotDomain.DomainSeparator(), perpsDomain.DomainSeparator(),
		"spot and perps domains must have distinct separators")
}

// TestDomainSeparator_ChainIDSensitivity verifies that the same domain name on
// different chain IDs produces different separators, preventing cross-chain replay.
func TestDomainSeparator_ChainIDSensitivity(t *testing.T) {
	domain1 := types.NewEIP712Domain(types.SpotDomainName, 286623)
	domain2 := types.NewEIP712Domain(types.SpotDomainName, 1) // Ethereum mainnet

	assert.NotEqual(t, domain1.DomainSeparator(), domain2.DomainSeparator(),
		"domains on different chains must have distinct separators")
}

// ── helpers ───────────────────────────────────────────────────────────────────

// ptr returns a pointer to v. It exists only to make inline pointer literals
// readable in table-driven test cases.
func ptr[T any](v T) *T { return &v }
