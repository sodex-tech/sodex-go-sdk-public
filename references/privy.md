# Sodex + Privy Integration

Use this reference when the user is building on top of [Privy](https://privy.io) (embedded wallets, social login) and needs their app or agent to trade on Sodex.

## When to use Privy

Privy is embedded-wallet infrastructure: users log in with email / social / passkey and get a self-custodial EVM wallet. Keys live in TEEs with Shamir-style splits — your app and Privy cannot read the raw private key.

For Sodex, Privy is useful as the **user-login + master-wallet** layer. All trading goes through a Sodex-native API key that the user manages themselves, driven by the `sodex` CLI. Privy is not in the signing path for trades.

> **Note.** Direct per-trade signing via a Privy embedded wallet (`useSignTypedData` → `ExchangeAction`) is not covered here: there's no JavaScript SDK yet that wraps Sodex's two-layer EIP-712 payload, nonce management, and wire-format conversion. Until that's available, the supported integration is the API-key flow below.

## Architecture

- **Master wallet** — the user's Privy embedded wallet. Holds deposited funds on-chain and is the identity that creates a Sodex API key in the Sodex UI.
- **API key** — a secondary EVM keypair the user creates on https://sodex.com/. The agent uses the API key's private key to sign trades. This key lives in the user's own secret store, not in Privy.

```
┌─────────────┐  1. Social login       ┌────────────────────────┐
│  End user   │ ─────────────────────▶ │  Privy embedded wallet │
│  (browser)  │                         │  (master wallet)       │
└─────────────┘                         └────────────────────────┘
                                                   │
                            2. User visits          │
                               sodex.com, deposits  │
                               funds, creates an    │
                               API key in the UI    ▼
                                         ┌─────────────────────┐
                                         │  Sodex API key      │
                                         │  (user-managed      │
                                         │   EVM keypair)      │
                                         └─────────────────────┘
                                                   │
                       3. Agent backend uses the   │
                          API key private key to   │
                          sign trades via CLI      ▼
                                         ┌─────────────────────┐
                                         │  Sodex REST API     │
                                         │  (POST /trade/...)  │
                                         └─────────────────────┘
```

## Set up the agent

Assumes the user has already logged in with Privy and has a funded Sodex account. Start from creating the API key.

1. **Create a Sodex API key** at https://sodex.com/apikeys. The UI returns a `keyName` and a `privateKey`. **Store the private key securely** (`.env` file, AWS Secrets Manager, 1Password, etc. — **not** committed to git). Never use the Privy master wallet private key — it's not exportable, and it's not the right key for trading anyway.

2. **Look up the account ID** (use the Privy embedded wallet address):
   ```
   sodex --testnet account-id 0xYourMasterWalletAddress
   ```
   Output: `{ "accountID": 12345, "userID": 12345 }`

3. **Configure the agent's environment** — pick the syntax for your shell:

   **macOS / Linux (bash, zsh):**
   ```bash
   export SODEX_PRIVATE_KEY=0x...        # API key private key from step 1
   export SODEX_API_KEY=my-bot-key-01    # API key name from step 1
   export SODEX_ACCOUNT_ID=12345         # from step 2
   export SODEX_TESTNET=1                # omit for mainnet
   ```

   **Windows PowerShell:**
   ```powershell
   $env:SODEX_PRIVATE_KEY = "0x..."
   $env:SODEX_API_KEY     = "my-bot-key-01"
   $env:SODEX_ACCOUNT_ID  = "12345"
   $env:SODEX_TESTNET     = "1"
   ```

   **Windows cmd.exe:**
   ```cmd
   set SODEX_PRIVATE_KEY=0x...
   set SODEX_API_KEY=my-bot-key-01
   set SODEX_ACCOUNT_ID=12345
   set SODEX_TESTNET=1
   ```

4. **Drive the CLI** — any language can spawn it and parse `--format json` output:
   ```
   sodex balance perps
   sodex orders place perps --symbol BTC-USD --side buy --type limit --price 70000 --qty 0.01
   sodex orders cancel perps --symbol-id 1 --order-id <orderID>
   sodex positions
   ```

See [references/trading.md](trading.md) for the full command reference and [references/account-query.md](account-query.md) for account / balance lookups.

## Can I use Privy server wallets to hold the API key?

**Not today.** Privy server wallets are Privy-custodied — you cannot upload an existing private key for Privy to sign with. You can only sign with keys Privy generated inside its TEE.

There's a theoretical path where the agent's signing key is a Privy-generated server wallet and the user registers that wallet's address on Sodex as their API key, but Sodex doesn't currently support registering an externally-generated address as an API key. Until that's added, agents must use a user-managed API key per the flow above.

## Gotchas

1. **Fund before trading.** A fresh Privy wallet has zero on-Sodex balance. Trading commands will fail with "insufficient balance" until the user deposits via the Sodex UI.
2. **Account ID is required for trading commands.** `SODEX_ACCOUNT_ID` (or `--account-id`) must be set; without it the CLI cannot construct a valid signing payload.
3. **API key private key is user-managed.** Privy does not store or rotate it. Treat it like any other production secret.
4. **Testnet first.** Sodex mainnet is chain ID 286623, testnet is 138565. Always validate the full flow on testnet before touching mainnet — set `SODEX_TESTNET=1` or pass `--testnet`.
5. **Privy embedded wallets don't expose raw keys.** The master wallet's private key is not retrievable from your code, which is why the Sodex API key has to be generated separately in the Sodex UI.

## Links

- Privy overview: https://docs.privy.io
- Sodex CLI / SKILL.md: https://raw.githubusercontent.com/sodex-tech/sodex-go-sdk-public/main/SKILL.md
- Sodex authentication deep dive: [authentication.md](authentication.md)
