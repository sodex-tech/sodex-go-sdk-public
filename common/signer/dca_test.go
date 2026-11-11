package signer_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	"github.com/sodex-tech/sodex-go-sdk-public/common/types"
	perpssigner "github.com/sodex-tech/sodex-go-sdk-public/perps/signer"
	spotsigner "github.com/sodex-tech/sodex-go-sdk-public/spot/signer"
)

// TestDcaNovaVectors pins JSON, hashes and signatures for all schedules and optional bounds in both domains.
func TestDcaNovaVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/dca_nova_vectors.json")
	require.NoError(t, err)
	var vectors []struct {
		Action    string `json:"action"`
		Request   string `json:"request"`
		Hash      string `json:"hash"`
		Domain    string `json:"domain"`
		ChainID   uint64 `json:"chainID"`
		Signature string `json:"signature"`
	}
	require.NoError(t, json.Unmarshal(data, &vectors))
	// Public fixture key 1 and nonce 123 match Nova's generated vectors.
	key, err := crypto.HexToECDSA("0000000000000000000000000000000000000000000000000000000000000001")
	require.NoError(t, err)
	for i, vector := range vectors {
		t.Run(fmt.Sprintf("%d/%s/%s", i, vector.Action, vector.Domain), func(t *testing.T) {
			var params types.ActionPayloadParams
			if vector.Action == "cancelDcaOrder" {
				params = &types.CancelDcaOrderRequest{}
			} else {
				params = &types.NewDcaOrderRequest{}
			}
			require.NoError(t, json.Unmarshal([]byte(vector.Request), params))
			encoded, err := json.Marshal(params)
			require.NoError(t, err)
			require.Equal(t, vector.Request, string(encoded))
			payload := &types.ActionPayload{Type: params.ActionName(), Params: params}
			hash, err := payload.Hash()
			require.NoError(t, err)
			require.Equal(t, vector.Hash, hex.EncodeToString(hash.Bytes()))
			if vector.Signature != "" {
				request := params.(*types.NewDcaOrderRequest)
				var signature []byte
				if vector.Domain == "spot" {
					signature, err = spotsigner.NewSigner(vector.ChainID, key).SignNewDcaOrderRequest(request, 123)
				} else {
					signature, err = perpssigner.NewSigner(vector.ChainID, key).SignNewDcaOrderRequest(request, 123)
				}
				require.NoError(t, err)
				require.Equal(t, vector.Signature, hex.EncodeToString(signature))
			}
		})
	}
}

// TestDcaCanonicalDecimals preserves full precision, explicit midnight and false flags while omitting absent prices/day.
func TestDcaCanonicalDecimals(t *testing.T) {
	min := decimal.RequireFromString("1e-18")
	request := &types.NewDcaOrderRequest{
		AccountID: 42, SymbolID: 7, Side: enums.OrderSideBuy,
		Amount:   decimal.RequireFromString("9007199254740993.123456789012345678900"),
		Schedule: types.DcaSchedule{Kind: "daily", Hour: 0}, MinPrice: &min,
	}
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	require.Equal(t, `{"accountID":42,"symbolID":7,"side":1,"amount":"9007199254740993.1234567890123456789","schedule":{"kind":"daily","hour":0},"minPrice":"0.000000000000000001","randomize":false,"immediate":false}`, string(encoded))
}
