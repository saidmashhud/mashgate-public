# Best practices

Patterns that keep a Mashgate integration correct, safe, and easy to evolve.
Most bugs in verticals come from ignoring one of these.

---

## 1. Let Mashgate own what it owns

- **Identity** is owned by mgID. Key your tables by `user_id`; don't duplicate
  email/phone/roles as your source of truth. → [Data modeling & identity](./guides/data-modeling-and-identity.md)
- **Money** is owned by the ledger. The ledger is the authoritative record of
  balances and movements. Don't keep your own "authoritative" balance column —
  read from Mashgate and reconcile via events. Treat any local balance as a cache.
- **Contracts** are owned by the core monorepo. Don't edit the generated SDK types
  or vendored protos here — open a PR upstream and re-sync. → [Contract source of truth](../README.md#contract-source-of-truth)

## 2. Make money operations idempotent

Any call that moves money or creates a resource takes an **idempotency key**.
Generate one per logical operation and reuse it on retries so a network blip never
double-charges or double-creates:

```go
key := idempotencyKeyFor(orderID)          // stable per operation, not per attempt
resp, err := mg.Checkout.Create(ctx, req, key)
```

- Derive the key from your domain (`order:123:charge`), not from `now()`/random.
- Persist it with the operation so retries — including after a crash — reuse it.

## 3. Be event-driven, and make handlers idempotent

Prefer reacting to platform events over polling. Subscribe via **HookLine**:

- **Verify the signature** on every webhook (the SDK provides
  `VerifyWebhookSignature`). Reject unsigned/invalid payloads.
- **Handlers must be idempotent** — the same event can be delivered more than once.
  Dedupe on the event id; an already-processed event is a no-op, not an error.
- **Return 2xx fast.** Acknowledge, then do slow work asynchronously; HookLine
  retries on non-2xx.
- Don't assume ordering. Design handlers so out-of-order delivery is safe.

## 4. Respect multi-tenancy

Every entity is scoped to a tenant. Carry `tenant_id` through your own data and
make it part of your uniqueness constraints (`UNIQUE (tenant_id, …)`). Never let
one tenant's request read or write another tenant's data — the gateway enforces
isolation on the Mashgate side; enforce it on yours too.

## 5. Take identity from the token, not the request

The authenticated `user_id` / `tenant_id` come from the validated JWT / request
context that the gateway populates. Never trust an id supplied in a request body
for authorization decisions.

## 6. Handle errors and retries deliberately

- **Retry only safe failures** — timeouts and 5xx, with backoff, and only when the
  call is idempotent (see #2). Don't blindly retry a non-idempotent write.
- **Don't retry 4xx** — fix the request. `PERMISSION_DENIED` usually means a tenant
  mismatch or a missing role, not a transient error.
- Surface Mashgate error codes to your own layer rather than swallowing them.

## 7. Keep secrets and keys safe

- API keys are tenant credentials — inject from the environment / a secret store,
  never commit them, never ship them to the browser. Frontend calls go through your
  backend, not directly with the tenant key.
- Rotate keys on a schedule; scope per environment (dev / staging / prod).

## 8. Pin and track versions

- This SDK is versioned **per language** (a Go `v1.x` ≠ a TS `v1.x`). Pin the
  version you build against.
- Each SDK release maps to a Mashgate contract snapshot
  (`contracts-sync/manifests/active.yaml`). Review it before upgrading the platform
  or the SDK. → [Versioning + compatibility](../README.md#versioning--compatibility)

## 9. Don't reach across the database boundary

No foreign keys from your DB into Mashgate's, and no direct DB access to Mashgate.
The only supported interface is the API/SDK. Cross-service integrity lives in your
application logic (and in events), not in shared SQL.

---

See also: [Building a vertical](./guides/building-a-vertical.md) ·
[Data modeling & identity](./guides/data-modeling-and-identity.md) ·
[Service catalog](./modules/service-catalog.md)
