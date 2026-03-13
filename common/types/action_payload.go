package types

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ActionPayloadParams is the interface that every signable action must implement.
// Each request type returns a unique action name string used as the "type" discriminator
// when the payload is JSON-encoded before hashing. This ensures that two structurally
// identical requests for different actions (e.g., a cancel vs. a replace) always
// produce distinct digests.
type ActionPayloadParams interface {
	ActionName() string
}

// ActionPayload is the canonical envelope for any exchange action before it is hashed.
//
// The two-field structure mirrors a tagged union: Type identifies the action kind, and
// Params carries the typed parameters. Both fields are included in the JSON serialisation
// so that the resulting hash commits to the action name as well as its data.
//
//	{
//	  "type":   "newOrder",
//	  "params": { "accountID": 1, "symbolID": 42, ... }
//	}
//
// See Hash for how this envelope is reduced to a 32-byte digest used in EIP-712.
type ActionPayload struct {
	Type   string              `json:"type"`
	Params ActionPayloadParams `json:"params"`
}

// Hash serialises the payload to canonical JSON and returns its keccak256 digest.
//
// This digest serves as the payloadHash field in ExchangeAction, acting as the
// first of two hashing layers in the signing pipeline:
//
//	ActionPayload ──JSON──▶ keccak256 ──▶ payloadHash (bytes32)
//	                                            │
//	                                 ExchangeAction.StructHash
//	                                            │
//	                                   EIP-712 final digest
//
// Compressing arbitrary-length action data into a fixed-size hash at this stage
// keeps the EIP-712 type schema simple and independent of the concrete action type.
func (ap *ActionPayload) Hash() (common.Hash, error) {
	bz, err := json.Marshal(ap)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(bz), nil
}
