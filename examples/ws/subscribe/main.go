// Command subscribe demonstrates subscribing to multiple WebSocket channels
// and routing push messages through typed handlers.
//
// Usage:
//
//	go run ./examples/ws/subscribe
//
// Subscribes to trades and the L2 order book for BTC-USD on the perps engine.
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
		ws.SubscribeParams{Channel: ws.ChannelTrade, Symbol: "BTC-USD"},
		handleTrade,
	); err != nil {
		log.Fatalf("Subscribe trade: %v", err)
	}

	if _, err := w.Subscribe(
		ws.SubscribeParams{Channel: ws.ChannelL2Book, Symbol: "BTC-USD", Level: 5},
		handleL2Book,
	); err != nil {
		log.Fatalf("Subscribe l2Book: %v", err)
	}

	log.Println("connecting… (Ctrl-C to quit)")
	if err := w.Connect(ctx); err != nil && !errorsIsContextCancelled(err) {
		log.Fatalf("Connect: %v", err)
	}
}

func handleTrade(push ws.Push) {
	var t ws.Trade
	if err := json.Unmarshal(push.Data, &t); err != nil {
		log.Printf("decode trade: %v", err)
		return
	}
	fmt.Printf("[trade]   %s %s @ %s qty=%s\n", t.Symbol, t.Side, t.Price, t.Quantity)
}

func handleL2Book(push ws.Push) {
	var book ws.L2Book
	if err := json.Unmarshal(push.Data, &book); err != nil {
		log.Printf("decode l2Book: %v", err)
		return
	}
	bestBid, bestAsk := "-", "-"
	if len(book.Bids) > 0 {
		bestBid = fmt.Sprintf("%s × %s", book.Bids[0][0], book.Bids[0][1])
	}
	if len(book.Asks) > 0 {
		bestAsk = fmt.Sprintf("%s × %s", book.Asks[0][0], book.Asks[0][1])
	}
	fmt.Printf("[%s] %s  bid %s  ask %s\n", push.Type, book.Symbol, bestBid, bestAsk)
}

func errorsIsContextCancelled(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}
