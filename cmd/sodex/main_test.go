package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sodex-tech/sodex-go-sdk-public/ws"
)

func capturePushOutput(t *testing.T, printer func(ws.Push), data string) string {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "push-output-*")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	stdout := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = stdout }()

	printer(ws.Push{Data: json.RawMessage(data)})
	if _, err := output.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	printed, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	return string(printed)
}

// TestCandlePrinter validates that the Gateway's single-object candle push appears in pretty output.
func TestCandlePrinter(t *testing.T) {
	data := `{"t":1700000000000,"T":1700000059999,"s":"BTC-USD","i":"1m","o":"100","h":"101","l":"99","c":"100.5","v":"2","q":"200","n":3,"x":true}`
	printed := capturePushOutput(t, candlePrinter(formatPretty), data)
	if !strings.Contains(printed, "BTC-USD") || !strings.Contains(printed, "C=100.5") || !strings.Contains(printed, "[CLOSED]") {
		t.Fatalf("candle output = %q", printed)
	}
}

// TestOrderUpdatesPrinter validates that spot and perps batch pushes print every order update.
func TestOrderUpdatesPrinter(t *testing.T) {
	for _, tc := range []struct {
		engine string
		data   string
	}{
		{"spot", `[{"E":1700000000000,"T":1700000000000,"s":"vBTC_vUSDC","i":11,"S":"BUY","o":"LIMIT","p":"100","q":"1","X":"FILLED","z":"1","x":"TRADE"},{"E":1700000000000,"T":1700000000000,"s":"vETH_vUSDC","i":12,"S":"SELL","o":"LIMIT","p":"200","q":"2","X":"CANCELED","z":"0","x":"CANCELED"}]`},
		{"perps", `[{"E":1700000000000,"T":1700000000000,"s":"BTC-USD","i":11,"S":"BUY","o":"LIMIT","p":"100","q":"1","X":"FILLED","z":"1","x":"TRADE","R":false},{"E":1700000000000,"T":1700000000000,"s":"ETH-USD","i":12,"S":"SELL","o":"LIMIT","p":"200","q":"2","X":"CANCELED","z":"0","x":"CANCELED","R":true}]`},
	} {
		t.Run(tc.engine, func(t *testing.T) {
			printed := capturePushOutput(t, orderUpdatesPrinter(formatPretty, tc.engine), tc.data)
			if !strings.Contains(printed, "oid=11") || !strings.Contains(printed, "oid=12") || strings.Count(strings.TrimSpace(printed), "\n") != 1 {
				t.Fatalf("order updates output = %q", printed)
			}
		})
	}
}

// TestFillsPrinter validates that spot and perps batch pushes print every fill and maker status.
func TestFillsPrinter(t *testing.T) {
	for _, tc := range []struct {
		engine string
		data   string
	}{
		{"spot", `[{"E":1700000000000,"T":1700000000000,"t":1,"s":"vBTC_vUSDC","i":11,"S":"BUY","p":"100","q":"1","f":"0.01","m":true},{"E":1700000000000,"T":1700000000000,"t":2,"s":"vETH_vUSDC","i":12,"S":"SELL","p":"200","q":"2","f":"0.02","m":false}]`},
		{"perps", `[{"E":1700000000000,"T":1700000000000,"t":1,"s":"BTC-USD","i":11,"S":"BUY","p":"100","q":"1","f":"0.01","m":true,"d":"LONG"},{"E":1700000000000,"T":1700000000000,"t":2,"s":"ETH-USD","i":12,"S":"SELL","p":"200","q":"2","f":"0.02","m":false,"d":"SHORT"}]`},
	} {
		t.Run(tc.engine, func(t *testing.T) {
			printed := capturePushOutput(t, fillsPrinter(formatPretty, tc.engine), tc.data)
			if !strings.Contains(printed, "maker") || !strings.Contains(printed, "taker") || strings.Count(strings.TrimSpace(printed), "\n") != 1 {
				t.Fatalf("fills output = %q", printed)
			}
		})
	}
}
