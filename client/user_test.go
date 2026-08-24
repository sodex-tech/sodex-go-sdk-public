package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
)

// TestApproveBuilderFee validates the aggregate user endpoint body, universal signature, and signer identity.
func TestApproveBuilderFee(t *testing.T) {
	privateKey, err := crypto.HexToECDSA("0123456789012345678901234567890123456789012345678901234567890123")
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	request := &ctypes.ApproveBuilderFeeRequest{AccountID: 1010, BuilderID: 9, MaxFeeRate: 20}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/user/"+address.Hex()+"/builders" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Chain") != strconv.FormatUint(TestnetChainID, 10) {
			t.Fatalf("X-API-Chain = %q", r.Header.Get("X-API-Chain"))
		}

		var body ctypes.ApproveBuilderFeeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body != *request {
			t.Fatalf("body = %+v, want %+v", body, *request)
		}

		nonce, err := strconv.ParseUint(r.Header.Get("X-API-Nonce"), 10, 64)
		if err != nil {
			t.Fatalf("parse nonce: %v", err)
		}
		signature, err := hex.DecodeString(r.Header.Get("X-API-Sign")[2:])
		if err != nil {
			t.Fatalf("decode signature: %v", err)
		}
		if len(signature) != 66 || signature[0] != byte(enums.SignatureTypeEIP712Universal) {
			t.Fatalf("unexpected signature wire format: %x", signature)
		}

		domain := ctypes.NewEIP712Domain(ctypes.UniversalDomainName, TestnetChainID)
		action := &ctypes.ApproveBuilderFeeAction{
			ChainID:    TestnetChainID,
			Nonce:      nonce,
			AccountID:  request.AccountID,
			BuilderID:  request.BuilderID,
			MaxFeeRate: request.MaxFeeRate,
		}
		publicKey, err := crypto.SigToPub(action.Hash(&domain).Bytes(), signature[1:])
		if err != nil {
			t.Fatalf("recover signature: %v", err)
		}
		if got := crypto.PubkeyToAddress(*publicKey); got != address {
			t.Fatalf("recovered signer = %s, want %s", got.Hex(), address.Hex())
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":null}`))
	}))
	defer srv.Close()

	c := New(Config{
		BaseURL:    srv.URL,
		ChainID:    TestnetChainID,
		PrivateKey: privateKey,
		HTTPClient: srv.Client(),
	})
	if err := c.ApproveBuilderFee(context.Background(), address.Hex(), request); err != nil {
		t.Fatalf("ApproveBuilderFee: %v", err)
	}
}
