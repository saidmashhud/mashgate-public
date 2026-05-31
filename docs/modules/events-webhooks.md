# Events & Webhooks

React to platform events (payments, refunds, wallet movements, KYC, …) via
**HookLine** — Mashgate's outbound webhook delivery system. You register an
endpoint, subscribe to event types, and HookLine POSTs signed payloads to your
URL with retries and a dead-letter queue. Covers the SDK `events` and
`webhooks` resources.

## When to use

- You want to be event-driven instead of polling — fulfill an order when a
  payment captures, credit a wallet when a deposit confirms, unlock a feature
  when KYC passes.
- You need a durable, retried, signed delivery channel for platform events.
- You're managing endpoints, subscriptions, delivery history, or replaying
  failed deliveries from the DLQ.

## Key operations

Two SDK surfaces exist. The top-level `Client` webhook methods (Go / TS /
Python) manage endpoints **and** verify signatures — start here. The richer
`events` client (Go) adds subscriptions, delivery history, and DLQ management.

### Register an endpoint

The signing secret is returned **once** on create — store it securely (env var
/ secret store), never log it. Leave event types empty to receive all.

**Go**

```go
ep, err := mg.CreateWebhookEndpoint(ctx, mashgate.CreateWebhookEndpointRequest{
    URL:         "https://api.example.com/webhooks/mashgate",
    Description: "order fulfillment",
    EventTypes:  []string{mashgate.EventPaymentCaptured, "wallet.credit"},
})
storeSecret(ep.SigningSecret) // returned only here

eps, err := mg.ListWebhookEndpoints(ctx)
err = mg.UpdateWebhookEndpoint(ctx, ep.EndpointID, mashgate.UpdateWebhookEndpointRequest{
    Status: "disabled",
})
err = mg.DeleteWebhookEndpoint(ctx, ep.EndpointID)
err = mg.TestWebhookEndpoint(ctx, ep.EndpointID) // send a synthetic test event
```

**TypeScript**

```typescript
const { endpoint } = await mg.webhooks.createEndpoint({
  url: "https://api.example.com/webhooks/mashgate",
  description: "order fulfillment",
  events: ["payment.captured", "wallet.credit"],
});
storeSecret(endpoint.signingSecret);

await mg.webhooks.listEndpoints();
await mg.webhooks.testEndpoint(endpoint.endpointId);
await mg.webhooks.deleteEndpoint(endpoint.endpointId);
```

**Python**

```python
ep = mg.webhooks.create_endpoint(
    url="https://api.example.com/webhooks/mashgate",
    description="order fulfillment",
    event_types=["payment.captured", "wallet.credit"],
)
store_secret(ep["endpoint"]["signing_secret"])
mg.webhooks.list_endpoints()
mg.webhooks.test_endpoint(ep["endpoint"]["endpoint_id"])
mg.webhooks.delete_endpoint(ep["endpoint"]["endpoint_id"])
```

### Verify signatures and handle events

Always verify before parsing. HookLine signs with HMAC-SHA256 over
`{timestamp}.{body}`, hex-encoded and prefixed `v1=`, sent in the `x-hl-signature`
header with `x-hl-timestamp`. The SDK rejects payloads older than 5 minutes
(replay protection). Read the **raw** body before any JSON parsing.

**Go** — `ConstructEvent` verifies + parses in one call:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    event, err := mg.ConstructEvent(
        body,
        r.Header.Get("x-hl-signature"),
        r.Header.Get("x-hl-timestamp"),
        os.Getenv("MASHGATE_WEBHOOK_SECRET"),
    )
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    // Dedupe on the canonical id; an already-seen event is a no-op.
    if alreadyProcessed(event.CanonicalID()) {
        w.WriteHeader(http.StatusOK)
        return
    }

    switch event.EventType { // legacy form; envelope v1 uses event.Topic
    case mashgate.EventPaymentCaptured:
        enqueueFulfillment(event.PayloadBytes()) // slow work async
    }
    w.WriteHeader(http.StatusOK) // 2xx fast
}
```

`VerifySignature(secret, timestamp, body, signature)` and `ParseEvent(body)` are
available separately if you need them. `PayloadBytes()` and `CanonicalID()`
transparently handle both envelope-v1 (`payload` / `id` / `topic`) and legacy
(`data` / `event_id` / `event_type`) shapes.

**TypeScript** — `webhooks.handleRequest` verifies and dispatches to handlers
registered with `on`:

```typescript
const secret = process.env.MASHGATE_WEBHOOK_SECRET!;

mg.webhooks.on("payment.captured", async (event) => {
  await fulfillOrder(eventPayload(event)); // eventPayload() picks payload|data
});
mg.webhooks.on("*", (event) => console.log("received", eventKey(event)));

app.post("/webhooks", express.raw({ type: "*/*" }), async (req, res) => {
  try {
    await mg.webhooks.handleRequest(
      req.body.toString(),
      req.headers["x-hl-signature"] as string,
      secret,
      req.headers["x-hl-timestamp"] as string,
    );
    res.json({ ok: true }); // 2xx fast
  } catch {
    res.status(401).end();
  }
});
```

Or verify only: `await mg.webhooks.verify(rawBody, signature, secret, timestamp)`
returns a boolean.

**Python** — `verify_webhook_signature(payload, signature, secret)` returns a
boolean (HMAC-SHA256 over the raw body):

```python
from mashgate.webhooks import verify_webhook_signature

@app.post("/webhooks")
def handler(request):
    body = request.get_data()  # raw bytes, before JSON parsing
    if not verify_webhook_signature(body, request.headers["X-Mashgate-Signature"], SECRET):
        return ("", 401)
    event = json.loads(body)
    if already_processed(event.get("id") or event.get("event_id")):
        return ("", 200)
    # dispatch on event["topic"] or event["event_type"], then return 2xx fast
    return ("", 200)
```

### Subscriptions, delivery history & DLQ (Go `events` client)

Attach the events client with `mg.WithEvents(mashgate.EventsConfig{})`, then
manage subscriptions, inspect deliveries, retry, and replay the dead-letter
queue.

**Go**

```go
mg = mg.WithEvents(mashgate.EventsConfig{})

sub, err := mg.Events.CreateSubscription(ctx, ep.EndpointID,
    []string{"payment.captured", "wallet.credit"})
subs, err := mg.Events.ListSubscriptions(ctx, ep.EndpointID)

deliveries, err := mg.Events.ListDeliveries(ctx, ep.EndpointID)
err = mg.Events.RetryDelivery(ctx, ep.EndpointID, deliveryID)

dlq, err := mg.Events.ListDLQ(ctx, ep.EndpointID)
replayed, err := mg.Events.ReplayDLQ(ctx, []string{deliveryID})

ep2, err := mg.Events.RotateSecret(ctx, ep.EndpointID) // ep2.SigningSecret
```

The events client retries on 429 / 5xx with exponential backoff + jitter and
propagates W3C trace context via `mashgate.WithTraceparent(ctx, ...)`.

**TypeScript / Python:** subscription, DLQ, and rotate-secret management are
currently Go-only; endpoint management and delivery retry are available on the
`webhooks` resource shown above.

## Events

HookLine delivers all platform events. Use the SDK constants instead of raw
strings. Both forms may be emitted during migration — dispatch should tolerate
either:

- **Legacy dotted** — `payment.created`, `payment.captured`, `refund.settled`,
  `checkout.completed`, … (`mashgate.EventPaymentCaptured` etc.), plus
  `wallet.credit` / `wallet.debit` / `wallet.transfer`.
- **Envelope v1** `<product>.<resource>.<verb>` — `payments.payment.created`,
  `payments.refund.completed`, `payments.checkout_session.completed`,
  `iam.user.registered`, `notifications.notification.sent`, … (the `Topic*`
  constants).

See [Service catalog](./service-catalog.md) for the full set and your contract
snapshot.

## Best practices

- **Verify the signature on every webhook** before parsing; reject
  unsigned/invalid payloads with 401. Never log the signing secret.
- **Make handlers idempotent.** The same event can be delivered more than once —
  dedupe on the event id (`CanonicalID()` in Go); a re-delivery is a no-op, not
  an error.
- **Return 2xx fast.** Acknowledge first, do slow work asynchronously; HookLine
  retries on non-2xx and dead-letters after exhausting attempts.
- **Don't assume ordering.** Design handlers so out-of-order delivery is safe.
- Read the raw body before JSON parsing — the signature covers the exact bytes.

More: [Best practices](../best-practices.md).

## See also

- [Wallets & ledger](./wallets-ledger.md) — source of `wallet.*` events.
- [KYC & compliance](./kyc-compliance.md) — source of `kyc.*` / `compliance.*` events.
- [Service catalog](./service-catalog.md) — mg-events, HookLine.
- [Building a vertical](../guides/building-a-vertical.md)
