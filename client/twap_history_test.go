package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	rpctypes "github.com/sodex-tech/sodex-go-sdk-public/client/types"
)

// TestTwapOrdersHistory verifies address paths and accountID/limit-only queries
// in both markets, preserving exact order IDs and decimals without user/account/symbol IDs.
func TestTwapOrdersHistory(t *testing.T) {
	for _, market := range []string{"spot", "perps"} {
		t.Run(market, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				want := url.Values{"accountID": {"9007199254740993"}, "limit": {"50"}}
				if r.Method != http.MethodGet || r.URL.Path != "/api/v1/"+market+"/accounts/0xabc/twaps/history" || !reflect.DeepEqual(r.URL.Query(), want) {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				_, _ = w.Write([]byte(`{"code":0,"data":[{"symbol":"BTC-USD","orderID":9007199254740993,"side":"SELL","quantity":"2","executedQty":"1.25","executedValue":"125.000000000000000001","minutes":5,"randomize":true,"reduceOnly":false,"createdAt":1000,"nextActiveAt":1001,"active":false}]}`))
			}))
			defer srv.Close()
			c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
			read := c.SpotTwapOrdersHistory
			if market == "perps" {
				read = c.PerpsTwapOrdersHistory
			}
			orders, err := read(context.Background(), "0xabc", HistoryFilter{AccountID: 9007199254740993, Symbol: "ignored", StartTime: 1000, EndTime: 2000, Limit: 50})
			if err != nil {
				t.Fatal(err)
			}
			want := []*rpctypes.TwapOrder{{Symbol: "BTC-USD", OrderID: 9007199254740993, Side: "SELL", Quantity: "2", ExecutedQty: "1.25", ExecutedValue: "125.000000000000000001", Minutes: 5, Randomize: true, CreatedAt: 1000, NextActiveAt: 1001}}
			if !reflect.DeepEqual(orders, want) {
				t.Fatalf("orders = %+v, want %+v", orders, want)
			}
			encoded, err := json.Marshal(orders)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{`"userID"`, `"accountID"`, `"symbolID"`} {
				if strings.Contains(string(encoded), key) {
					t.Fatalf("TWAP orders expose %s: %s", key, encoded)
				}
			}
		})
	}
}

// TestTwapFillHistory verifies the independent fill endpoint for both markets,
// no client-imposed defaults, and decoding of the existing account trade fields.
func TestTwapFillHistory(t *testing.T) {
	for _, market := range []string{"spot", "perps"} {
		t.Run(market, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/"+market+"/accounts/0xabc/twaps/trades" || r.URL.RawQuery != "" {
					t.Errorf("unexpected request: %s", r.URL)
				}
				_, _ = w.Write([]byte(`{"code":0,"data":[{"symbol":"BTC-USD","tradeID":9007199254740993,"orderID":9007199254740995,"clOrdID":"twap:1:1","side":"BUY","price":"100.000000000000000001","quantity":"0.1","fee":"-0.000000000000000001","builderFee":"0.000000000000000001","feeCoin":"vUSDC","time":1000,"isMaker":true}]}`))
			}))
			defer srv.Close()
			c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
			read := c.SpotTwapFillHistory
			if market == "perps" {
				read = c.PerpsTwapFillHistory
			}
			fills, err := read(context.Background(), "0xabc", HistoryFilter{})
			if err != nil {
				t.Fatal(err)
			}
			builderFee := "0.000000000000000001"
			want := []*rpctypes.AccountTrade{{Symbol: "BTC-USD", TradeID: 9007199254740993, OrderID: 9007199254740995, ClOrdID: "twap:1:1", Side: "BUY", Price: "100.000000000000000001", Quantity: "0.1", Fee: "-0.000000000000000001", BuilderFee: &builderFee, FeeCoin: "vUSDC", Timestamp: 1000, IsMaker: true}}
			if !reflect.DeepEqual(fills, want) {
				t.Fatalf("fills = %+v, want %+v", fills, want)
			}
		})
	}
}

// TestPerpsPositionsFunding verifies funding is decoded from existing positions,
// preserving large IDs, signed/zero amounts and serialization of empty Go values.
func TestPerpsPositionsFunding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/perps/accounts/0xabc/positions" || r.URL.RawQuery != "" {
			t.Errorf("unexpected request: %s", r.URL)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"blockHeight":20,"positions":[{"id":9007199254740993,"cumFundingFee":"-0.000000000000000001"},{"id":2,"cumFundingFee":"0"}]}}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	actual, err := c.PerpsPositions(context.Background(), "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	want := []*rpctypes.PerpsPosition{{ID: 9007199254740993, CumFundingFee: "-0.000000000000000001"}, {ID: 2, CumFundingFee: "0"}}
	if actual.BlockHeight != 20 || !reflect.DeepEqual(actual.Positions, want) {
		t.Fatalf("funding = %+v, want %+v", actual, want)
	}
	encoded, err := json.Marshal(rpctypes.PerpsPosition{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"cumFundingFee":""`) {
		t.Fatalf("funding field was omitted: %s", encoded)
	}
}

// TestTwapReadsEmptyAndError verifies every added method handles empty accounts
// and preserves application errors instead of returning invented history rows.
func TestTwapReadsEmptyAndError(t *testing.T) {
	for _, failure := range []bool{false, true} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if failure {
				_, _ = w.Write([]byte(`{"code":1,"error":"unavailable"}`))
				return
			}
			_, _ = w.Write([]byte(`{"code":0,"data":[]}`))
		}))
		c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
		for _, read := range []func() (int, error){
			func() (int, error) {
				rows, err := c.SpotTwapOrdersHistory(context.Background(), "0xabc", HistoryFilter{})
				return len(rows), err
			},
			func() (int, error) {
				rows, err := c.PerpsTwapOrdersHistory(context.Background(), "0xabc", HistoryFilter{})
				return len(rows), err
			},
			func() (int, error) {
				rows, err := c.SpotTwapFillHistory(context.Background(), "0xabc", HistoryFilter{})
				return len(rows), err
			},
			func() (int, error) {
				rows, err := c.PerpsTwapFillHistory(context.Background(), "0xabc", HistoryFilter{})
				return len(rows), err
			},
		} {
			count, err := read()
			if count != 0 || failure && (err == nil || !strings.Contains(err.Error(), "unavailable")) || !failure && err != nil {
				t.Errorf("failure=%v: count=%d err=%v", failure, count, err)
			}
		}
		srv.Close()
	}
}
