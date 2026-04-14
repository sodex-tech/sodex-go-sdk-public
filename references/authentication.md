# Sodex Authentication & Signing

> **Recommended: Use the `sodex` CLI to avoid implementing signing manually.** This reference is only needed if you are calling the REST trading API directly from code.

## Required Headers

| Header | Value | Required? |
|--------|-------|-----------|
| `Content-Type` | `application/json` | Always |
| `X-API-Sign` | `0x01` + ECDSA signature hex (132 chars total) | Always |
| `X-API-Nonce` | Millisecond timestamp, monotonically increasing | Always |
| `X-API-Key` | API key name (e.g. `"mm-bot-key-01"`) | **YES if using API key** |
| `X-API-Chain` | Chain ID (286623 mainnet, 138565 testnet) | Recommended |

> **⚠️ #1 mistake: forgetting `X-API-Key` header.** If you sign with an API key's private key but omit the `X-API-Key` header, the server cannot identify which account to route to and signature verification will fail. This is the most common cause of authentication errors. Only omit this header if you are signing with the master wallet key directly (not recommended for bots).

## Signing Process

### Step 1: Build Signing Payload

Wrap the request body in a type/params envelope:

```json
{"type":"<action>","params":{...request body...}}
```

Action types: `newOrder`, `cancelOrder`, `transferAsset`, `updateLeverage`, `updateMargin`, `scheduleCancel`.

### Step 2: Compute payloadHash

```
payloadHash = keccak256(compact_json(signing_payload))
```

### Step 3: Sign EIP-712

```
Domain: {
  name: "spot" | "futures",
  version: "1",
  chainId: 286623 | 138565,
  verifyingContract: 0x0000000000000000000000000000000000000000
}

Type: ExchangeAction(bytes32 payloadHash, uint64 nonce)
```

Use `"spot"` for Spark engine endpoints, `"futures"` for Bolt engine endpoints.

### Step 4: Construct X-API-Sign

Prepend byte `0x01` to the 65-byte ECDSA signature:

```
X-API-Sign: 0x01<65-byte-signature-hex>
```

Total length: 132 characters (0x + 01 + 128 hex chars).

## Critical Rules

1. **`X-API-Key` header is required when using API keys.** Omitting it is the #1 auth error. Only omit if signing with master wallet key.
2. **Compact JSON** — no whitespace or newlines. Use `JSON.stringify()` or `json.Marshal()`.
3. **Key order must match Go struct field order** — the server re-marshals via `json.Marshal` to verify. Refer to [sodex-go-sdk](https://github.com/sodex-tech/sodex-go-sdk-public) struct definitions.
4. **DecimalString fields are strings** — `"price":"95000"` not `"price":95000`.
5. **`omitempty` fields omitted when unset** — optional pointer fields must not appear. Non-optional fields (`modifier`, `reduceOnly`, `positionSide`) must always be present.
6. **HTTP body = params only** — the request body does not include the `type` wrapper. Same field order as signing payload.
7. **Signing domain must match endpoint** — use `"futures"` for perps, `"spot"` for spot. Wrong domain = rejected.
8. **v-byte conversion** — after ECDSA signing, convert v from 27/28 to 0/1, then prepend `0x01`.
9. **All APIs reject trailing zeros** — `"0.4060"` fails, `"0.406"` works. Use `parseFloat(x).toString()`.

## Nonce Rules

- Must be uint64 millisecond timestamp
- Accepted range: `(now - 2 days, now + 1 day)`
- Must be strictly increasing per API key
- 100 highest nonces stored per address; new nonces must exceed the smallest

## API Keys

- Each master account supports up to 5 API keys
- Create and revoke keys via the Sodex web UI at https://sodex.com/
- API keys are EVM addresses; sign with the API key's private key

## Signing Example

Perps market buy — signing payload:
```json
{"type":"newOrder","params":{"accountID":12345,"symbolID":1,"orders":[{"clOrdID":"my-order-1","modifier":1,"side":1,"type":2,"timeInForce":3,"quantity":"0.001","reduceOnly":false,"positionSide":1}]}}
```

HTTP request body (params only):
```json
{"accountID":12345,"symbolID":1,"orders":[{"clOrdID":"my-order-1","modifier":1,"side":1,"type":2,"timeInForce":3,"quantity":"0.001","reduceOnly":false,"positionSide":1}]}
```

## Perps Order Field Order

Fields must appear in this exact order:

```
clOrdID, modifier, side, type, timeInForce, price, quantity, funds,
stopPrice, stopType, triggerType, reduceOnly, positionSide
```

Omit `omitempty` fields (`price`, `quantity`, `funds`, `stopPrice`, `stopType`, `triggerType`) when not set.
