# Sodex + Privy Integration

Use this reference when the user is building on top of [Privy](https://privy.io) (embedded wallets, social login, server wallets) and needs their app or agent to trade on Sodex.

## When to use Privy

Privy is embedded-wallet infrastructure: users log in with email / social / passkey and get a self-custodial EVM wallet. Keys live in TEEs with Shamir-style splits — your app and Privy cannot read the raw private key.

**Good fits:**

- **Consumer app** — users sign trades themselves through their Privy embedded wallet (no MetaMask required).
- **Autonomous agent / trading bot** — an offline backend signs trades on behalf of the user via a Privy **server wallet** that holds a Sodex API key.

## Recommended architecture (user + agent)

```
┌─────────────┐   1. Social login       ┌──────────────────┐
│  End user   │ ──────────────────────▶ │  Privy embedded  │
│  (browser)  │                          │  wallet (master) │
└─────────────┘                          └──────────────────┘
       │                                          │
       │ 2. One-time: user visits                 │
       │    sodex.com, connects                   │
       │    their Privy wallet, and               │
       │    creates an API key in the UI          │
       ▼                                          ▼
┌───────────────────────┐              ┌──────────────────────┐
│  Sodex web UI         │              │  Privy server wallet │
│  (create API key)     │              │  (holds API-key      │
│                       │              │   private key)       │
└───────────────────────┘              └──────────────────────┘
                                                  │
                                3. Agent backend  │
                                   uses server    │
                                   wallet to sign │
                                   every trade    ▼
                                         ┌──────────────────────┐
                                         │  Sodex REST API      │
                                         │  (POST /trade/...)   │
                                         └──────────────────────┘
```

1. User logs in with Privy → gets a master embedded wallet (EOA).
2. User visits `https://sodex.com/`, connects the Privy embedded wallet, and creates an API key through the UI. The UI returns a key name and a keypair.
3. Provision a **Privy server wallet** with the API key's private key so the backend can sign trades offline. Every trade is signed by this server wallet — user doesn't need to be online.

> **Note on wallet pre-generation.** Privy embedded wallets don't expose raw keys, so you cannot "move" one into a server wallet. The API keypair is created out-of-band (via Sodex UI); you then import the private key into a Privy server wallet using the pre-generated-key flow, or generate a fresh keypair on Privy and register its public key as your Sodex API key in the UI.

## Client side — sign trades with a Privy embedded wallet

If the user signs trades directly (no agent), use the React hook `useSignTypedData`. Sodex trading signatures are `ExchangeAction` under the engine-specific domain (`spot` for Spark, `futures` for Bolt).

```ts
import { useSignTypedData, useWallets } from '@privy-io/react-auth';

const { signTypedData } = useSignTypedData();
const { wallets } = useWallets();
const wallet = wallets[0]; // the user's embedded wallet

// `payloadHash` is keccak256(compact-JSON(ActionPayload{type, params})).
// See references/authentication.md for the full signing pipeline.
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
// After receiving signature: wrap in the Sodex 66-byte wire format
// (0x01 type byte || r || s || v-27). See references/authentication.md.
```

## Server side — sign with a Privy server wallet

Install `@privy-io/server-auth` (Node), or use the REST API directly. Server wallets require `PRIVY_APP_ID + PRIVY_APP_SECRET` and (recommended) a P-256 authorization key.

**Node.js — sign a Sodex trade with a server wallet:**

```ts
import { PrivyClient } from '@privy-io/server-auth';

const privy = new PrivyClient(process.env.PRIVY_APP_ID!, process.env.PRIVY_APP_SECRET!);

// `walletId` ("wallet-…") references the server wallet that holds the
// Sodex API key's private key. Create it once and store the id.
const { signature } = await privy.walletApi.ethereum.signTypedData({
  walletId,
  typedData: {
    domain: { name: 'futures', version: '1', chainId: 286623, verifyingContract: '0x0000000000000000000000000000000000000000' },
    types: {
      ExchangeAction: [
        { name: 'payloadHash', type: 'bytes32' },
        { name: 'nonce',       type: 'uint64'  },
      ],
    },
    primaryType: 'ExchangeAction',
    message: { payloadHash, nonce },
  },
});
```

**REST equivalent** — useful for Go / Python / Rust backends:

```bash
curl -X POST https://api.privy.io/v1/wallets/$WALLET_ID/rpc \
  -u "$PRIVY_APP_ID:$PRIVY_APP_SECRET" \
  -H "privy-app-id: $PRIVY_APP_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "eth_signTypedData_v4",
    "params": { "typed_data": { "domain": {...}, "types": {...}, "primaryType": "ExchangeAction", "message": {...} } }
  }'
```

## Go / Python backends using the Sodex SDK

Both Sodex SDKs expect a raw `ExchangeAction` digest to sign. When the private key lives in Privy (not locally), the flow is:

1. Compute the EIP-712 digest using the SDK primitives — `ExchangeAction{payloadHash, nonce}.Hash(domain)` in Go, or `ExchangeAction(...).hash(domain)` in Python.
2. Send the digest to Privy's `/v1/wallets/{id}/rpc` using `method: "personal_sign"` to get a 65-byte ECDSA signature.
3. Prepend `0x01` (SignatureType) → 66-byte Sodex wire signature.
4. Attach to your HTTP request via `X-API-Sign` header plus `X-API-Key: {key_name}` and `X-API-Nonce: {nonce}`.

## Gotchas

1. **`chainId` is mandatory in the EIP-712 domain.** Privy does not inject it — if you omit it, the signature recovers to a wrong address and Sodex rejects it silently.
2. **Wallet `id` ≠ wallet `address`.** Server-SDK methods take the `wallet-...` id; client-SDK hooks take the `0x...` address. Mixing them up throws confusing "wallet not found" errors.
3. **Raw private key is not exposed.** You cannot export it programmatically to feed into `client.New(Config{PrivateKey: ...})`. Every request goes through Privy's signing API.
4. **Policy engine for agents.** Add a server-wallet policy restricting signatures to the Sodex domain (`name: "futures"` / `"spot"`, `chainId: 286623`) so a compromised agent can't sign arbitrary transactions.
5. **Authorization keys (P-256).** For server wallets, configure an authorization key and sign each request body — this is defence-in-depth against an app-secret leak.
6. **Domain must match the engine.** Perps endpoints require `name: "futures"`; spot requires `name: "spot"`. Using the wrong one gets signatures rejected silently.
7. **Privy signing adds latency** (~100–300 ms for server-wallet RPCs, plus user-interaction time for embedded-wallet prompts). For latency-sensitive strategies the server-wallet path is the only viable one.

## Links

- Privy embedded-wallet EIP-712 signing: https://docs.privy.io/wallets/using-wallets/ethereum/sign-typed-data
- Privy server wallets quickstart: https://docs.privy.io/guide/server-wallets/quickstart/api
- Privy server-wallet create / sign SDK: https://docs.privy.io/guide/server-wallets/create
- Sodex authentication deep dive: [authentication.md](authentication.md)
