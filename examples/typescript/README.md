# Mashgate SDK — TypeScript quickstart

A single runnable `index.ts` covering the three things every integration needs:

1. **Client init** from environment variables.
2. **Create a hosted checkout session** (and, optionally, a single payment).
3. **Receive webhooks** and verify their signature with the SDK helper.

## Setup

```sh
npm install
cp .env.example .env   # then fill in your values, or just export them:

export MASHGATE_BASE_URL="https://api.mashgate.uz"   # or http://localhost:9661 for local dev
export MASHGATE_API_KEY="mg_test_..."
```

The example depends on `@mashgate/sdk` from npm. To run it against this repo's
local SDK source instead, point the dependency at a relative path before
installing:

```jsonc
// package.json
"@mashgate/sdk": "file:../../sdk/typescript"
```

## Run

```sh
npm run checkout   # create a hosted checkout session, print the redirect URL
npm run payment    # create a single server-side payment, print the payment id
npm run webhook    # start an Express receiver on :3000 that verifies signatures
```

## What it shows

- `new MashgateClient({ baseUrl, apiKey, timeout, maxRetries, idempotencyKey })`
- `mg.checkout.createSession(...)` → `session.url` is where you redirect the buyer.
- `mg.payments.create(...)` → `payment.paymentId` / `payment.status`.
- `await verifyWebhookSignature(rawBody, signature, secret, timestamp)` using the
  `x-hl-signature` and `x-hl-timestamp` headers, against the **raw** request body.
- `eventKey()` / `eventPayload()` + `WebhookTopic` for routing both envelope-v1
  and legacy event shapes.
- `MashgateError` (carries `status`, `code`, `requestId`).

## Testing the webhook locally

Expose the receiver with a tunnel (e.g. `ngrok http 3000`), register the public
URL as a webhook endpoint in the Mashgate console, copy its signing secret into
`MASHGATE_WEBHOOK_SECRET`, then trigger a test delivery from the console.
