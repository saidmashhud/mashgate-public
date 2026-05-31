# Building a vertical on Mashgate

A "vertical" is your application (a marketplace, a fintech app, a dating app, …)
built **on top of** Mashgate. Mashgate is the shared backend (identity, payments,
wallets, chat, notifications, …); your vertical is a **separate service** that
calls Mashgate through this SDK and owns the domain logic that makes your product
your product.

> Mental model: Mashgate is the plumbing. You don't fork it or add columns to its
> tables — you run your own service and your own database alongside it, and call
> its modules. This is exactly how the first-party verticals are built.

---

## The shape of a vertical

```
your-app/
├── backend/
│   ├── mashgate/      ← thin wrapper around this SDK (client init, auth)
│   ├── events/        ← HookLine adapter (subscribe to / publish events)
│   ├── domain/        ← YOUR product logic + YOUR database
│   └── roles/         ← your domain roles (layered on top of Mashgate IAM)
│   └── migrations/    ← YOUR Postgres schema
├── frontend/
└── infra/             ← your own deploy (Helm, etc.)
```

Two databases, two responsibilities:

| Owns | Lives in | Examples |
|------|----------|----------|
| **Identity, money, platform primitives** | Mashgate | users, wallets, ledger, payments, KYC, chat, notifications |
| **Your domain** | Your DB | matches, listings, orders, swipes — keyed to Mashgate ids |

See [Data modeling & identity](./data-modeling-and-identity.md) for how the two
join (short version: key your tables by the Mashgate id; never `ALTER` Mashgate's
tables and never foreign-key into its database).

---

## Step 0 — Provision a tenant

Everything in Mashgate is multi-tenant. Get a **tenant** and an **API key** from
the Mashgate console/admin. You'll pass these on every call; the gateway scopes
and isolates by tenant.

## Step 1 — Initialize the SDK

```ts
import { MashgateClient } from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: process.env.MASHGATE_BASE_URL!,   // your tenant's API endpoint
  apiKey:  process.env.MASHGATE_API_KEY!,
});
```

```go
mg := mashgate.New(os.Getenv("MASHGATE_BASE_URL"), os.Getenv("MASHGATE_API_KEY"))
```

Keep this behind a small wrapper in your `mashgate/` package — one place to inject
the key, set timeouts, and attach the authenticated user/tenant to outgoing calls.

## Step 2 — Map your product needs to modules

Pick only what you need. Each is a resource on the client (`mg.auth`, `mg.checkout`,
`mg.chat`, …) — see the [service catalog](../modules/service-catalog.md).

| You need… | Module |
|-----------|--------|
| Sign-up / login, sessions, phone-OTP, OIDC | **Identity (mgID)** — `auth` |
| Identity / age verification, screening | **KYC / Compliance** (fintech pack) |
| Files, media, avatars | **Storage (mgStorage)** |
| In-app messaging | **Chat (mgChat)** — `chat` |
| Email / SMS / push | **Notifications (mgNotify)** |
| Checkout, subscriptions, invoices, payment links | **Payments / Billing** — `checkout`, `billing` |
| Multi-currency wallets, ledger, transfers, crypto | **Wallets & Ledger** — `chain` |
| Runtime flags, A/B | **Feature flags** |
| Fraud / abuse / policy | **Risk** — fraud, guard |
| React to platform events | **Events / Webhooks (HookLine)** — `events` |
| Per-tenant logs / analytics / metering | **Observability** — `analytics` |

## Step 3 — Build your domain

This is the part Mashgate does **not** provide. Model it in your own DB and code:
the matching engine, the order state machine, the listing search — whatever your
product is. Reference Mashgate entities by id (see the identity guide). Define your
own domain roles on top of Mashgate IAM (e.g. `free` / `premium` / `moderator`).

## Step 4 — Go event-driven where it pays off

Don't poll. Subscribe to platform events via **HookLine** and react:

```
payment.succeeded   → unlock the premium feature the user just bought
user.created        → create the matching profile row in your DB
wallet.credit       → update the user's coin balance view
```

Verify webhook signatures (the SDK helps); make handlers **idempotent** (the same
event may be delivered more than once).

## Step 5 — Deploy as your own service

Your vertical ships independently (its own image, its own Helm chart / infra) and
talks to the Mashgate gateway with its tenant API key. Mashgate stays the shared
backend; you scale and release on your own cadence.

---

## What to read next

- [Data modeling & identity](./data-modeling-and-identity.md) — how your tables
  join to Mashgate ids.
- [Best practices](../best-practices.md) — idempotency, money, multi-tenancy,
  webhooks, error handling.
- [Service catalog](../modules/service-catalog.md) — the full module/RPC list.
