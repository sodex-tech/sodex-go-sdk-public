package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ptypes "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
	stypes "github.com/sodex-tech/sodex-go-sdk-public/spot/types"
)

// TestTradingResponses keeps item-level successes and failures from the spot and perps batch response shapes.
func TestTradingResponses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/spot/trade/orders/batch":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"code":0,"clOrdID":"one","orderID":42},{"code":-1,"clOrdID":"two","error":"insufficient balance"}]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/spot/trade/orders/batch":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"code":0,"clOrdID":"cancel-one","orderID":42,"origClOrdID":"one"}]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/perps/trade/orders":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"code":-1,"error":"order not found"}]}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	key, err := crypto.HexToECDSA("0123456789012345678901234567890123456789012345678901234567890123")
	if err != nil {
		t.Fatal(err)
	}
	c := New(Config{BaseURL: srv.URL, PrivateKey: key, HTTPClient: srv.Client()})
	ctx := context.Background()

	placed, err := c.PlaceSpotOrders(ctx, &stypes.BatchNewOrderRequest{AccountID: 1, Orders: []*stypes.BatchNewOrderItem{{SymbolID: 2, ClOrdID: "one", Side: enums.OrderSideBuy, Type: enums.OrderTypeMarket, TimeInForce: enums.TimeInForceIOC}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(placed) != 2 || placed[0].Code != 0 || placed[0].OrderID == nil || *placed[0].OrderID != 42 || placed[1].Code != -1 || placed[1].Error == nil || *placed[1].Error != "insufficient balance" || placed[1].OrderID != nil {
		t.Fatalf("unexpected place results: %+v", placed)
	}

	cancelledSpot, err := c.CancelSpotOrders(ctx, &stypes.BatchCancelOrderRequest{AccountID: 1, Cancels: []*stypes.BatchCancelOrderItem{{SymbolID: 2, ClOrdID: "cancel-one"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(cancelledSpot) != 1 || cancelledSpot[0].OrigClOrdID == nil || *cancelledSpot[0].OrigClOrdID != "one" {
		t.Fatalf("unexpected spot cancel results: %+v", cancelledSpot)
	}

	orderID := uint64(42)
	cancelledPerps, err := c.CancelPerpsOrders(ctx, &ptypes.CancelOrderRequest{AccountID: 1, Cancels: []*ptypes.CancelOrder{{SymbolID: 2, OrderID: &orderID}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(cancelledPerps) != 1 || cancelledPerps[0].Code != -1 || cancelledPerps[0].Error == nil || *cancelledPerps[0].Error != "order not found" || cancelledPerps[0].ClOrdID != nil {
		t.Fatalf("unexpected perps cancel results: %+v", cancelledPerps)
	}
}

// TestPerpsBalances decodes collateral fields from the nested balance response without a spot-only locked field.
func TestPerpsBalances(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/perps/accounts/0xabc/balances" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"blockTime":100,"blockHeight":20,"balances":[{"id":1,"coin":"vUSDC","total":"10","collateral":"8","marginRatio":"0.5","price":"1"}]}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	balances, err := c.PerpsBalances(context.Background(), "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 1 || balances[0].Collateral != "8" || balances[0].MarginRatio != "0.5" || balances[0].Price == nil || *balances[0].Price != "1" {
		t.Fatalf("unexpected perps balances: %+v", balances)
	}
}

// TestUpdateLeverage treats a null success payload as success and still reports application errors.
func TestUpdateLeverage(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/perps/trade/leverage" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		calls++
		if calls == 1 {
			_, _ = w.Write([]byte(`{"code":0,"data":null}`))
		} else {
			_, _ = w.Write([]byte(`{"code":-1,"error":"invalid leverage"}`))
		}
	}))
	defer srv.Close()

	key, err := crypto.HexToECDSA("0123456789012345678901234567890123456789012345678901234567890123")
	if err != nil {
		t.Fatal(err)
	}
	c := New(Config{BaseURL: srv.URL, PrivateKey: key, HTTPClient: srv.Client()})
	req := &ptypes.UpdateLeverageRequest{AccountID: 1, SymbolID: 2, Leverage: 5, MarginMode: enums.MarginModeCross}
	if err := c.UpdateLeverage(context.Background(), req); err != nil {
		t.Fatalf("success response: %v", err)
	}
	var apiErr *ErrAPI
	if err := c.UpdateLeverage(context.Background(), req); !errors.As(err, &apiErr) || apiErr.Code != -1 || apiErr.Message != "invalid leverage" {
		t.Fatalf("error response: %v", err)
	}
}

// TestOrderBooks sends the gateway's limit parameter for both engines and preserves snapshot metadata and level arrays.
func TestOrderBooks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "25" || r.URL.Query().Has("depth") {
			t.Errorf("unexpected orderbook query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"blockTime":100,"blockHeight":20,"updateID":7,"bids":[["1","2"]],"asks":[["3","4"]]}}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	for _, query := range []func() (*OrderBook, error){
		func() (*OrderBook, error) { return c.SpotOrderBook(context.Background(), "vBTC_vUSDC", 25) },
		func() (*OrderBook, error) { return c.PerpsOrderBook(context.Background(), "BTC-USD", 25) },
	} {
		book, err := query()
		if err != nil {
			t.Fatal(err)
		}
		if book.BlockTime != 100 || book.BlockHeight != 20 || book.UpdateID != 7 || len(book.Bids) != 1 || book.Bids[0].Price != "1" {
			t.Fatalf("unexpected orderbook: %+v", book)
		}
		encoded, err := json.Marshal(book)
		if err != nil {
			t.Fatal(err)
		}
		var wire struct {
			Bids [][]string `json:"bids"`
		}
		if err := json.Unmarshal(encoded, &wire); err != nil || len(wire.Bids) != 1 || wire.Bids[0][0] != "1" {
			t.Fatalf("orderbook level did not preserve array shape: %s (%v)", encoded, err)
		}
	}
}

// TestAPIResponse decodes the gateway's timestamp and error fields in the public envelope type.
func TestAPIResponse(t *testing.T) {
	var response APIResponse[json.RawMessage]
	if err := json.Unmarshal([]byte(`{"code":-1,"timestamp":123,"error":"invalid request"}`), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != -1 || response.Timestamp != 123 || response.Error == nil || *response.Error != "invalid request" {
		t.Fatalf("unexpected envelope: %+v", response)
	}
}
