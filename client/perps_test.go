package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestPerpsPositionsParsesPositionsWrapper verifies position data and block metadata survive the wrapper.
func TestPerpsPositionsParsesPositionsWrapper(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/perps/accounts/0xabc/positions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 0,
			"data": {
				"blockTime": 1780563616021,
				"blockHeight": 161169726,
				"positions": [{
					"id": 6279292,
					"symbol": "BTC-USD",
					"marginMode": "CROSS",
					"positionSide": "BOTH",
					"size": "0.00025",
					"initialMargin": "3.1764",
					"avgEntryPrice": "63528",
					"cumOpenCost": "15.882",
					"cumTradingFee": "0.0063528",
					"cumClosedSize": "0",
					"avgClosePrice": "0",
					"maxSize": "0.00025",
					"realizedPnL": "-0.0063528",
					"leverage": 5,
					"active": true,
					"isTakenOver": false,
					"takeOverPrice": "0",
					"createdAt": 1780563600000,
					"updatedAt": 1780563616021
				}]
			}
		}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	positions, err := c.PerpsPositions(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("PerpsPositions: %v", err)
	}
	if len(positions.Positions) != 1 || positions.BlockHeight != 161169726 {
		t.Fatalf("unexpected position snapshot: %+v", positions)
	}

	p := positions.Positions[0]
	if p.ID != 6279292 || p.Symbol != "BTC-USD" || p.PositionSide != "BOTH" {
		t.Fatalf("unexpected position identity: %+v", p)
	}
	if p.Size != "0.00025" {
		t.Fatalf("size = %q, want 0.00025", p.Size)
	}
	if p.InitialMargin != "3.1764" {
		t.Fatalf("initialMargin = %q, want 3.1764", p.InitialMargin)
	}
	if p.AvgEntryPrice != "63528" {
		t.Fatalf("avgEntryPrice = %q, want 63528", p.AvgEntryPrice)
	}
	if p.RealizedPnL != "-0.0063528" || p.Leverage != 5 || !p.Active {
		t.Fatalf("unexpected position values: %+v", p)
	}
}
