package types

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type ActionPayloadParams interface {
	ActionName() string
}

type ActionPayload struct {
	Type   string              `json:"type"`
	Params ActionPayloadParams `json:"params"`
}

func (ap *ActionPayload) Hash() (common.Hash, error) {
	bz, err := json.Marshal(ap)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(bz), nil
}
