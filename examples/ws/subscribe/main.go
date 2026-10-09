// Command subscribe demonstrates subscribing to multiple WebSocket channels
// and routing push messages through typed handlers.
//
// Usage:
//
//	go run ./examples/ws/subscribe
//
// Subscribes to trades and the ten-level order book for BTC-USD on the perps engine.
// Runs until the context is cancelled (Ctrl-C).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/sodex-tech/sodex-go-sdk-public/client"
	"github.com/sodex-tech/sodex-go-sdk-public/ws"
	wstypes "github.com/sodex-tech/sodex-go-sdk-public/ws/types"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	w, err := ws.NewClient(client.TestnetBaseURL, "perps")
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	w.OnError(func(err error) { log.Printf("ws error: %v", err) })

	// Subscribe BEFORE Connect — the SDK queues subscriptions and sends them
	// as soon as the socket is open, then automatically re-subscribes on reconnect.
	if _, err := w.Subscribe(
		ws.SubscribeParams{Channel: ws.ChannelTrade, Symbols: []string{"BTC-USD"}},
		handleTrade,
	); err != nil {
		log.Fatalf("Subscribe trade: %v", err)
	}

	if _, err := w.Subscribe(
		ws.SubscribeParams{Channel: ws.ChannelL4Book, Symbol: "BTC-USD", Level: 10},
		handleOrderBook,
	); err != nil {
		log.Fatalf("Subscribe l4Book: %v", err)
	}

	log.Println("connecting… (Ctrl-C to quit)")
	if err := w.Connect(ctx); err != nil && !errorsIsContextCancelled(err) {
		log.Fatalf("Connect: %v", err)
	}
}

func handleTrade(push ws.Push) {
	var trades []*wstypes.WsTrade
	if err := json.Unmarshal(push.Data, &trades); err != nil {
		log.Printf("decode trade: %v", err)
		return
	}
	for _, t := range trades {
		fmt.Printf("[trade]   %s %s @ %s qty=%s\n", t.Symbol, t.Side, t.Price, t.Quantity)
	}
}

func handleOrderBook(push ws.Push) {
	var symbol string
	var bids, asks [][]string
	if push.Type == "snapshot" {
		var book wstypes.WsDepthSnapshot
		if err := json.Unmarshal(push.Data, &book); err != nil {
			log.Printf("decode l4Book snapshot: %v", err)
			return
		}
		symbol, bids, asks = book.Symbol, book.Bids, book.Asks
	} else {
		var book wstypes.WsDepthUpdate
		if err := json.Unmarshal(push.Data, &book); err != nil {
			log.Printf("decode l4Book update: %v", err)
			return
		}
		symbol, bids, asks = book.Symbol, book.Bids, book.Asks
	}
	bestBid, bestAsk := "-", "-"
	if len(bids) > 0 {
		bestBid = fmt.Sprintf("%s × %s", bids[0][0], bids[0][1])
	}
	if len(asks) > 0 {
		bestAsk = fmt.Sprintf("%s × %s", asks[0][0], asks[0][1])
	}
	fmt.Printf("[%s] %s  bid %s  ask %s\n", push.Type, symbol, bestBid, bestAsk)
}

func errorsIsContextCancelled(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}
