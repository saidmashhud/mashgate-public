# Getting started — TypeScript

The TypeScript SDK (`@mashgate/sdk`) covers the full Mashgate surface as typed
resources on one client: `payments`, `checkout`, `auth`, `wallet` / `walletAdmin`,
`webhooks`, `chat`, `notify`, `storage`, `flags`, `billing`, `iam`, `mail`, and
more. It runs on Node 18+ and any runtime with a global `fetch`.

## Install

```sh
npm install @mashgate/sdk
```

This documents **v1.7.0**. The package ships ESM with bundled type
declarations — no `@types` package needed.

## Prerequisites

Everything in Mashgate is multi-tenant. Before your first call you need a
**tenant** and an **API key** from the Mashgate console — see
[Building a vertical → Step 0](../guides/building-a-vertical.md#step-0--provision-a-tenant).

Keys are environment-scoped: `mg_test_...` for sandbox, `mg_live_...` for
production. Keep the key out of source:

```sh
export MASHGATE_BASE_URL="https://api.mashgate.uz"   # or http://localhost:9661 for local dev
export MASHGATE_API_KEY="mg_test_..."
```

## Initialize the client

```ts
import { MashgateClient } from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: process.env.MASHGATE_BASE_URL!,
  apiKey: process.env.MASHGATE_API_KEY!,
});
```

The constructor accepts more options when you need them:

```ts
const mg = new MashgateClient({
  baseUrl: process.env.MASHGATE_BASE_URL!,
  apiKey: process.env.MASHGATE_API_KEY!,
  timeout: 10_000,                          // ms, default 30_000
  maxRetries: 3,                            // retry on 5xx / network, default 0
  idempotencyKey: () => crypto.randomUUID(), // auto-attached to POST/PUT/PATCH
});
```

If you authenticate end users (rather than using an API key), call
`mg.setAccessToken(token)` and the client sends `Authorization: Bearer`.

## Your first call

Create a hosted checkout session and redirect your customer to it:

```ts
import { MashgateClient } from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: process.env.MASHGATE_BASE_URL!,
  apiKey: process.env.MASHGATE_API_KEY!,
});

const session = await mg.checkout.createSession({
  currency: "UZS",
  successUrl: "https://example.com/success?session={sessionId}",
  cancelUrl: "https://example.com/cancel",
  lineItems: [
    { name: "Pro plan", quantity: 1, unitPrice: { amount: "150000.00", currency: "UZS" } },
  ],
});

console.log("redirect customer to:", session.url);
```

`session` is a typed `CheckoutSession` with `sessionId`, `status`
(`pending | completed | expired | cancelled`), `totalAmount`, `url`,
`successUrl`, `cancelUrl`, `expiresAt`, and `createdAt`.

### Or: a single payment

```ts
const payment = await mg.payments.create({
  amount: "150000.00",
  currency: "UZS",
  description: "Pro plan",
});
console.log(payment.paymentId, payment.status);
```

### Errors

Non-2xx responses throw a `MashgateError` carrying `status`, `code`, and
`requestId` (include it in support requests):

```ts
import { MashgateError } from "@mashgate/sdk";

try {
  await mg.checkout.getSession("missing");
} catch (e) {
  if (e instanceof MashgateError) {
    console.error(e.status, e.code, e.requestId, e.message);
  }
}
```

### Verify webhooks

```ts
import { verifyWebhookSignature } from "@mashgate/sdk";

// `rawRequestBody` must be the RAW bytes (e.g. express.raw), not parsed JSON.
// HookLine signs HMAC-SHA256 over `{timestamp}.{body}` and sends the signature
// as `x-hl-signature: v1=<hex>` plus `x-hl-timestamp` (Unix ms).
const ok = await verifyWebhookSignature(
  rawRequestBody,
  req.header("x-hl-signature")!,
  process.env.WEBHOOK_SECRET!,
  req.header("x-hl-timestamp")!,
);
```

## Next steps

- [Building a vertical](../guides/building-a-vertical.md) — the end-to-end path:
  provision a tenant, wire modules, react to events, deploy.
- [Best practices](../best-practices.md) — idempotency, money/ledger as source of
  truth, multi-tenancy, webhook handling, error handling.
- [Service catalog](../modules/service-catalog.md) — the full module / RPC
  reference.
