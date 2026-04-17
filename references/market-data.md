# Sodex Market Data

No authentication required. All endpoints are public.

## CLI

```bash
sodex markets [spot|perps]                          # List all trading pairs
sodex markets [spot|perps] --format json            # JSON output
sodex tickers [spot|perps]                          # 24h stats
sodex orderbook [spot|perps] SYMBOL --depth N       # Order book snapshot
```

## REST Endpoints

Prefix with `/api/v1/spot` or `/api/v1/perps`.

| Endpoint | Description |
|----------|-------------|
| `GET /markets/symbols` | All trading pairs with rules, symbolID, tick/step size, fees |
| `GET /markets/tickers` | 24h stats: lastPx, highPx, lowPx, volume, changePct. Perps adds markPrice, indexPrice, fundingRate, openInterest |
| `GET /markets/bookTickers` | Best bid/ask for all symbols |
| `GET /markets/{symbol}/orderbook?limit=20` | Order book depth. Returns bids/asks as `[[price, qty], ...]` |
| `GET /markets/{symbol}/klines?interval=1m&limit=100` | OHLCV candles. Optional: `startTime`, `endTime` (ms) |
| `GET /markets/{symbol}/trades?limit=50` | Recent public trades (max 500) |
| `GET /markets/mark-prices` | Mark prices (perps only) |

`{symbol}` is the internal name: `vBTC_vUSDC` (spot), `BTC-USD` (perps).

Candle intervals: `1m`, `3m`, `5m`, `15m`, `30m`, `1h`, `2h`, `4h`, `6h`, `8h`, `12h`, `1D`, `3D`, `1W`, `1M`.

## Rate Limits

| Endpoint | Weight |
|----------|--------|
| symbols, tickers, bookTickers | 2 |
| orderbook (depth ≤100) | 5 |
| orderbook (101–500) | 10 |
| klines, trades | 20 |

Total budget: 1200 weight/minute/IP.

## Examples

### CLI (preferred for agents)

```bash
sodex tickers perps                                # All perps 24h stats
sodex tickers perps --format json                  # JSON for programmatic use
sodex tickers spot                                 # All spot 24h stats
sodex orderbook perps BTC-USD --depth 10           # BTC-USD order book
sodex markets perps --format json                  # List all perps pairs
```

### REST API

```bash
# Get BTC-USD order book on perps
curl -s "https://mainnet-gw.sodex.dev/api/v1/perps/markets/BTC-USD/orderbook?limit=10"

# Get 1h candles for spot BTC
curl -s "https://mainnet-gw.sodex.dev/api/v1/spot/markets/vBTC_vUSDC/klines?interval=1h&limit=50"

# Get all perps tickers
curl -s "https://mainnet-gw.sodex.dev/api/v1/perps/markets/tickers"
```
