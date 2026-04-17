# Sodex WebSocket API

Real-time streaming data. Market data channels need no auth. Account channels need a wallet address (no signing).

## Endpoints

| Network | Spot | Perps |
|---------|------|-------|
| Mainnet | `wss://mainnet-gw.sodex.dev/ws/spot` | `wss://mainnet-gw.sodex.dev/ws/perps` |
| Testnet | `wss://testnet-gw.sodex.dev/ws/spot` | `wss://testnet-gw.sodex.dev/ws/perps` |

## CLI

```bash
sodex subscribe trades perps BTC-USD                 # Live trades
sodex subscribe orderbook perps BTC-USD --depth 10   # Live order book
sodex subscribe candle perps BTC-USD --interval 1m   # Live candles
sodex subscribe bbo perps BTC-USD                    # Best bid/offer
sodex subscribe mark-price BTC-USD                   # Mark price
sodex subscribe ticker perps                         # All tickers live
sodex subscribe order-updates perps --address 0x...  # Order status changes
sodex subscribe fills perps --address 0x...          # Trade fills
```

Use `--format json` for structured output. All commands block until Ctrl+C.

## Protocol

### Subscribe / Unsubscribe

```json
{"op":"subscribe","params":{"channel":"<channel>","symbol":"BTC-USD"}}
{"op":"unsubscribe","params":{"channel":"<channel>","symbol":"BTC-USD"}}
```

### Push Messages

```json
{"channel":"trade","type":"snapshot|update","data":[...]}
```

### Keep-alive

Send `{"op":"ping"}` every 30 seconds. Server responds `{"op":"pong"}`. Connection drops after 60s of silence.

## Market Data Channels (no auth)

| Channel | Params | Key fields |
|---------|--------|------------|
| `trade` | `symbol` | s, p (price), q (qty), S (side), T (time) |
| `l2Book` | `symbol`, `level` (opt) | a (asks), b (bids) as [[price, qty], ...] |
| `candle` | `symbol`, `interval` | o, h, l, c, v, i (interval), x (closed) |
| `ticker` | `symbols` (array) | c (close), h, l, v, P (change%) |
| `allTicker` | — | Same as ticker, all symbols |
| `bookTicker` | `symbols` (array) | a (ask), A (ask qty), b (bid), B (bid qty) |
| `markPrice` | `symbol` (perps) | p (mark), i (index), r (funding), oi (OI) |
| `allMarkPrice` | — (perps) | Same, all symbols |

Candle intervals: `1m`, `3m`, `5m`, `15m`, `30m`, `1h`, `2h`, `4h`, `6h`, `8h`, `12h`, `1D`, `3D`, `1W`, `1M`.
Book levels: 10, 20, 100, 500, 1000.

## Account Channels (need address, no signing)

| Channel | Params | Key fields |
|---------|--------|------------|
| `accountOrderUpdate` | `user` (address) | i (orderID), X (status), p, q, z (filled), x (execType) |
| `accountTrade` | `user` (address) | p, q, f (fee), m (isMaker), s (symbol) |
| `accountUpdate` | `user` (address) | Balance changes |
| `accountState` | `user` (address) | Full state snapshot |

## Examples

```json
{"op":"subscribe","params":{"channel":"trade","symbol":"BTC-USD"}}
{"op":"subscribe","params":{"channel":"l2Book","symbol":"BTC-USD","level":10}}
{"op":"subscribe","params":{"channel":"candle","symbol":"BTC-USD","interval":"5m"}}
{"op":"subscribe","params":{"channel":"ticker","symbols":["BTC-USD","ETH-USD"]}}
{"op":"subscribe","params":{"channel":"accountOrderUpdate","user":"0xYourAddress"}}
```

## Rate Limits

| Limit | Value |
|-------|-------|
| Connections per IP | 10 |
| New connections per IP per minute | 30 |
| Subscriptions per IP | 1000 |
| Messages per IP per minute | 2000 |
