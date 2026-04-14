# Sodex + Privy Integration

Use this reference when the user is building on top of [Privy](https://privy.io) (embedded wallets, social login) and needs their app to trade on Sodex.

## When to use Privy

Privy is embedded-wallet infrastructure: users log in with email / social / passkey and get a self-custodial EVM wallet. Keys live in TEEs with Shamir-style splits — **your app cannot read the raw private key, and neither can Privy staff**.

**Good fit:** consumer apps where users sign trades themselves via their Privy embedded wallet (no MetaMask required).

**Not a fit today:** hosting a Sodex API key inside a Privy server wallet for an offline agent. Privy server wallets cannot import an existing private key — they only sign with Privy-generated keys. See "Can I use Privy server wallets for automated agents?" below.

## Architecture

There are two signing paths on Sodex:

- **Master wallet** — the user's primary EVM address. Signs fund transfers between engines and is used as the identity when creating an API key in the Sodex UI.
- **API key** — a secondary EVM keypair the user creates on https://sodex.com/ to sign trades. Recommended for all repeated trading.

With Privy, the master wallet is the Privy embedded wallet. The API key is managed entirely by the user, outside of Privy:

```
┌─────────────┐   1. Social login     ┌────────────────────────┐
│  End user   │ ────────────────────▶ │  Privy embedded wallet │
│  (browser)  │                        │  (master wallet)       │
└─────────────┘                        └────────────────────────┘
                                                 │
                                 2. User signs   │
                                    trades via   │ eth_signTypedData_v4
                                    Privy hook   │
                                                 ▼
                                        ┌───────────────────┐
                                        │  Sodex REST API   │
                                        └───────────────────┘
```

For a pure user-signs-every-trade product, no API key is needed — every trade can be signed directly by the Privy embedded wallet.

## Client side — sign trades with a Privy embedded wallet

Trades are `ExchangeAction` EIP-712 messages under the engine-specific domain (`spot` for Spark, `futures` for Bolt).

```ts
import { useSignTypedData, useWallets } from '@privy-io/react-auth';

const { signTypedData } = useSignTypedData();
const { wallets } = useWallets();
const wallet = wallets[0]; // the user's Privy embedded wallet

// payloadHash = keccak256(compact-JSON(ActionPayload{type, params})).
// See references/authentication.md for how to compute it.
const { signature } = await signTypedData(
  {
    domain: {
      name: 'futures',         // "spot" for spot trades, "futures" for perps
      version: '1',
      chainId: 286623,         // MUST be set; Privy does not inject it
      verifyingContract: '0x0000000000000000000000000000000000000000',
    },
    types: {
      ExchangeAction: [
        { name: 'payloadHash', type: 'bytes32' },
        { name: 'nonce',       type: 'uint64'  },
      ],
    },
    primaryType: 'ExchangeAction',
    message: { payloadHash, nonce },
  },
  { address: wallet.address },
);
// `signature` is a 65-byte 0x-prefixed hex string (r || s || v, v ∈ {27, 28}).
// Convert to Sodex wire format: [0x01 type byte || r || s || (v - 27)] = 66 bytes.
// Send as X-API-Sign header. See references/authentication.md.
```

**Important: use `eth_signTypedData_v4` (via `useSignTypedData`), NOT `personal_sign`.**
`personal_sign` prepends the `"\x19Ethereum Signed Message:\n32"` prefix and re-hashes, which would make the signature incompatible with Sodex's EIP-712 verification. `eth_signTypedData_v4` signs the EIP-712 digest directly.

## Backend / agent flow — API key is user-managed

If the user wants an agent that trades without their direct interaction, they must manage a Sodex API key themselves. Privy is **not** in this path.

1. User logs in with Privy (embedded wallet = master wallet).
2. User creates a Sodex API key at https://sodex.com/. The UI generates a new EVM keypair and returns both the key name and the private key. **The user is responsible for storing this private key securely** (`.env` file, AWS Secrets Manager, 1Password, etc. — **not** committed to git).
3. The agent invokes the `sodex` CLI, which reads the API key from env vars and handles signing, nonces, and symbol resolution automatically:

```bash
# Install once.
go install github.com/sodex-tech/sodex-go-sdk-public/cmd/sodex@latest

# Agent configures these from its secure store (never hard-coded).
export SODEX_PRIVATE_KEY=0x…         # the API key's private key
export SODEX_API_KEY=my-bot-key-01   # the API key name created in the UI
export SODEX_TESTNET=1               # omit for mainnet

# Place orders, manage positions, query state — all via the CLI.
sodex markets perps --format json
sodex balance perps
sodex orders place perps --symbol BTC-USDT --side buy --type limit --price 70000 --qty 0.01
sodex orders cancel perps --symbol-id 1 --order-id <orderID>
sodex positions
```

For programmatic control, drive the CLI from any language via subprocess and parse the `--format json` output. See [references/trading.md](trading.md) for the full command reference.

## Can I use Privy server wallets for automated agents?

**Not today.** Privy server wallets are Privy-custodied — you cannot upload an existing private key for Privy to sign with. You can only sign with keys Privy generated inside its TEE.

There's a theoretical path where the agent's signing key is a Privy-generated server wallet, and the user registers that wallet's address on Sodex as their API key — but Sodex doesn't currently support registering an externally-generated public key as an API key through the UI. Until that's added, agents must use a user-managed API key per the flow above.

## Gotchas

1. **`chainId` is mandatory in the EIP-712 domain.** Privy does not inject it — omit it and the signature recovers to a wrong address, so Sodex rejects it silently.
2. **Use `eth_signTypedData_v4`, never `personal_sign`.** Only typed-data signing produces a raw EIP-712 digest signature; `personal_sign` adds a prefix that breaks verification.
3. **Domain must match the engine.** Perps endpoints require `name: "futures"`; spot requires `name: "spot"`. Wrong domain = rejected signature.
4. **`v` byte conversion.** Privy returns signatures with `v ∈ {27, 28}`; Sodex expects `v ∈ {0, 1}`. Subtract 27 before prepending the `0x01` type byte.
5. **Privy does not export raw keys.** The Privy embedded wallet's private key is not retrievable from your code. If you need a key in a backend (e.g. for the Sodex SDK's `PrivateKey` field), that key must come from somewhere other than Privy — typically a user-managed Sodex API key.
6. **Network config.** Sodex mainnet is chain ID 286623, testnet is 138565. If you display transactions in Privy's UI, add these as custom chains in your Privy dashboard.
7. **Embedded-wallet signing is interactive.** `signTypedData` may prompt the user for confirmation; for silent / high-frequency signing you cannot avoid the UI unless you use headless signing (typically only available under specific Privy enterprise configurations).

## Links

- Privy embedded-wallet EIP-712 signing: https://docs.privy.io/wallets/using-wallets/ethereum/sign-typed-data
- Privy server wallets overview: https://docs.privy.io/guide/server-wallets
- Sodex authentication deep dive: [authentication.md](authentication.md)
