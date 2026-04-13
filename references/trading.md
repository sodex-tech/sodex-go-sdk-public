# Sodex Trading

**Authentication required.** All trading commands need a private key and account ID.

## Security

- **Use API keys for bots.** Create them in the Sodex web UI. Never use your master key.
- **Mainnet trades require CONFIRM.** Always show a confirmation card and wait for user approval.
- **Use testnet first** (`--testnet`).

## CLI (recommended)

The CLI handles signing, nonces, and symbol resolution automatically.

```bash
# Setup
sodex account-id 0xYourAddress                     # Find account ID

# Place orders
sodex orders place spot --symbol BTC/USDC --side buy --price 95000 --qty 0.01
sodex orders place perps --symbol BTC-USD --side buy --type market --qty 0.1
sodex orders place perps --symbol BTC-USD --side sell --type market --qty 0.1 --reduce-only

# Post-only order
sodex orders place perps --symbol BTC-USD --side buy --price 70000 --qty 0.1 --tif gtx

# Cancel
sodex orders cancel spot --symbol-id 1 --order-id <OID> --cl-ord-id <CLOID>

# Leverage
sodex leverage BTC-USD 10
sodex leverage ETH-USD 5 --mode cross

# Transfer between accounts
sodex transfer --from 1001 --to 1002 --coin 0 --amount 100
```

Auth flags (cross-platform): `--private-key 0x<key> --api-key <name> --account-id <ID> --testnet`

Time-in-force: `gtc` (default), `ioc`, `fok`, `gtx` (post-only).

**⚠️ GTX (PostOnly) warning:** On Sodex, GTX orders may be partially filled before the remaining quantity is rejected. Check fill status after placement.

## REST Endpoints

Prefix with `/api/v1/spot` or `/api/v1/perps`. Requires signed headers (see `references/authentication.md`).

| Endpoint | Method | Signing action type | Description |
|----------|--------|---------------------|-------------|
| `/trade/orders` | POST | `newOrder` | Place order(s) |
| `/trade/orders/batch` | POST | `batchNewOrder` | Batch place (spot only) |
| `/trade/orders/replace` | POST | `replaceOrder` | Modify existing orders |
| `/trade/orders` | DELETE | `cancelOrder` | Cancel order(s) |
| `/trade/orders/batch` | DELETE | `batchCancelOrder` | Batch cancel (spot only) |
| `/accounts/transfers` | POST | `transferAsset` | Transfer between accounts |
| `/trade/leverage` | POST | `updateLeverage` | Update leverage (perps) |
| `/trade/margin` | POST | `updateMargin` | Update isolated margin (perps) |

## Request Bodies

> **⚠️ Trailing zeros:** All Sodex APIs reject price/quantity strings with trailing zeros. `"0.4060"` → rejected, `"0.406"` → accepted. Always strip: use `parseFloat(x).toString()` (JS) or equivalent.

> **⚠️ clOrdID:** Must be unique among open orders per account. Regex: `^[0-9a-zA-Z_-]{1,36}$`.

**Place Spot Order** — `POST /api/v1/spot/trade/orders`
```json
{"accountID":43933,"symbolID":1,"clOrdID":"my-001","side":1,"type":1,"timeInForce":1,"price":"95000","quantity":"0.01"}
```

**Place Perps Order** — `POST /api/v1/perps/trade/orders`
```json
{"accountID":43933,"symbolID":1,"orders":[{"clOrdID":"my-001","modifier":1,"side":1,"type":1,"timeInForce":1,"price":"70000","quantity":"0.1","reduceOnly":false,"positionSide":1}]}
```

**Cancel Spot** — `DELETE /api/v1/spot/trade/orders`
```json
{"accountID":43933,"symbolID":1,"clOrdID":"cancel-001","orderID":123456}
```

**Cancel Perps** — `DELETE /api/v1/perps/trade/orders`
```json
{"accountID":43933,"cancels":[{"symbolID":1,"orderID":123456}]}
```

**Leverage** — `POST /api/v1/perps/trade/leverage`
```json
{"accountID":43933,"symbolID":1,"leverage":10,"marginMode":1}
```

**Transfer (between sub-accounts)** — `POST /api/v1/spot/accounts/transfers`
```json
{"id":1,"fromAccountID":1001,"toAccountID":1002,"coinID":0,"amount":"100","type":4}
```

**Transfer (between perps and spot)** — uses magic `toAccountID = 999`:

| Direction | Endpoint | Transfer type | Signing domain |
|-----------|----------|---------------|----------------|
| Perps → Spot | `/api/v1/perps/accounts/transfers` | `SPOT_WITHDRAW` (5) | `"futures"` |
| Spot → Perps | `/api/v1/spot/accounts/transfers` | `PERPS_WITHDRAW` (3) | `"spot"` |

```json
{"id":1,"fromAccountID":43933,"toAccountID":999,"coinID":0,"amount":"100","type":5}
```

## Rate Limits

- Orders: 1200/minute per (account, API key)
- Max open orders: 1000 (spot + perps combined)
- Batch weight: `1 + floor(N/40)`

## Error Messages

| Error | Resolution |
|-------|------------|
| `"not authenticated"` | Set `--private-key` or `SODEX_PRIVATE_KEY` |
| `"symbol X not found"` | Run `sodex markets <engine>` for valid names |
| `"sodex API error (code 401)"` | Check: 1) `X-API-Key` header present? 2) signing domain matches endpoint? 3) private key matches registered API key? |
| `"sodex API error (code 400)"` | Check: trailing zeros in price/qty? clOrdID unique? field order correct? |
| `"sodex API error (code 429)"` | Rate limited — back off |
| `"nonce is invalid"` | Nonce outside (now-2d, now+1d) window or already used |
| Signature rejected silently | Wrong domain (`"futures"` vs `"spot"`), missing `X-API-Key`, or v-byte not converted (27→0) |
