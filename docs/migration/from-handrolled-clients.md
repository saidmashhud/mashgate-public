# Migration - from Hand-Rolled Clients

Use this when a downstream product has its own Mashgate HTTP client, copied
types, or endpoint-specific wrappers. The goal is to move one flow at a time to
`mashgate-public`, without changing your product domain code in the same PR.

## When to migrate

Migrate when:

- You need upstream fixes, new endpoints, or webhook helpers without copying
  code.
- The integration will be maintained for more than a short experiment.
- The public SDK already exposes the flow you are about to implement.

Delay migration when:

- Your hand-rolled client owns product-specific orchestration that should stay
  in your service layer.
- You are in a hard release freeze and cannot run a focused regression pass.

## Go - from `internal/mashgate`

Before:

```go
import oldmg "my-service/internal/mashgate"

c := oldmg.New(baseURL, tenantID, apiKey)
check, err := c.KYC.Request(ctx, oldmg.RequestCheckRequest{...}, idempotencyKey)
```

After:

```go
import mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
import "github.com/saidmashhud/mashgate-public/sdk/go/fintech"

c := mashgate.NewWithTenant(baseURL, tenantID, apiKey)
check, err := c.KYC.Request(ctx, fintech.RequestCheckRequest{...}, idempotencyKey)
```

Step-by-step:

1. `cd backend && go get github.com/saidmashhud/mashgate-public/sdk/go@latest`
2. Replace the local client constructor with `mashgate.New(...)` for general
   API calls, or `mashgate.NewWithTenant(...)` for KYC, compliance, merchant,
   and wallet-ledger flows.
3. Import `sdk/go/fintech` only for fintech request and enum types.
4. Move domain-specific retry/idempotency decisions out of the old client and
   into your service layer.
5. Delete the local client package only after the migrated flow is covered by
   tests.

`fintech.New(baseURL, tenantID, apiKey)` still exists as a compatibility shim,
but new integrations should use `NewWithTenant` so one client owns the whole
Mashgate surface.

## TypeScript - from copied types or manual `fetch`

Before:

```ts
const response = await fetch(`${baseUrl}/v1/checkout/sessions`, {
  method: "POST",
  headers: { "Content-Type": "application/json", "X-API-Key": apiKey },
  body: JSON.stringify(body),
});
const session = await response.json();
```

After:

```ts
import { MashgateClient } from "@mashgate/sdk";

const mg = new MashgateClient({ baseUrl, apiKey });
const session = await mg.checkout.createSession(body);
```

Step-by-step:

1. `npm install @mashgate/sdk`
2. Replace copied DTO imports with exported SDK types where available.
3. Replace manual calls flow-by-flow: checkout, payments, webhooks, billing,
   IAM, mail, storage, and walletAdmin are first-class resources.
4. For KYC, compliance, and merchant flows, use the REST endpoint via
   `mg.request(...)` until typed TS resources land, or call them from a Go
   service using `NewWithTenant`.
5. Remove copied type packages after the last call site is migrated.

## Python - from ad-hoc `requests`

Before:

```python
import requests

resp = requests.post(
    f"{base_url}/v1/payments",
    headers={"X-API-Key": api_key},
    json={"amount": "150000.00", "currency": "UZS"},
)
payment = resp.json()
```

After:

```python
from mashgate import MashgateClient

mg = MashgateClient(base_url=base_url, api_key=api_key)
payment = mg.payments.create(amount="150000.00", currency="UZS")
```

The Python SDK exposes the same 25 resource namespaces as Go and TypeScript.
For flows that are intentionally not typed yet, such as KYC/Compliance/Merchant,
use the public `mg.request(...)` helper with the REST path and keep the
translation in one small adapter.

## Shared gotchas

### API key header

Tenant API keys authenticate as `X-API-Key`. Do not send them as
`Authorization: Bearer`; the gateway treats Bearer as a user JWT.

### Idempotency keys

Keep your existing idempotency scheme. Mutating money and provisioning calls
must receive a stable key derived from your domain id, for example
`checkout:<order_id>` or `wallet:<user_id>:create`.

### Event payload parsing

Do not parse webhook events before signature verification. Verify the raw body
using the SDK helper, then use `eventKey` / `eventPayload` helpers so both
Envelope v1 and legacy event shapes work during rollout.

### Trace propagation

Go fintech calls can carry trace context with `fintech.WithTraceparent(ctx, tp)`.
For TypeScript and Python, pass `traceparent` as an extra header when calling
lower-level request helpers until typed trace helpers are added.

## Rollback

Rollback is ordinary dependency rollback:

1. Restore the old local client package from git.
2. Remove the public SDK dependency from the service.
3. Open an issue in `mashgate-public` with the missing endpoint, wire mismatch,
   or SDK behavior that forced the rollback.
