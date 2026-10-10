package types

import (
	"encoding/json"
	"testing"
)

func TestPushTypesDecode(t *testing.T) {
	var ticker WsTicker
	if err := json.Unmarshal([]byte(`{"E":10,"s":"BTC-USD","c":"100.5","Q":"0.2","v":"5"}`), &ticker); err != nil {
		t.Fatal(err)
	}
	if ticker.Ticker == nil || ticker.Symbol != "BTC-USD" || ticker.ClosePx.String() != "100.5" || ticker.LastSz.String() != "0.2" {
		t.Fatalf("ticker: %+v", ticker)
	}

	var order WsPerpsOrderUpdate
	if err := json.Unmarshal([]byte(`{"E":11,"T":9,"s":"BTC-USD","i":42,"R":true,"sp":"99","x":"TRADE"}`), &order); err != nil {
		t.Fatal(err)
	}
	if order.PerpsOrderUpdate == nil || order.PerpsOrder == nil || order.SpotOrder == nil || order.OrderID != 42 || !order.ReduceOnly || order.StopPrice == nil || order.StopPrice.String() != "99" {
		t.Fatalf("perps order update: %+v", order)
	}

	var fill WsUserPerpsTrade
	if err := json.Unmarshal([]byte(`{"E":12,"T":9,"t":1,"s":"BTC-USD","i":42,"bf":"0.01","d":"LONG"}`), &fill); err != nil {
		t.Fatal(err)
	}
	if fill.UserPerpsTrade == nil || fill.UserTrade == nil || fill.BuilderFee == nil || *fill.BuilderFee != "0.01" || fill.Dir != "LONG" {
		t.Fatalf("perps fill: %+v", fill)
	}
}
