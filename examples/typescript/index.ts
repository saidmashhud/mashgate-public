/**
 * Mashgate SDK — TypeScript quickstart
 *
 * Demonstrates the three things every integration needs:
 *   1. Initialize the client from environment variables.
 *   2. Create a hosted checkout session (and, optionally, a single payment).
 *   3. Receive webhooks and verify their signature with the SDK helper.
 *
 * Run the checkout demo:
 *     export MASHGATE_BASE_URL="https://api.mashgate.uz"   # or http://localhost:9661 for local dev
 *     export MASHGATE_API_KEY="mg_test_..."
 *     npm run checkout
 *
 * Run the webhook receiver (Express on :3000):
 *     export MASHGATE_WEBHOOK_SECRET="whsec_..."
 *     npm run webhook
 */
import express from "express";
import {
  MashgateClient,
  MashgateError,
  verifyWebhookSignature,
  eventKey,
  eventPayload,
  WebhookTopic,
  type CheckoutSession,
  type PaymentIntent,
  type WebhookEvent,
} from "@mashgate/sdk";

// ── 1. Client init from env ─────────────────────────────────────────────
//
// Keys are environment-scoped: `mg_test_...` for sandbox, `mg_live_...` for
// production. Never hard-code them. `baseUrl` falls back to MASHGATE_API_URL
// so either env var name works.
function makeClient(): MashgateClient {
  const baseUrl = process.env.MASHGATE_BASE_URL ?? process.env.MASHGATE_API_URL;
  const apiKey = process.env.MASHGATE_API_KEY;

  if (!baseUrl || !apiKey) {
    throw new Error(
      "Set MASHGATE_BASE_URL (or MASHGATE_API_URL) and MASHGATE_API_KEY in your environment.",
    );
  }

  return new MashgateClient({
    baseUrl,
    apiKey,
    timeout: 10_000, // ms, default 30_000
    maxRetries: 2, // retry on 5xx / network errors, default 0
    // Auto-attached to every POST/PUT/PATCH so retries are safe.
    idempotencyKey: () => crypto.randomUUID(),
  });
}

// ── 2a. Create a hosted checkout session ────────────────────────────────
//
// Mashgate hosts the payment page; you redirect the customer to `session.url`
// and they come back to your `successUrl` / `cancelUrl`.
async function createCheckout(mg: MashgateClient): Promise<CheckoutSession> {
  const session = await mg.checkout.createSession({
    currency: "UZS",
    successUrl: "https://example.com/success?session={sessionId}",
    cancelUrl: "https://example.com/cancel",
    expiresInMinutes: 30,
    lineItems: [
      {
        name: "Pro plan",
        description: "Monthly subscription",
        quantity: 1,
        unitPrice: { amount: "150000.00", currency: "UZS" },
      },
    ],
    metadata: { orderId: "order-1001" },
  });

  console.log("Checkout session created:");
  console.log("  sessionId:  ", session.sessionId);
  console.log("  status:     ", session.status);
  console.log("  total:      ", session.totalAmount.amount, session.totalAmount.currency);
  console.log("  redirect to:", session.url); // send the customer here
  return session;
}

// ── 2b. (Optional) Create a single server-side payment ──────────────────
//
// Use this when you already hold a tokenized payment method (e.g. a saved
// card) and want to charge it directly instead of using the hosted page.
async function createPayment(mg: MashgateClient): Promise<PaymentIntent> {
  const payment = await mg.payments.create({
    amount: "150000.00",
    currency: "UZS",
    orderId: "order-1001",
    autoCapture: true,
  });

  console.log("Payment created:");
  console.log("  paymentId:", payment.paymentId);
  console.log("  status:   ", payment.status);
  console.log("  amount:   ", payment.amount.amount, payment.amount.currency);
  return payment;
}

// ── 3. Minimal webhook receiver ─────────────────────────────────────────
//
// Mashgate (and HookLine on its behalf) POSTs events here with two headers:
//   x-hl-signature  →  "v1=<hex hmac-sha256>"
//   x-hl-timestamp  →  Unix epoch milliseconds (signed alongside the body)
//
// IMPORTANT: verify against the RAW request body, before any JSON parsing —
// re-serializing changes bytes and breaks the HMAC. `express.raw()` gives us
// the untouched Buffer.
function buildWebhookApp(): express.Express {
  const secret = process.env.MASHGATE_WEBHOOK_SECRET;
  if (!secret) {
    throw new Error("Set MASHGATE_WEBHOOK_SECRET to your endpoint signing secret.");
  }

  const app = express();

  app.post(
    "/webhooks/mashgate",
    express.raw({ type: "application/json" }),
    async (req, res) => {
      const signature = req.header("x-hl-signature");
      const timestamp = req.header("x-hl-timestamp");
      if (!signature || !timestamp) {
        return res.status(400).send("missing signature headers");
      }

      // Real async signature: (payload, signatureHeader, secret, timestamp).
      // `req.body` is a Buffer (Uint8Array) of the raw bytes.
      const ok = await verifyWebhookSignature(req.body, signature, secret, timestamp);
      if (!ok) {
        return res.status(401).send("invalid signature");
      }

      // Safe to parse now. Use eventKey()/eventPayload() so both envelope-v1
      // (`topic`/`payload`) and legacy (`eventType`/`data`) emissions work.
      const event = JSON.parse(req.body.toString("utf8")) as WebhookEvent;
      const topic = eventKey(event);

      switch (topic) {
        case WebhookTopic.CheckoutSessionCompleted:
          console.log("Checkout completed:", eventPayload(event));
          break;
        case WebhookTopic.PaymentCompleted:
          console.log("Payment completed:", eventPayload(event));
          break;
        default:
          console.log("Received event:", topic, eventPayload(event));
      }

      // Ack fast (2xx) so Mashgate stops retrying; do slow work async.
      res.status(200).json({ received: true });
    },
  );

  return app;
}

// ── Entrypoint: pick a mode via the first CLI arg ───────────────────────
async function main() {
  const mode = process.argv[2] ?? "checkout";

  try {
    if (mode === "webhook") {
      const port = Number(process.env.PORT ?? 3000);
      buildWebhookApp().listen(port, () => {
        console.log(`Webhook receiver listening on http://localhost:${port}/webhooks/mashgate`);
      });
      return;
    }

    const mg = makeClient();
    if (mode === "payment") {
      await createPayment(mg);
    } else {
      await createCheckout(mg);
    }
  } catch (err) {
    // Non-2xx API responses raise MashgateError with status/code/requestId.
    if (err instanceof MashgateError) {
      console.error(`MashgateError ${err.status} ${err.code ?? ""}: ${err.message}`);
      if (err.requestId) console.error("requestId:", err.requestId);
      process.exit(1);
    }
    throw err;
  }
}

main();
