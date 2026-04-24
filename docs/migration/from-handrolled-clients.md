# Migration — from Hand-Rolled Clients

If your product currently has a hand-rolled Mashgate client (like Kiro's
`backend/internal/mashgate/` or `frontend/packages/mashgate-types/`, or
Zist/Vint ad-hoc HTTP calls), this is the migration path.

---

## When to migrate

Migrate when:
- You want upstream fixes / new endpoints without copying them.
- You're about to implement a flow that the public SDK already has.
- You're going to maintain at least one service consuming Mashgate for > 6 months.

Delay migration when:
- Your hand-rolled client has domain-specific logic not in the public SDK.
- You're in a hard feature-freeze and risk-averse.

---

## Go — from `internal/mashgate` (Kiro pattern)

Before:

```go
// backend/internal/mashgate/client.go
import "internal/mashgate"

c := mashgate.New(baseURL, tenantID, apiKey)
check, err := c.KYC.Request(ctx, mashgate.RequestCheckRequest{...}, idempotencyKey)
```

After:

```go
import "github.com/saidmashhud/mashgate-public/sdk/go/fintech"

c := fintech.New(baseURL, tenantID, apiKey)
check, err := c.KYC.Request(ctx, fintech.RequestCheckRequest{...}, idempotencyKey)
```

### Step-by-step

1. `cd backend && go get github.com/saidmashhud/mashgate-public/sdk/go/fintech@latest`
2. Replace imports: `sed -i '' 's|internal/mashgate|github.com/saidmashhud/mashgate-public/sdk/go/fintech|g' $(grep -rl internal/mashgate --include='*.go')`
3. Rename package usage: `mashgate.` → `fintech.` in call sites.
4. Delete `internal/mashgate/` directory.
5. `go mod tidy && go build ./... && go test ./...`.

### What moved

| Kiro hand-rolled symbol | Public SDK equivalent |
|-------------------------|------------------------|
| `internal/mashgate.Client` | `fintech.Client` |
| `internal/mashgate.KYCService` | `fintech.KYCService` |
| `internal/mashgate.ComplianceService` | `fintech.ComplianceService` |
| `internal/mashgate.MerchantService` | `fintech.MerchantService` |
| `internal/mashgate.WalletService` | `fintech.WalletService` |
| All enum string consts (KycStatusPassed, MerchantStatusAccepted, …) | Same names under `fintech.*` |
| `WithTraceparent(ctx, tp)` | `fintech.WithTraceparent(ctx, tp)` |

Public SDK adds:
- Automatic retry with exponential backoff (planned v0.2).
- Trace context helpers integrated with OTel (planned v1.1).

---

## TypeScript — from `@kiro/mashgate-types`

Before:

```ts
import type * as kyc from "@kiro/mashgate-types/kyc";
import type * as wallet from "@kiro/mashgate-types/wallet";

// types only — client was hand-rolled per-service
const response: kyc.RequestCheckResponse = await fetch(...);
```

After:

```ts
import { MashgateClient, fintech } from "@mashgate/sdk";

const client = new MashgateClient({ baseUrl, apiKey, tenantId });
const response = await client.fintech.kyc.request({ ... });
// response is typed as fintech.RequestCheckResponse
```

### Step-by-step

1. `pnpm remove @kiro/mashgate-types && pnpm add @mashgate/sdk`
2. Replace `import type * as X from "@kiro/mashgate-types/X"` with `import type { X } from "@mashgate/sdk"` (or use namespaced `fintech.X`).
3. Replace manual `fetch` / axios calls to Mashgate with `client.*` methods.
4. Delete `packages/mashgate-types/` directory (no longer needed).
5. Update `pnpm-workspace.yaml` to remove the deleted package.

---

## Python — from ad-hoc `requests` usage

Before:

```python
import requests

resp = requests.post(
    f"{MASHGATE_API}/v1/kyc/checks",
    headers={"Authorization": f"Bearer {api_key}", "X-Tenant-ID": tenant_id},
    json={...},
)
```

After:

```python
from mashgate import Mashgate

client = Mashgate(api_key=api_key, tenant_id=tenant_id)
check = client.fintech.kyc.request(subject_id="...", subject_type="individual", ...)
```

---

## Shared gotchas

### Idempotency keys

Keep your existing idempotency key scheme. The public SDK forwards whatever you pass in — it doesn't auto-generate keys (by design, to avoid silent double-submit masking).

### Status enum values

Public SDK uses the same proto enum string values as Mashgate wire format (`KYC_STATUS_PASSED`, `MERCHANT_STATUS_ACCEPTED`, etc.). If your hand-rolled client used different internal names, you'll need a mapping layer in your domain code.

### Event payload parsing

Public SDK includes envelope v1 type definitions. If you were parsing envelope payloads manually, switch to `mashgate.EnvelopeV1<YourPayloadType>` (Go) / `EnvelopeV1<YourPayloadType>` (TS) / `mashgate.EnvelopeV1` (Python).

### Trace propagation

Public SDK Go: `ctx = fintech.WithTraceparent(ctx, r.Header.Get("traceparent"))`
Public SDK TS: `client.withTrace(traceparent).fintech.kyc.request(...)`
Public SDK Python: `client.set_trace(traceparent)`

All three emit `traceparent` + `tracestate` on outbound requests, matching HookLine webhook delivery conventions.

---

## Rollback

If migration causes a regression, revert by:
1. Restore `internal/mashgate/` or `packages/mashgate-types/` from git.
2. Remove the public SDK dependency.
3. Open an issue in this repo describing the incompatibility.

The public SDK won't break your hand-rolled client — it's just a separate code path. Migrations can be incremental: migrate one service at a time.
