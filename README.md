# Mashgate SDK

Official Mashgate SDK for Go, TypeScript, and Python.

> **Status:** initial extraction (v0.x). Targeting `v1.0.0` coordinated release — see [ROADMAP.md](ROADMAP.md).

**Mashgate** is a multi-tenant, API-first **Backend-as-a-Service** — the backend building blocks an application needs (identity, payments, wallets, billing, events, notifications, storage, and more) exposed as a coherent set of services behind one gateway, so product teams ship features instead of re-building plumbing. This repository is the single public entry point for integrating with Mashgate from your own application — whether that's a marketplace, commerce ops, or fintech product.

**Repository:** `github.com/saidmashhud/mashgate-public`
**License:** Apache 2.0

---

## What is Mashgate

Mashgate is the shared backend for a family of products. Instead of every app re-implementing auth, money movement, payments, webhooks, and the rest, Mashgate provides them as **multi-tenant building blocks** with a single, contract-first API. You provision a tenant, get an API key, and call the modules you need.

### Modules

| Domain | What it gives you |
|--------|-------------------|
| **Identity** (mgID) | Users, organizations & multi-tenancy, roles/permissions (IAM), API keys, OIDC, sessions, phone/OTP & WebAuthn auth. |
| **Payments** | Checkout sessions, a payments orchestrator, card processing, invoices, and payment links. |
| **Wallets & Ledger** | A double-entry ledger as the single source of truth for money, multi-currency wallets, atomic inter-wallet transfers — plus on-chain crypto wallets (Solana / TRON / EVM) via the chain layer. |
| **Billing & Subscriptions** | Subscription plans, metered usage, credits & promos, dunning. |
| **Events & Webhooks** | An event stream and outbound webhook delivery (HookLine) with signed payloads. |
| **Notifications** (mgNotify) | Transactional messaging across channels (email, SMS, push, messengers). |
| **Storage** (mgStorage) | S3-compatible object storage. |
| **Feature Flags** | Runtime flags & targeting. |
| **Risk & Compliance** | KYC, compliance screening, fraud checks, and a policy guard. |
| **Observability** | Structured logs, analytics, and usage metering per tenant. |

### How it fits together

- **Multi-tenant by default.** Every call is scoped to a tenant; isolation is enforced at the gateway (authorization) and in each service.
- **Contract-first.** The API is defined as Protobuf/OpenAPI in the core repo; this SDK is generated from those contracts (see [Contract source of truth](#contract-source-of-truth)), so client types track the server exactly.
- **gRPC microservices behind a gateway.** Internally Mashgate is a mesh of focused services; externally you talk to one API surface over REST/gRPC with an API key (and OIDC where applicable).
- **Money is handled carefully.** The ledger is the authoritative record; wallet, payment, and billing flows reconcile against it rather than holding their own truth.

> This SDK is the integration layer. The platform itself (deployment, infrastructure, internal service topology) lives in the private Mashgate core monorepo and is not part of this repo.

---

## Supported languages

| Language | Package | Min version | Status |
|----------|---------|-------------|--------|
| Go | `github.com/saidmashhud/mashgate-public/sdk/go` | Go 1.22 | stable (v0.x) |
| TypeScript | `@mashgate/sdk` (npm) | Node 18 | stable (v0.x) |
| Python | `mashgate-sdk` (PyPI) | Python 3.10 | stable (v0.x) |

---

## Quick start

### Go

```go
import "github.com/saidmashhud/mashgate-public/sdk/go"

client := mashgate.New("https://api.mashgate.uz", os.Getenv("MASHGATE_API_KEY"))
session, err := client.CreateCheckout(ctx, mashgate.CreateCheckoutRequest{
    Amount:   "100000", // minor units
    Currency: "UZS",
})
```

### Fintech Pack (Go)

```go
import "github.com/saidmashhud/mashgate-public/sdk/go/fintech"

fc := fintech.New("https://api.mashgate.uz", tenantID, apiKey)

check, err := fc.KYC.Request(ctx, fintech.RequestCheckRequest{
    SubjectID:   userID,
    SubjectType: fintech.KycSubjectIndividual,
    CheckType:   fintech.KycCheckFull,
}, idempotencyKey)
```

### TypeScript

```ts
import { MashgateClient } from "@mashgate/sdk";

const client = new MashgateClient({
  baseUrl: "https://api.mashgate.uz",
  apiKey: process.env.MASHGATE_API_KEY!,
});

const session = await client.checkout.create({ amount: "100000", currency: "UZS" });
```

### Python

```python
from mashgate import Mashgate

client = Mashgate(api_key=os.environ["MASHGATE_API_KEY"])
session = client.checkout.create(amount="100000", currency="UZS")
```

---

## Repository structure

```
mashgate-public/
├── sdk/
│   ├── go/                Core Mashgate SDK
│   │   └── fintech/       Fintech Pack (KYC, compliance, merchant, wallet)
│   ├── typescript/        Core Mashgate SDK (npm: @mashgate/sdk)
│   └── python/            Core Mashgate SDK (PyPI: mashgate-sdk)
├── docs/
│   ├── getting-started/   Per-language quickstarts
│   ├── modules/           Per-module guides (auth, payments, events, …)
│   └── migration/         Migration guides (from in-tree SDK, from hand-rolled clients)
├── contracts-sync/        Pinned snapshots + generators for protos/openapi/events
├── examples/              Working examples per language
├── tests/                 Cross-language contract + compat tests
└── tooling/               Generator wrappers + release scripts
```

See [`docs/specs/sdk-repository-separation.md`](https://github.com/saidmashhud/mashgate/blob/main/docs/specs/sdk-repository-separation.md) in the Mashgate core monorepo for the full RFC governing this repo.

---

## Which SDK to use

| If you're building... | Import |
|-----------------------|--------|
| A marketplace / commerce app on Mashgate (Zist, Vint) | Core `sdk/go` or `@mashgate/sdk` |
| A fintech/crypto app (Kiro) | Core SDK + `sdk/go/fintech` (Fintech Pack types) |
| Just webhook signature verification | Core SDK — `mashgate.VerifyWebhookSignature(...)` |
| HookLine webhook delivery | **Not this repo** — see [`github.com/saidmashhud/hookline`](https://github.com/saidmashhud/hookline) (separate product) |

The HookLine SDK lives in its own repo by design (ADR-0013 in the Mashgate core repo). It's a separate product surface with its own release cycle.

---

## Versioning + compatibility

Semantic versioning **per language**. A Go v1.3.0 does not require a matching TS v1.3.0.

Each release pins to a Mashgate contract snapshot (see `contracts-sync/manifests/`). Compatibility with Mashgate platform versions is tracked in [docs/compatibility-matrix.md](docs/compatibility-matrix.md).

- **Major** — breaking change; ≥90 day deprecation notice; 6-month security support after deprecation.
- **Minor** — new methods / modules, backwards compatible.
- **Patch** — bug fixes, docs, internals.

---

## Migration

Migrating from the in-tree SDK in the Mashgate monorepo (`mashgate/sdk/*`) or hand-rolled clients (Kiro's `internal/mashgate`, Zist/Vint ad-hoc)?

- [`docs/migration/from-monorepo-v0.x.md`](docs/migration/from-monorepo-v0.x.md) — from in-tree SDK
- [`docs/migration/from-handrolled-clients.md`](docs/migration/from-handrolled-clients.md) — from Kiro/Zist/Vint custom clients

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the PR process, language-specific build commands, and release flow.

---

## Contract source of truth

Contracts (protobuf, OpenAPI, event schemas) live in the **Mashgate core monorepo** at `mashgate/contracts/`. This repo consumes them via pinned snapshots (see `contracts-sync/`). Do **not** modify contracts here — open an issue/PR in the core repo instead.
