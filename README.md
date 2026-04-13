# sodex-go-sdk

Official Go SDK for the Sodex exchange. Provides EIP-712 request signing for both the **Spark** (spot) and **Bolt** (perpetuals) trading engines.

## Requirements

- Go 1.24+

## Installation

```bash
go get github.com/sodex-tech/sodex-go-sdk-public
```

## Overview

Every authenticated action sent to the Sodex exchange must carry an EIP-712 signature. The SDK handles all cryptographic details so that callers only need to:

1. Build a typed request struct.
2. Call the appropriate `Sign*` method with the current nonce.
3. Attach the returned 66-byte signature to the HTTP request header.

### Signing Pipeline

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

### Wire Format

Every signature returned by the SDK is exactly **66 bytes**:

| Offset  | Length | Description                               |
|---------|--------|-------------------------------------------|
| `[0]`   | 1      | `SignatureType` — always `0x01` (EIP-712) |
| `[1:66]`| 65     | ECDSA signature: `r ‖ s ‖ v`             |

## Package Layout

```
sodex-go-sdk/
├── common/
│   ├── enums/          # Shared enum types (OrderSide, OrderType, SignatureType, …)
│   ├── types/          # Core EIP-712 primitives and shared request types
│   └── signer/         # Engine-agnostic EVMSigner (signing & verification core)
├── spot/
│   ├── types/          # Spark-specific request types
│   └── signer/         # Spark signer — wraps EVMSigner with the "spot" domain
└── perps/
    ├── types/          # Bolt-specific request types
    └── signer/         # Bolt signer — wraps EVMSigner with the "futures" domain
```

## Usage

### Spot (Spark Engine)

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"

    ssigner "github.com/sodex-tech/sodex-go-sdk-public/spot/signer"
    stypes  "github.com/sodex-tech/sodex-go-sdk-public/spot/types"
    "github.com/sodex-tech/sodex-go-sdk-public/common/enums"
)

privateKey, err := crypto.HexToECDSA("your-private-key-hex")
if err != nil {
    log.Fatal(err)
}

s := ssigner.NewSigner(286623, privateKey) // chainID 286623

// Place a batch of limit buy orders
req := &stypes.BatchNewOrderRequest{
    AccountID: 1001,
    Orders: []*stypes.BatchNewOrderItem{
        {
            SymbolID:    42,
            ClOrdID:     "order-001",
            Side:        enums.OrderSideBuy,
            Type:        enums.OrderTypeLimit,
            TimeInForce: enums.TimeInForceGTC,
            Price:       decimalPtr(decimal.NewFromFloat(50000.0)),
            Quantity:    decimalPtr(decimal.NewFromFloat(0.1)),
        },
    },
}

sig, err := s.SignBatchNewOrderRequest(req, nonce)
// Attach sig to the HTTP request as the signature header.
```

### Perps (Bolt Engine)

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"

    psigner "github.com/sodex-tech/sodex-go-sdk-public/perps/signer"
    ptypes  "github.com/sodex-tech/sodex-go-sdk-public/perps/types"
    "github.com/sodex-tech/sodex-go-sdk-public/common/enums"
)

privateKey, err := crypto.HexToECDSA("your-private-key-hex")
if err != nil {
    log.Fatal(err)
}

s := psigner.NewSigner(286623, privateKey) // chainID 286623

// Place a perpetuals order
req := &ptypes.NewOrderRequest{
    AccountID: 1001,
    SymbolID:  101,
    Orders: []*ptypes.RawOrder{
        {
            ClOrdID:      "perp-001",
            Side:         enums.OrderSideBuy,
            Type:         enums.OrderTypeLimit,
            TimeInForce:  enums.TimeInForceGTC,
            Price:        decimalPtr(decimal.NewFromFloat(50000.0)),
            Quantity:     decimalPtr(decimal.NewFromFloat(1.0)),
            PositionSide: enums.PositionSideLong,
        },
    },
}

sig, err := s.SignNewOrderRequest(req, nonce)
```

## Supported Actions

### Common (available on both engines)

| Method                      | Request Type              | Description                        |
|-----------------------------|---------------------------|------------------------------------|
| `SignTransferAssetRequest`  | `TransferAssetRequest`    | Inter-account asset transfer       |
| `SignReplaceOrderRequest`   | `ReplaceOrderRequest`     | Batch order replacement            |
| `SignScheduleCancelRequest` | `ScheduleCancelRequest`   | Scheduled mass cancellation        |

### Spot (Spark engine only)

| Method                        | Request Type              | Description              |
|-------------------------------|---------------------------|--------------------------|
| `SignBatchNewOrderRequest`    | `BatchNewOrderRequest`    | Batch order placement    |
| `SignBatchCancelOrderRequest` | `BatchCancelOrderRequest` | Batch order cancellation |

### Perps (Bolt engine only)

| Method                     | Request Type            | Description                  |
|----------------------------|-------------------------|------------------------------|
| `SignNewOrderRequest`      | `NewOrderRequest`       | Order placement              |
| `SignCancelOrderRequest`   | `CancelOrderRequest`    | Order cancellation           |
| `SignUpdateLeverageRequest`| `UpdateLeverageRequest` | Position leverage adjustment |
| `SignUpdateMarginRequest`  | `UpdateMarginRequest`   | Position margin adjustment   |

## Nonce

The nonce is a monotonically increasing counter per account per engine. The exchange rejects any request whose nonce has already been consumed. Callers are responsible for tracking and incrementing the nonce; the SDK does not maintain any nonce state.

## Security

- **Cross-engine replay protection** — the EIP-712 domain encodes the engine name (`"spot"` or `"futures"`), so a spot signature cannot be accepted by the perps engine.
- **Session replay protection** — the nonce field in every `ExchangeAction` ensures that a captured signature cannot be re-submitted.
- **Key handling** — private keys are passed by the caller on every `Sign*` call and are never stored inside the SDK.
