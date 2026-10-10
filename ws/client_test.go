package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	wstypes "github.com/sodex-tech/sodex-go-sdk-public/ws/types"
)

// TestDispatchRoutesBySubscriptionParams checks symbol and coin filtering for array pushes
// and symbol/interval filtering for single-object pushes on shared channels.
func TestDispatchRoutesBySubscriptionParams(t *testing.T) {
	c, err := NewClient("http://example.com", "perps")
	if err != nil {
		t.Fatal(err)
	}

	var btcTrades, ethTrades []*wstypes.WsTrade
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelTrade, Symbols: []string{"BTC-USD"}}, func(push Push) {
		if err := json.Unmarshal(push.Data, &btcTrades); err != nil {
			t.Fatal(err)
		}
	})
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelTrade, Symbols: []string{"ETH-USD"}}, func(push Push) {
		if err := json.Unmarshal(push.Data, &ethTrades); err != nil {
			t.Fatal(err)
		}
	})
	c.dispatch([]byte(`{"channel":"trade","type":"update","data":[{"s":"BTC-USD"},{"s":"ETH-USD"}]}`))
	if len(btcTrades) != 1 || btcTrades[0].Symbol != "BTC-USD" || len(ethTrades) != 1 || ethTrades[0].Symbol != "ETH-USD" {
		t.Fatalf("trade routing: BTC=%+v ETH=%+v", btcTrades, ethTrades)
	}

	var btcPrices, usdcPrices []*wstypes.WsCoinPrice
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelCoinPrice, Coins: []string{"BTC"}}, func(push Push) {
		if err := json.Unmarshal(push.Data, &btcPrices); err != nil {
			t.Fatal(err)
		}
	})
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelCoinPrice, Coins: []string{"vUSDC"}}, func(push Push) {
		if err := json.Unmarshal(push.Data, &usdcPrices); err != nil {
			t.Fatal(err)
		}
	})
	c.dispatch([]byte(`{"channel":"coinPrice","type":"update","data":[{"a":"BTC","p":"100"},{"a":"vUSDC","p":"1"}]}`))
	if len(btcPrices) != 1 || btcPrices[0].Coin != "BTC" || len(usdcPrices) != 1 || usdcPrices[0].Coin != "vUSDC" {
		t.Fatalf("coin price routing: BTC=%+v vUSDC=%+v", btcPrices, usdcPrices)
	}

	var oneMinute, fiveMinute int
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelCandle, Symbol: "BTC-USD", Interval: "1m"}, func(Push) { oneMinute++ })
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelCandle, Symbol: "BTC-USD", Interval: "5m"}, func(Push) { fiveMinute++ })
	c.dispatch([]byte(`{"channel":"candle","type":"update","data":{"s":"BTC-USD","i":"5m"}}`))
	if oneMinute != 0 || fiveMinute != 1 {
		t.Fatalf("candle routing: 1m=%d 5m=%d", oneMinute, fiveMinute)
	}

	var btcBooks, ethBooks int
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelL4Book, Symbol: "BTC-USD"}, func(Push) { btcBooks++ })
	_, _ = c.Subscribe(SubscribeParams{Channel: ChannelL4Book, Symbol: "ETH-USD"}, func(Push) { ethBooks++ })
	c.dispatch([]byte(`{"channel":"l4Book","type":"update","data":{"s":"BTC-USD"}}`))
	if btcBooks != 1 || ethBooks != 0 {
		t.Fatalf("book routing: BTC=%d ETH=%d", btcBooks, ethBooks)
	}
}

// TestUnsubscribeSendsEachRequest verifies that removing one of two trade
// subscriptions sends its exact unsubscribe request while the other remains.
func TestUnsubscribeSendsEachRequest(t *testing.T) {
	requests := make(chan Request, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for i := 0; i < 3; i++ {
			var req Request
			if err := conn.ReadJSON(&req); err != nil {
				return
			}
			requests <- req
		}
	}))
	defer server.Close()

	c, err := NewClient(server.URL, "perps")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.dial(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	id, err := c.Subscribe(SubscribeParams{Channel: ChannelTrade, Symbols: []string{"BTC-USD"}}, func(Push) {})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Subscribe(SubscribeParams{Channel: ChannelTrade, Symbols: []string{"ETH-USD"}}, func(Push) {}); err != nil {
		t.Fatal(err)
	}
	if err := c.Unsubscribe(id); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		select {
		case req := <-requests:
			wantOp := "subscribe"
			if i == 2 {
				wantOp = "unsubscribe"
			}
			if req.Op != wantOp {
				t.Fatalf("request %d op = %q, want %q", i, req.Op, wantOp)
			}
			if i == 2 {
				var params SubscribeParams
				if err := json.Unmarshal(req.Params, &params); err != nil {
					t.Fatal(err)
				}
				if req.ID != id || len(params.Symbols) != 1 || params.Symbols[0] != "BTC-USD" {
					t.Fatalf("unsubscribe request = %+v, params = %+v", req, params)
				}
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for request %d", i)
		}
	}
}

// TestConnectStopsOnContextCancellation verifies cancellation closes an idle
// socket promptly and does not report the expected close as a read error.
func TestConnectStopsOnContextCancellation(t *testing.T) {
	accepted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		close(accepted)
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	c, err := NewClient(server.URL, "perps")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	readErrors := make(chan error, 1)
	c.OnError(func(err error) { readErrors <- err })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	connected := make(chan error, 1)
	go func() { connected <- c.Connect(ctx) }()
	select {
	case <-accepted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for WebSocket connection")
	}
	cancel()
	select {
	case err := <-connected:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Connect returned %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Connect did not return promptly after cancellation")
	}
	select {
	case err := <-readErrors:
		t.Fatalf("unexpected read error on cancellation: %v", err)
	default:
	}
}

// TestAccountChannelRequiresSeparateClients checks that ambiguous account
// pushes cannot be delivered to handlers for two accounts on one connection.
func TestAccountChannelRequiresSeparateClients(t *testing.T) {
	c, err := NewClient("http://example.com", "perps")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Subscribe(SubscribeParams{Channel: ChannelAccountTrade, User: "0x1"}, func(Push) {}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Subscribe(SubscribeParams{Channel: ChannelAccountTrade, User: "0x2"}, func(Push) {}); err == nil {
		t.Fatal("expected second account subscription to be rejected")
	}
}

// TestBookUpdateSequenceFields checks snapshot and delta IDs using their distinct Gateway payload types.
func TestBookUpdateSequenceFields(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload string
		firstID uint64
		lastID  uint64
	}{
		{name: "snapshot", payload: `{"channel":"l4Book","type":"snapshot","data":{"s":"BTC-USD","u":7,"a":[],"b":[]}}`, lastID: 7},
		{name: "update", payload: `{"channel":"l4Book","type":"update","data":{"s":"BTC-USD","U":8,"u":10,"a":[["100","1"]],"b":[]}}`, firstID: 8, lastID: 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var push Push
			if err := json.Unmarshal([]byte(tc.payload), &push); err != nil {
				t.Fatal(err)
			}
			if push.Type == "snapshot" {
				var book wstypes.WsDepthSnapshot
				if err := json.Unmarshal(push.Data, &book); err != nil || book.UpdateID != tc.lastID {
					t.Fatalf("snapshot = %+v, error = %v", book, err)
				}
			} else {
				var book wstypes.WsDepthUpdate
				if err := json.Unmarshal(push.Data, &book); err != nil || book.FirstUpdateID != tc.firstID || book.LastUpdateID != tc.lastID {
					t.Fatalf("update = %+v, error = %v", book, err)
				}
			}
		})
	}
}

// TestAccountTradeBuilderFee checks the copied account trade type retains the optional builder fee.
func TestAccountTradeBuilderFee(t *testing.T) {
	var trade wstypes.WsUserSpotTrade
	if err := json.Unmarshal([]byte(`{"s":"BTC-USD","f":"0.02","bf":"0.01"}`), &trade); err != nil {
		t.Fatal(err)
	}
	if trade.BuilderFee == nil || *trade.BuilderFee != "0.01" {
		t.Fatalf("account trade: %+v", trade)
	}
}
