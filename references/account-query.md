# Sodex Account Queries

No authentication required — only a wallet address is needed.

## CLI

```bash
sodex account-id [ADDRESS]                          # Look up account ID + user ID
sodex balance [spot|perps] [ADDRESS]                # Asset balances
sodex orders list [spot|perps]                      # Open orders
sodex positions                                     # Perps positions (open)
```

All commands support `--format json` and `--testnet`.

## Find Your Account ID

The account ID is required for trading. Look it up from your wallet address:

```bash
sodex account-id 0xYourAddress
```

Output:
```
Address:    0x...
Account ID: 43933
User ID:    43933
```

Or via REST:
```
GET /api/v1/spot/accounts/{address}/state
```
Response: `{"code":0,"data":{"user":"0x...","aid":43933,"uid":43933,"B":[...],"O":[...]}}`

The `aid` field is your account ID.

## REST Endpoints

Prefix with `/api/v1/spot` or `/api/v1/perps`. Optional query param `accountID` specifies a sub-account (defaults to primary).

| Endpoint | Description |
|----------|-------------|
| `GET /accounts/{address}/state` | Account ID, user ID, balances, open orders |
| `GET /accounts/{address}/balances` | Balances: coinID, coin, total, locked |
| `GET /accounts/{address}/orders` | Open orders. Optional: `?symbol=X` |
| `GET /accounts/{address}/positions` | Open positions (perps only) |
| `GET /accounts/{address}/orders/history` | Historical orders. Optional: `symbol`, `startTime`, `endTime`, `limit` |
| `GET /accounts/{address}/trades` | Trade history. Optional: `symbol`, `startTime`, `endTime`, `limit` |
| `GET /accounts/{address}/fee-rate` | Fee tier info |
| `GET /accounts/{address}/api-keys` | API key list |
| `GET /accounts/{address}/funding-history` | Funding payments (perps only) |

## Rate Limits

| Endpoint | Weight |
|----------|--------|
| balances, orders, positions, state, api-keys | 5 |
| fee-rate | 2 |
| order history, trades, funding history | 20 + floor(items/20) |

## Examples

```bash
# Look up account ID
curl -s "https://mainnet-gw.sodex.dev/api/v1/spot/accounts/0xYourAddress/state"

# Get spot balances
curl -s "https://mainnet-gw.sodex.dev/api/v1/spot/accounts/0xYourAddress/balances"

# Get perps positions
curl -s "https://mainnet-gw.sodex.dev/api/v1/perps/accounts/0xYourAddress/positions"

# Get trade history (last 50)
curl -s "https://mainnet-gw.sodex.dev/api/v1/spot/accounts/0xYourAddress/trades?limit=50"
```
