# @mashgate/sdk

Official TypeScript SDK for the Mashgate Payment Gateway API. Mirrors
the gRPC contracts in
[mashgate/contracts/proto/v1/](https://github.com/saidmashhud/mashgate/tree/main/contracts/proto/v1)
and reaches the gateway over REST (gRPC-JSON transcoding via Envoy).

**Repo:** <https://github.com/saidmashhud/mashgate-public>

```bash
npm install @mashgate/sdk
```

## Quick start

```ts
import { MashgateClient } from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: "https://api.mashgate.uz",
  apiKey: "mg_test_key",
});

const payment = await mg.payments.create({
  amount: "100.00",
  currency: "UZS",
  description: "Subscription",
});
```

## Wallet APIs

The SDK exposes two distinct wallet surfaces:

- **`mg.wallet`** — *end-user view* (saved payment methods, balance,
  movements). Use with a customer-issued JWT.
- **`mg.walletAdmin`** — *admin / merchant view* (full
  `wallet.v1.WalletService`). Use with an admin JWT or service account
  API key.

### `walletAdmin` — full WalletService

```ts
import {
  MashgateClient,
  Currency,
  Network,
  Mint,
  WalletType,
} from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: "https://api.mashgate.uz",
  apiKey: process.env.MASHGATE_API_KEY!,
});

// Off-chain wallet
const w = await mg.walletAdmin.create({
  subject_id: "user-123",
  subject_type: "user",
  wallet_type: WalletType.Fiat,
  currency: Currency.UZS,
  idempotency_key: "idem-create-1",
});

// On-chain wallet (BIP-39 mnemonic returned ONCE — surface to user, never persist)
const chain = await mg.walletAdmin.createChain({
  subject_id: "user-123",
  subject_type: "user",
  currency: Currency.USDC,
  network: Network.Solana,
});
showOnceToEndUser(chain.mnemonic);

// Deposit address
//   - SPL token: pass `mint` → returns the Associated Token Account.
//   - Native asset: leave `mint` empty → returns the wallet owner address.
const ata = await mg.walletAdmin.depositAddress(
  chain.wallet.wallet_id,
  Network.Solana,
  Mint.USDCSolanaMainnet,
);
const sol = await mg.walletAdmin.depositAddress(
  chain.wallet.wallet_id,
  Network.Solana,
  "",
);

// Withdraw — `mint` selects SPL token, empty / undefined = native SOL.
const tx = await mg.walletAdmin.withdraw(chain.wallet.wallet_id, {
  amount: "10.50",
  destination_type: "crypto_address",
  destination_id: "RecipientSolanaAddr",
  network: Network.Solana,
  mint: Mint.USDCSolanaMainnet,
  idempotency_key: "idem-w-1",
});

// Compliance / fraud
await mg.walletAdmin.freeze(w.wallet_id, "fraud-investigation");
await mg.walletAdmin.unfreeze(w.wallet_id, "case-resolved");

// Pagination — opaque cursor; empty cursor = first page.
let resp = await mg.walletAdmin.list({ limit: 50 });
while (resp.next_cursor) {
  resp = await mg.walletAdmin.list({ limit: 50, cursor: resp.next_cursor });
}

const single = await mg.walletAdmin.getTransaction(w.wallet_id, "tx-xxx");
```

### Typed constants

`Currency`, `Network`, `Mint`, `WalletType`, `WalletStatus`,
`TransactionType`, `TransactionStatus`, `TransactionReason` are
const-as-object enums — IDE autocomplete on known values, literal-union
type for compile-time check, plain JSON string on the wire. Untyped
string literals (`"USDC"`, `"SOLANA"`) stay assignable, so existing
callers don't break.

```ts
Currency.USDC; // "USDC"
Network.Solana; // "SOLANA"
Mint.USDCSolanaMainnet; // "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
```

`Mint` allows arbitrary strings outside the listed mainnet whitelist —
server-side (`chain-rpc.derive_spl_ata`) is the authoritative validator.

## Errors

Non-2xx responses raise `MashgateError`:

```ts
import { MashgateError } from "@mashgate/sdk";

try {
  await mg.walletAdmin.get("missing");
} catch (e) {
  if (e instanceof MashgateError) {
    console.error(e.status, e.code, e.message);
  }
}
```

## Webhooks

```ts
import { verifyWebhookSignature } from "@mashgate/sdk";

app.post("/webhooks/mashgate", express.raw({ type: "application/json" }), async (req, res) => {
  const ok = await verifyWebhookSignature(
    req.body,
    req.header("x-hl-signature")!,
    process.env.WEBHOOK_SECRET!,
    req.header("x-hl-timestamp")!,
  );
  if (!ok) return res.status(401).end();
  // handle event
});
```

## Development

```bash
npm install
npm run build           # tsc → dist/
npm test                # vitest run
```
