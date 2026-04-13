package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingInterval   = 30 * time.Second
	pongWait       = 30 * time.Second
	writeWait      = 10 * time.Second
	reconnectDelay = time.Second
)

// Handler is a callback invoked for each push message on a subscribed channel.
type Handler func(push Push)

// Client is a WebSocket client for the Sodex real-time API.
// It manages the connection lifecycle, automatic reconnection,
// and subscription routing.
type Client struct {
	url       string
	conn      *websocket.Conn
	mu        sync.Mutex
	subs      map[int64]*subscription
	handlers  map[string][]int64 // channel identifier → subscription IDs
	nextID    atomic.Int64
	onError   func(err error)
	done      chan struct{}
	closeOnce sync.Once
}

type subscription struct {
	id      int64
	params  SubscribeParams
	handler Handler
}

// NewClient creates a new WebSocket client.
// engine is "spot" or "perps". baseURL is the HTTP base URL (e.g. "https://testnet-gw.sodex.dev").
func NewClient(baseURL, engine string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("ws: invalid base URL: %w", err)
	}
	scheme := "wss"
	if u.Scheme == "http" {
		scheme = "ws"
	}
	wsURL := fmt.Sprintf("%s://%s/ws/%s", scheme, u.Host, engine)

	c := &Client{
		url:      wsURL,
		subs:     make(map[int64]*subscription),
		handlers: make(map[string][]int64),
		done:     make(chan struct{}),
	}
	return c, nil
}

// OnError sets an error callback for connection/read errors.
func (c *Client) OnError(fn func(error)) {
	c.mu.Lock()
	c.onError = fn
	c.mu.Unlock()
}

// Connect establishes the WebSocket connection and starts the read/ping loops.
// It blocks until the context is cancelled or Close is called.
// On disconnect, it automatically reconnects and re-subscribes.
func (c *Client) Connect(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return nil
		default:
		}

		if err := c.dial(ctx); err != nil {
			c.emitError(fmt.Errorf("ws: connect %s: %w", c.url, err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c.done:
				return nil
			case <-time.After(reconnectDelay):
				continue
			}
		}

		// Re-subscribe all existing subscriptions after reconnect.
		c.resubscribe()

		// Block on read loop until disconnect.
		c.readLoop(ctx)

		// Connection lost — close and retry.
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return nil
		case <-time.After(reconnectDelay):
		}
	}
}

func (c *Client) dial(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return err
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait + pingInterval))
	})
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

func (c *Client) resubscribe() {
	c.mu.Lock()
	subs := make([]*subscription, 0, len(c.subs))
	for _, s := range c.subs {
		subs = append(subs, s)
	}
	c.mu.Unlock()

	for _, s := range subs {
		if err := c.sendSubscribe("subscribe", s.id, s.params); err != nil {
			c.emitError(fmt.Errorf("ws: resubscribe %s: %w", s.params.Channel, err))
		}
	}
}

// Subscribe registers a handler for a channel and sends the subscribe message.
// Returns a subscription ID that can be used to Unsubscribe.
func (c *Client) Subscribe(params SubscribeParams, handler Handler) (int64, error) {
	id := c.nextID.Add(1)
	sub := &subscription{id: id, params: params, handler: handler}

	identifier := channelIdentifier(params)

	c.mu.Lock()
	c.subs[id] = sub
	c.handlers[identifier] = append(c.handlers[identifier], id)
	c.mu.Unlock()

	if err := c.sendSubscribe("subscribe", id, params); err != nil {
		return 0, err
	}
	return id, nil
}

// Unsubscribe removes a subscription by ID.
func (c *Client) Unsubscribe(id int64) error {
	c.mu.Lock()
	sub, ok := c.subs[id]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("ws: subscription %d not found", id)
	}
	delete(c.subs, id)

	identifier := channelIdentifier(sub.params)
	ids := c.handlers[identifier]
	for i, sid := range ids {
		if sid == id {
			c.handlers[identifier] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	// If no more handlers for this identifier, send unsubscribe to server.
	shouldUnsub := len(c.handlers[identifier]) == 0
	if shouldUnsub {
		delete(c.handlers, identifier)
	}
	c.mu.Unlock()

	if shouldUnsub {
		return c.sendSubscribe("unsubscribe", id, sub.params)
	}
	return nil
}

// Close terminates the WebSocket connection and stops reconnection.
func (c *Client) Close() error {
	c.closeOnce.Do(func() { close(c.done) })
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ── internal ─────────────────────────────────────────────────────────────────

func (c *Client) sendSubscribe(op string, id int64, params SubscribeParams) error {
	raw, _ := json.Marshal(params)
	msg := Request{Op: op, ID: id, Params: raw}
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return nil // will re-subscribe on reconnect
	}
	_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(msg)
}

func (c *Client) readLoop(ctx context.Context) {
	// Start ping ticker.
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-pingTicker.C:
				c.mu.Lock()
				conn := c.conn
				c.mu.Unlock()
				if conn == nil {
					return
				}
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteJSON(Request{Op: "ping"}); err != nil {
					return
				}
			case <-ctx.Done():
				return
			case <-c.done:
				return
			}
		}
	}()

	for {
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		if conn == nil {
			return
		}

		_ = conn.SetReadDeadline(time.Now().Add(pongWait + pingInterval))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			c.emitError(fmt.Errorf("ws: read: %w", err))
			return
		}

		c.dispatch(raw)
	}
}

func (c *Client) dispatch(raw []byte) {
	// Try to parse as a pong or ack response first.
	var resp Response
	if err := json.Unmarshal(raw, &resp); err == nil {
		if resp.Op == "pong" {
			return
		}
		if resp.Op == "error" {
			c.emitError(fmt.Errorf("ws: server error (code %s): %s", resp.Code, resp.Error))
			return
		}
		if resp.Op == "subscribe" || resp.Op == "unsubscribe" {
			if resp.Success != nil && !*resp.Success {
				c.emitError(fmt.Errorf("ws: %s failed: %s", resp.Op, resp.Error))
			}
			return
		}
	}

	// Parse as push message.
	var push Push
	if err := json.Unmarshal(raw, &push); err != nil || push.Channel == "" {
		return
	}

	// Route to handlers.
	identifier := push.Channel
	c.mu.Lock()
	ids := append([]int64(nil), c.handlers[identifier]...)
	c.mu.Unlock()

	for _, id := range ids {
		c.mu.Lock()
		sub, ok := c.subs[id]
		c.mu.Unlock()
		if ok {
			sub.handler(push)
		}
	}
}

func (c *Client) emitError(err error) {
	c.mu.Lock()
	fn := c.onError
	c.mu.Unlock()
	if fn != nil {
		fn(err)
	} else {
		log.Println(err)
	}
}

func channelIdentifier(p SubscribeParams) string {
	return p.Channel
}
