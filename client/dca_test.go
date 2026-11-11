package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	"github.com/sodex-tech/sodex-go-sdk-public/common/signer"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
)

const dcaPlanID uint64 = 9007199254740993
const dcaPlanJSON = `{"symbol":"BTC-USD","orderID":9007199254740993,"amount":"100.000000000000000001","side":"BUY","schedule":{"kind":"monthly","hour":0,"day":31},"minPrice":"1","maxPrice":"90000","randomize":false,"immediate":true,"executedQty":"0.001","executedValue":"90","failures":2,"createdAt":1700000000000,"nextActiveAt":1700086400000,"status":"PRUNED"}`

// TestDcaQueries checks both markets' active/history/fill routes, supported filters, exact IDs and plan/child linkage.
func TestDcaQueries(t *testing.T) {
	for _, market := range []string{"spot", "perps"} {
		t.Run(market, func(t *testing.T) {
			root := "/api/v1/" + market + "/accounts/0xabc/dcas"
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				want := url.Values{"accountID": {"9007199254740993"}}
				if r.URL.Path == root {
					want.Set("symbol", "BTC-USD")
				} else {
					want.Set("limit", "1000")
				}
				assert.Equal(t, want, r.URL.Query())
				switch r.URL.Path {
				case root:
					_, _ = io.WriteString(w, `{"code":0,"data":{"blockTime":10,"blockHeight":20,"dcas":[`+strings.Replace(dcaPlanJSON, "PRUNED", "ACTIVE", 1)+`]}}`)
				case root + "/history":
					_, _ = io.WriteString(w, `{"code":0,"data":[`+dcaPlanJSON+`]}`)
				case root + "/trades":
					_, _ = io.WriteString(w, `{"code":0,"data":[{"symbol":"BTC-USD","tradeID":9007199254740995,"orderID":9007199254740994,"clOrdID":"dca:9007199254740993","side":"BUY","price":"90000","quantity":"0.001","fee":"0.1","builderFee":"0.02","feeCoin":"vUSDC","time":1700000000000,"isMaker":false}]}`)
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
			}))
			defer srv.Close()
			c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
			active, history, fills := c.SpotDcaOrders, c.SpotDcaOrdersHistory, c.SpotDcaFillHistory
			if market == "perps" {
				active, history, fills = c.PerpsDcaOrders, c.PerpsDcaOrdersHistory, c.PerpsDcaFillHistory
			}
			filter := HistoryFilter{AccountID: dcaPlanID, Symbol: "BTC-USD", Limit: 1000, StartTime: 1, EndTime: 2}
			ctx := context.Background()
			snapshot, err := active(ctx, "0xabc", filter)
			require.NoError(t, err)
			require.Len(t, snapshot.Dcas, 1)
			require.EqualValues(t, 10, snapshot.BlockTime)
			require.EqualValues(t, 20, snapshot.BlockHeight)
			require.Equal(t, dcaPlanID, snapshot.Dcas[0].OrderID)
			require.Equal(t, "ACTIVE", snapshot.Dcas[0].Status)
			plans, err := history(ctx, "0xabc", filter)
			require.NoError(t, err)
			require.Len(t, plans, 1)
			plan := plans[0]
			require.Equal(t, "PRUNED", plan.Status)
			require.Equal(t, "100.000000000000000001", plan.Amount)
			require.EqualValues(t, 0, plan.Schedule.Hour)
			require.EqualValues(t, 31, *plan.Schedule.Day)
			require.Equal(t, "1", *plan.MinPrice)
			require.Equal(t, "90000", *plan.MaxPrice)
			require.EqualValues(t, 2, plan.Failures)
			trades, err := fills(ctx, "0xabc", filter)
			require.NoError(t, err)
			require.Len(t, trades, 1)
			require.Equal(t, dcaPlanID+1, trades[0].OrderID)
			require.Equal(t, dcaPlanID+2, trades[0].TradeID)
			require.Equal(t, "dca:9007199254740993", trades[0].ClOrdID)
			require.Equal(t, "0.02", *trades[0].BuilderFee)
			require.EqualValues(t, 1700000000000, trades[0].Timestamp)
		})
	}
}

// TestDcaSignedWrites checks create/cancel bodies, single results, nonce progression and signatures in both market domains.
func TestDcaSignedWrites(t *testing.T) {
	key, err := crypto.HexToECDSA("0000000000000000000000000000000000000000000000000000000000000001")
	require.NoError(t, err)
	for _, market := range []string{"spot", "perps"} {
		t.Run(market, func(t *testing.T) {
			type submission struct {
				method, path string
				header       http.Header
				body         []byte
			}
			calls := make(chan submission, 2)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				calls <- submission{r.Method, r.URL.Path, r.Header.Clone(), body}
				_, _ = io.WriteString(w, `{"code":0,"data":{"code":0,"orderID":9007199254740993}}`)
			}))
			defer srv.Close()
			c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client(), PrivateKey: key, APIKeyName: "dca-test"})
			create, cancel := c.PlaceSpotDcaOrder, c.CancelSpotDcaOrder
			domainName := ctypes.SpotDomainName
			if market == "perps" {
				create, cancel = c.PlacePerpsDcaOrder, c.CancelPerpsDcaOrder
				domainName = ctypes.PerpsDomainName
			}
			ctx := context.Background()
			newPlan := &ctypes.NewDcaOrderRequest{AccountID: 42, SymbolID: 7, Side: enums.OrderSideBuy, Amount: decimal.NewFromInt(100), Schedule: ctypes.DcaSchedule{Kind: "daily", Hour: 0}}
			cancelPlan := &ctypes.CancelDcaOrderRequest{AccountID: 42, SymbolID: 7, OrderID: dcaPlanID}
			created, err := create(ctx, newPlan)
			require.NoError(t, err)
			require.Equal(t, 0, created.Code)
			require.Equal(t, dcaPlanID, *created.OrderID)
			cancelled, err := cancel(ctx, cancelPlan)
			require.NoError(t, err)
			require.Equal(t, dcaPlanID, *cancelled.OrderID)
			domain := ctypes.NewEIP712Domain(domainName, DefaultChainID)
			verifier := signer.NewEVMSigner(&domain)
			var previousNonce uint64
			for i, params := range []ctypes.ActionPayloadParams{newPlan, cancelPlan} {
				call := <-calls
				require.Equal(t, []string{http.MethodPost, http.MethodDelete}[i], call.method)
				require.Equal(t, "/api/v1/"+market+"/trade/dcas", call.path)
				body, err := json.Marshal(params)
				require.NoError(t, err)
				require.Equal(t, string(body), string(call.body))
				require.Equal(t, "dca-test", call.header.Get("X-API-Key"))
				nonce, err := strconv.ParseUint(call.header.Get("X-API-Nonce"), 10, 64)
				require.NoError(t, err)
				require.Greater(t, nonce, previousNonce)
				previousNonce = nonce
				sig, err := hex.DecodeString(strings.TrimPrefix(call.header.Get("X-API-Sign"), "0x"))
				require.NoError(t, err)
				address, err := verifier.RecoverPublicKeyFromRequest(params, nonce, sig)
				require.NoError(t, err)
				require.Equal(t, crypto.PubkeyToAddress(key.PublicKey).Bytes(), address)
			}
		})
	}
}

// TestDcaWritesRequireAuthentication verifies all DCA writes reject missing credentials before sending a request.
func TestDcaWritesRequireAuthentication(t *testing.T) {
	c := New(Config{})
	ctx := context.Background()
	_, err := c.PlaceSpotDcaOrder(ctx, &ctypes.NewDcaOrderRequest{})
	require.ErrorIs(t, err, ErrNotAuthenticated)
	_, err = c.PlacePerpsDcaOrder(ctx, &ctypes.NewDcaOrderRequest{})
	require.ErrorIs(t, err, ErrNotAuthenticated)
	_, err = c.CancelSpotDcaOrder(ctx, &ctypes.CancelDcaOrderRequest{})
	require.ErrorIs(t, err, ErrNotAuthenticated)
	_, err = c.CancelPerpsDcaOrder(ctx, &ctypes.CancelDcaOrderRequest{})
	require.ErrorIs(t, err, ErrNotAuthenticated)
}

// TestDcaReadsEmptyAndError checks omitted filters, empty histories, null active collections and API errors for every DCA read.
func TestDcaReadsEmptyAndError(t *testing.T) {
	for _, market := range []string{"spot", "perps"} {
		for _, failure := range []bool{false, true} {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Empty(t, r.URL.RawQuery)
				if failure {
					_, _ = io.WriteString(w, `{"code":1,"error":"activity indexer unavailable"}`)
				} else if strings.HasSuffix(r.URL.Path, "/dcas") {
					_, _ = io.WriteString(w, `{"code":0,"data":{"blockTime":0,"blockHeight":0,"dcas":null}}`)
				} else {
					_, _ = io.WriteString(w, `{"code":0,"data":[]}`)
				}
			}))
			c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
			active, history, fills := c.SpotDcaOrders, c.SpotDcaOrdersHistory, c.SpotDcaFillHistory
			if market == "perps" {
				active, history, fills = c.PerpsDcaOrders, c.PerpsDcaOrdersHistory, c.PerpsDcaFillHistory
			}
			ctx := context.Background()
			plans, err := active(ctx, "0xabc", HistoryFilter{})
			if failure {
				require.ErrorContains(t, err, "activity indexer unavailable")
			} else {
				require.NoError(t, err)
				require.Empty(t, plans.Dcas)
			}
			rows, err := history(ctx, "0xabc", HistoryFilter{})
			if failure {
				require.ErrorContains(t, err, "activity indexer unavailable")
			} else {
				require.NoError(t, err)
				require.Empty(t, rows)
			}
			trades, err := fills(ctx, "0xabc", HistoryFilter{})
			if failure {
				require.ErrorContains(t, err, "activity indexer unavailable")
			} else {
				require.NoError(t, err)
				require.Empty(t, trades)
			}
			srv.Close()
		}
	}
}
