package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	rpctypes "github.com/sodex-tech/sodex-go-sdk-public/client/types"
	"github.com/sodex-tech/sodex-go-sdk-public/common/enums"
	ctypes "github.com/sodex-tech/sodex-go-sdk-public/common/types"
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

// TestAdditionalTradingResponses checks the remaining signed methods decode order results and accept null success data.
func TestAdditionalTradingResponses(t *testing.T) {
	modifyCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Sign") == "" {
			t.Error("missing signature header")
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /api/v1/spot/trade/orders/replace":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"code":0,"clOrdID":"spot-replace","orderID":51}]}`))
		case "POST /api/v1/perps/trade/orders":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"code":0,"clOrdID":"perps-new","orderID":52},{"code":-1,"clOrdID":"perps-rejected","error":"invalid price"}]}`))
		case "POST /api/v1/perps/trade/orders/modify":
			modifyCalls++
			if modifyCalls == 1 {
				_, _ = w.Write([]byte(`{"code":0,"data":{"code":0}}`))
			} else {
				_, _ = w.Write([]byte(`{"code":-1,"error":"order not found"}`))
			}
		case "POST /api/v1/perps/trade/orders/replace":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"code":0,"clOrdID":"perps-replace","orderID":53}]}`))
		case "POST /api/v1/perps/trade/margin", "POST /api/v1/spot/trade/orders/schedule-cancel", "POST /api/v1/perps/trade/orders/schedule-cancel":
			_, _ = w.Write([]byte(`{"code":0,"data":null}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	c := New(Config{BaseURL: srv.URL, PrivateKey: key, HTTPClient: srv.Client()})
	ctx := context.Background()
	orderID := uint64(42)
	price := decimal.NewFromInt(100)
	replace := &ctypes.ReplaceOrderRequest{AccountID: 1, Orders: []*ctypes.ReplaceParams{{SymbolID: 2, ClOrdID: "spot-replace", OrigOrderID: &orderID, Price: &price}}}

	spotReplaced, err := c.ReplaceSpotOrders(ctx, replace)
	if err != nil || len(spotReplaced) != 1 || spotReplaced[0].OrderID == nil || *spotReplaced[0].OrderID != 51 {
		t.Fatalf("spot replace = %+v, error = %v", spotReplaced, err)
	}
	placed, err := c.PlacePerpsOrder(ctx, &ptypes.NewOrderRequest{AccountID: 1, SymbolID: 2, Orders: []*ptypes.RawOrder{
		{ClOrdID: "perps-new", Side: enums.OrderSideBuy, Type: enums.OrderTypeLimit, TimeInForce: enums.TimeInForceGTC, Price: &price},
		{ClOrdID: "perps-rejected", Side: enums.OrderSideBuy, Type: enums.OrderTypeLimit, TimeInForce: enums.TimeInForceGTC, Price: &price},
	}})
	if err != nil || len(placed) != 2 || placed[0].OrderID == nil || *placed[0].OrderID != 52 || placed[1].Code != -1 || placed[1].Error == nil || *placed[1].Error != "invalid price" {
		t.Fatalf("perps place = %+v, error = %v", placed, err)
	}
	modify := &ptypes.ModifyOrderRequest{AccountID: 1, SymbolID: 2, OrderID: &orderID, Price: &price}
	modified, err := c.ModifyPerpsOrder(ctx, modify)
	if err != nil || modified == nil || modified.Code != 0 {
		t.Fatalf("perps modify = %+v, error = %v", modified, err)
	}
	var apiErr *ErrAPI
	if _, err := c.ModifyPerpsOrder(ctx, modify); !errors.As(err, &apiErr) || apiErr.Message != "order not found" {
		t.Fatalf("perps modify error = %v", err)
	}
	replace.Orders[0].ClOrdID = "perps-replace"
	perpsReplaced, err := c.ReplacePerpsOrders(ctx, replace)
	if err != nil || len(perpsReplaced) != 1 || perpsReplaced[0].OrderID == nil || *perpsReplaced[0].OrderID != 53 {
		t.Fatalf("perps replace = %+v, error = %v", perpsReplaced, err)
	}
	if err := c.UpdateMargin(ctx, &ptypes.UpdateMarginRequest{AccountID: 1, SymbolID: 2, Amount: decimal.NewFromInt(1)}); err != nil {
		t.Fatalf("update margin: %v", err)
	}
	schedule := &ctypes.ScheduleCancelRequest{AccountID: 1}
	if err := c.ScheduleSpotCancel(ctx, schedule); err != nil {
		t.Fatalf("schedule spot cancel: %v", err)
	}
	if err := c.SchedulePerpsCancel(ctx, schedule); err != nil {
		t.Fatalf("schedule perps cancel: %v", err)
	}
}

// TestTransferReceipts checks both signed transfer endpoints expose the Gateway transfer ID.
func TestTransferReceipts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || (r.URL.Path != "/api/v1/spot/accounts/transfers" && r.URL.Path != "/api/v1/perps/accounts/transfers") {
			t.Errorf("unexpected transfer request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"id":73}}`))
	}))
	defer srv.Close()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	c := New(Config{BaseURL: srv.URL, PrivateKey: key, HTTPClient: srv.Client()})
	req := &ctypes.TransferAssetRequest{ID: 73, FromAccountID: 1, ToAccountID: 2, CoinID: 0, Amount: decimal.NewFromInt(1), Type: enums.TransferAssetTypeInternal}
	for _, transfer := range []func(context.Context, *ctypes.TransferAssetRequest) (*rpctypes.TransferAssetResponse, error){c.SpotTransfer, c.PerpsTransfer} {
		result, err := transfer(context.Background(), req)
		if err != nil || result == nil || result.ID != 73 {
			t.Fatalf("transfer receipt = %+v, error = %v", result, err)
		}
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
	if balances.BlockHeight != 20 || len(balances.Balances) != 1 || balances.Balances[0].Collateral != "8" || balances.Balances[0].MarginRatio != "0.5" || balances.Balances[0].Price == nil || *balances.Balances[0].Price != "1" {
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
	for _, query := range []func() (*rpctypes.OrderBook, error){
		func() (*rpctypes.OrderBook, error) { return c.SpotOrderBook(context.Background(), "vBTC_vUSDC", 25) },
		func() (*rpctypes.OrderBook, error) { return c.PerpsOrderBook(context.Background(), "BTC-USD", 25) },
	} {
		book, err := query()
		if err != nil {
			t.Fatal(err)
		}
		if book.BlockTime != 100 || book.BlockHeight != 20 || book.UpdateID != 7 || len(book.Bids) != 1 || book.Bids[0][0] != "1" {
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

// TestMarketResponseFields checks spot/perps rule and ticker fields that the gateway includes in public market responses.
func TestMarketResponseFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/spot/markets/symbols":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"id":1,"name":"vBTC_vUSDC","baseCoinID":0,"baseCoinPrecision":8,"quoteCoinID":2,"quoteCoinPrecision":6,"marketMinQuantity":"0.01","marketMaxQuantity":"100","maxNotional":"100000","buyLimitUpRatio":"0.1"}]}`))
		case "/api/v1/perps/markets/symbols":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"id":3,"name":"BTC-USD","quoteCoinID":2,"quoteCoinPrecision":6,"marketMinQuantity":"0.001","marketMaxQuantity":"1000","maxNotional":"125000","openInterestCap":"1000000","maxLeverage":10,"initLeverage":5,"marginTiers":[{"maxNotionalValue":"125000","maintenanceMarginRate":"0.05","maxLeverage":10,"maintenanceDeduction":"0"}],"fundingInterval":3600,"interestRate":"0.0001"}]}`))
		case "/api/v1/perps/markets/tickers":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"symbol":"BTC-USD","lastPx":"100","lastSz":"0.5","vwap":"99.5","openTime":1000,"closeTime":2000,"nextFundingTime":3000,"markPrice":"100.1"}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	spot, err := c.SpotSymbols(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(spot) != 1 || spot[0].BaseCoinID != 0 || spot[0].MarketMinQty != "0.01" || spot[0].MaxNotional != "100000" || spot[0].BuyLimitUpRatio != "0.1" {
		t.Fatalf("spot symbols: %+v", spot)
	}
	perps, err := c.PerpsSymbols(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(perps) != 1 || perps[0].MaxLeverage != 10 || len(perps[0].MarginTiers) != 1 || perps[0].MarginTiers[0].MaintenanceMarginRate != "0.05" || perps[0].FundingInterval != 3600 || perps[0].OpenInterestCap != "1000000" {
		t.Fatalf("perps symbols: %+v", perps)
	}
	tickers, err := c.PerpsTickers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tickers) != 1 || tickers[0].LastSz == nil || *tickers[0].LastSz != "0.5" || tickers[0].VWAP == nil || *tickers[0].VWAP != "99.5" || tickers[0].OpenTime != 1000 || tickers[0].NextFundingTime != 3000 {
		t.Fatalf("perps tickers: %+v", tickers)
	}
}

// TestAccountResponseFields checks omitted spot order amounts and perps order controls survive response decoding.
func TestAccountResponseFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/spot/accounts/0xabc/orders":
			_, _ = w.Write([]byte(`{"code":0,"data":{"blockTime":100,"orders":[{"orderID":1,"type":"MARKET","funds":"100","builder":{"builderID":7,"feeRate":10}}]}}`))
		case "/api/v1/perps/accounts/0xabc/orders":
			_, _ = w.Write([]byte(`{"code":0,"data":{"blockTime":100,"orders":[{"orderID":2,"price":"100","origQty":"1","positionSide":"BOTH","reduceOnly":false,"stopPrice":"99","stopType":"STOP_LOSS","triggerType":"MARK_PRICE","positionID":3,"primaryOrderID":4,"attachedOrderIDs":[5,6]}]}}`))
		case "/api/v1/spot/accounts/0xabc/trades":
			_, _ = w.Write([]byte(`{"code":0,"data":[{"tradeID":9,"builderFee":"0.01"}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	spot, err := c.SpotOrders(context.Background(), "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	if spot.BlockTime != 100 || len(spot.Orders) != 1 || spot.Orders[0].Price != nil || spot.Orders[0].Quantity != nil || spot.Orders[0].Funds == nil || *spot.Orders[0].Funds != "100" || spot.Orders[0].Builder == nil || spot.Orders[0].Builder.BuilderID != 7 {
		t.Fatalf("spot orders: %+v", spot)
	}
	perps, err := c.PerpsOrders(context.Background(), "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	if perps.BlockTime != 100 || len(perps.Orders) != 1 || perps.Orders[0].Price == nil || *perps.Orders[0].Price != "100" || perps.Orders[0].ReduceOnly || perps.Orders[0].PositionSide != "BOTH" || perps.Orders[0].StopPrice == nil || *perps.Orders[0].StopPrice != "99" || perps.Orders[0].PositionID == nil || *perps.Orders[0].PositionID != 3 || len(perps.Orders[0].AttachedOrderIDs) != 2 {
		t.Fatalf("perps orders: %+v", perps)
	}
	trades, err := c.SpotUserTrades(context.Background(), "0xabc", HistoryFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 || trades[0].BuilderFee == nil || *trades[0].BuilderFee != "0.01" {
		t.Fatalf("spot trades: %+v", trades)
	}
}
