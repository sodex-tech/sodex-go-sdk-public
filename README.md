# sodex-go-sdk-public

Official Go SDK for the Sodex exchange. Provides:

- **REST client** — a ready-to-use HTTP client for market data and authenticated trading across the **Spark** (spot) and **Bolt** (perpetuals) engines.
- **WebSocket client** — an auto-reconnecting subscriber for real-time market data and account updates.
- **EIP-712 signing** — low-level signing primitives for advanced users who want to sign requests without using the REST client.

## Requirements

- Go 1.24+

## Installation

```bash
go get github.com/sodex-tech/sodex-go-sdk-public
```

## Quickstart

### Market data (no key required)

```go
import (
    "context"
    "fmt"
    "log"

    "github.com/sodex-tech/sodex-go-sdk-public/client"
)

func main() {
    c := client.New(client.Config{BaseURL: client.TestnetBaseURL})

    tickers, err := c.PerpsTickers(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    for _, t := range tickers[:3] {
        fmt.Printf("%s last=%s vol=%s\n", t.Symbol, t.LastPrice, t.Volume)
    }
}
```

### Authenticated trading

```go
import (
    "context"
    "log"

    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"

    "github.com/sodex-tech/sodex-go-sdk-public/client"
    "github.com/sodex-tech/sodex-go-sdk-public/common/enums"
)

func main() {
    pk, err := crypto.HexToECDSA("your-private-key-hex")
    if err != nil {
        log.Fatal(err)
    }

    c := client.New(client.Config{
        BaseURL:    client.TestnetBaseURL,
        ChainID:    client.TestnetChainID,
        PrivateKey: pk,
    })

    // One-call helper for a single limit order.
    res, err := c.PlacePerpsLimitOrder(
        context.Background(),
        /* accountID */ 1001,
        /* symbolID  */ 1,
        /* clOrdID   */ "my-order-001",
        enums.OrderSideBuy,
        enums.PositionSideLong,
        enums.TimeInForceGTC,
        decimal.NewFromFloat(50000.0),
        decimal.NewFromFloat(0.01),
        /* reduceOnly */ false,
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("placed: %+v", res)
}
```

### WebSocket subscription

```go
import (
    "context"
    "log"

    "github.com/sodex-tech/sodex-go-sdk-public/ws"
)

func main() {
    w, err := ws.NewClient(client.TestnetBaseURL, "perps")
    if err != nil {
        log.Fatal(err)
    }

    _, err = w.Subscribe(
        ws.SubscribeParams{Channel: ws.ChannelTrade, Symbol: "BTC-USD"},
        func(push ws.Push) { log.Printf("%s %s", push.Channel, string(push.Data)) },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Connect blocks until the context is cancelled.
    log.Fatal(w.Connect(context.Background()))
}
```

More complete examples live in [`examples/`](./examples):

| Path | Shows |
|---|---|
| [`examples/rest/trade`](./examples/rest/trade) | Place + cancel a perps limit order |
| [`examples/rest/account`](./examples/rest/account) | Query balances, orders, positions |
| [`examples/ws/subscribe`](./examples/ws/subscribe) | Subscribe to trades + order book |
| [`examples/signer`](./examples/signer) | Low-level EIP-712 signing only |

## Packages

```
sodex-go-sdk-public/
├── client/          # REST client (perps + spot, market data + trading)
├── ws/              # WebSocket client (auto-reconnect, channel routing)
├── common/
│   ├── enums/       # Shared enums (OrderSide, OrderType, TimeInForce, …)
│   ├── types/       # Shared request types (Transfer, Replace, ScheduleCancel)
│   └── signer/      # Engine-agnostic EVMSigner core
├── perps/
│   ├── types/       # Bolt-specific request types
│   └── signer/      # Bolt signer — "futures" EIP-712 domain
├── spot/
│   ├── types/       # Spark-specific request types
│   └── signer/      # Spark signer — "spot" EIP-712 domain
├── cmd/sodex/       # CLI tool (built on top of the SDK)
└── examples/        # Runnable end-to-end examples
```

Most users only need `client/`, `ws/`, and `common/enums`. The `perps/`, `spot/`, and `common/types` packages are only needed when building custom order request structs by hand.

## Configuration

Exported constants in `client/`:

| Constant | Value |
|---|---|
| `client.DefaultBaseURL` | `https://mainnet-gw.sodex.dev` |
| `client.TestnetBaseURL` | `https://testnet-gw.sodex.dev` |
| `client.DefaultChainID` | `286623` (mainnet) |
| `client.TestnetChainID` | `138565` (testnet) |

`client.Config` fields:

- `BaseURL` — API root. Defaults to mainnet if empty.
- `ChainID` — EVM chain ID for EIP-712 domain separation. Defaults to mainnet.
- `PrivateKey` — `*ecdsa.PrivateKey`. Leave nil for read-only access.
- `APIKeyName` — optional API key name (sets `X-API-Key` on signed requests). Empty = master-wallet auth.
- `HTTPClient` — optional custom `*http.Client`. Defaults to a 30-second-timeout client.

The client tracks a strictly-monotonic millisecond nonce internally, so callers never manage nonces when using the REST client.

## Advanced: low-level signing

Every authenticated action sent to the Sodex exchange must carry an EIP-712 signature. The `client` package handles this automatically, but the `spot/signer` and `perps/signer` packages expose the primitives directly for callers who build their own HTTP layer.

### Signing pipeline

```
ActionPayload{type, params}
  └─▶ JSON-encode ──▶ keccak256 ──▶ payloadHash

ExchangeAction{payloadHash, nonce}
  └─▶ EIP-712 StructHash
        └─▶ keccak256(0x19 0x01 | domainSeparator | structHash) ──▶ digest

crypto.Sign(digest, privateKey)
  └─▶ [SignatureType byte | 65-byte ECDSA sig]  (66 bytes total)
```

Each engine uses its own EIP-712 domain (name `"spot"` for Spark, `"futures"` for Bolt), so a signature produced for one engine is cryptographically invalid on the other.

### Wire format

Every signature returned by the SDK is exactly **66 bytes**:

| Offset  | Length | Description                               |
|---------|--------|-------------------------------------------|
| `[0]`   | 1      | `SignatureType` — always `0x01` (EIP-712) |
| `[1:66]`| 65     | ECDSA signature: `r ‖ s ‖ v`             |

### Spot (Spark engine)

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"

    ssigner "github.com/sodex-tech/sodex-go-sdk-public/spot/signer"
    stypes  "github.com/sodex-tech/sodex-go-sdk-public/spot/types"
    "github.com/sodex-tech/sodex-go-sdk-public/common/enums"
)

privateKey, _ := crypto.HexToECDSA("your-private-key-hex")
s := ssigner.NewSigner(286623, privateKey)

price := decimal.NewFromFloat(50000.0)
qty := decimal.NewFromFloat(0.1)
req := &stypes.BatchNewOrderRequest{
    AccountID: 1001,
    Orders: []*stypes.BatchNewOrderItem{{
        SymbolID:    42,
        ClOrdID:     "order-001",
        Side:        enums.OrderSideBuy,
        Type:        enums.OrderTypeLimit,
        TimeInForce: enums.TimeInForceGTC,
        Price:       &price,
        Quantity:    &qty,
    }},
}

sig, err := s.SignBatchNewOrderRequest(req, nonce)
// Attach sig to the HTTP request as the signature header.
```

### Perps (Bolt engine)

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"

    psigner "github.com/sodex-tech/sodex-go-sdk-public/perps/signer"
    ptypes  "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
    "github.com/sodex-tech/sodex-go-sdk-public/common/enums"
)

privateKey, _ := crypto.HexToECDSA("your-private-key-hex")
s := psigner.NewSigner(286623, privateKey)

price := decimal.NewFromFloat(50000.0)
qty := decimal.NewFromFloat(1.0)
req := &ptypes.NewOrderRequest{
    AccountID: 1001,
    SymbolID:  101,
    Orders: []*ptypes.RawOrder{{
        ClOrdID:      "perp-001",
        Side:         enums.OrderSideBuy,
        Type:         enums.OrderTypeLimit,
        TimeInForce:  enums.TimeInForceGTC,
        Price:        &price,
        Quantity:     &qty,
        PositionSide: enums.PositionSideLong,
    }},
}

sig, err := s.SignNewOrderRequest(req, nonce)
```

### Supported sign actions

**Common** (both engines)

| Method                      | Request type            |
|-----------------------------|-------------------------|
| `SignTransferAssetRequest`  | `TransferAssetRequest`  |
| `SignReplaceOrderRequest`   | `ReplaceOrderRequest`   |
| `SignScheduleCancelRequest` | `ScheduleCancelRequest` |

**Spot (Spark)**

| Method                        | Request type              |
|-------------------------------|---------------------------|
| `SignBatchNewOrderRequest`    | `BatchNewOrderRequest`    |
| `SignBatchCancelOrderRequest` | `BatchCancelOrderRequest` |

**Perps (Bolt)**

| Method                      | Request type            |
|-----------------------------|-------------------------|
| `SignNewOrderRequest`       | `NewOrderRequest`       |
| `SignCancelOrderRequest`    | `CancelOrderRequest`    |
| `SignModifyOrderRequest`    | `ModifyOrderRequest`    |
| `SignUpdateLeverageRequest` | `UpdateLeverageRequest` |
| `SignUpdateMarginRequest`   | `UpdateMarginRequest`   |

### Nonce

The nonce is a millisecond unix timestamp, strictly monotonic per account per engine. The exchange rejects any nonce already consumed or outside the window `(now − 2 days, now + 1 day)`. The `client` package tracks the nonce automatically; when using the low-level signers directly, callers must supply it.

## Security

- **Cross-engine replay protection** — the EIP-712 domain encodes the engine name (`"spot"` or `"futures"`), so a spot signature cannot be accepted by the perps engine.
- **Session replay protection** — the nonce field in every `ExchangeAction` ensures a captured signature cannot be re-submitted.
- **Key handling** — private keys are passed by the caller on every `Sign*` call and are never stored inside the SDK.
