# Getting started — Python

The Python SDK gives you a synchronous, typed client (`httpx` under the hood) for
the core Mashgate flows: payments, checkout, identity/auth, wallets, webhooks, and
several BaaS modules.

> **Coverage note.** The Python SDK ships **fewer resources** than Go and
> TypeScript today. The modules currently available on the client are:
> `auth`, `payments`, `checkout`, `wallet`, `wallet_admin`, `risk`, `webhooks`,
> `developer`, `settings`, `chat`, `notify`, `storage`, `logs`, and `flags`.
> For anything outside that list — billing, subscriptions, invoices, payment
> links, guard, iam, metering, analytics, chain, local payments, mail — use the
> [Go](./go.md) or [TypeScript](./typescript.md) SDK, which cover the full surface.

## Install

```sh
pip install mashgate
```

This documents **v0.7.0** (Development Status: Beta). Requires **Python 3.9+**.
The only runtime dependency is `httpx`.

> The PyPI distribution name is `mashgate`; you import it as `mashgate`.

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

```python
import os
from mashgate import MashgateClient

mg = MashgateClient(
    base_url=os.environ["MASHGATE_BASE_URL"],
    api_key=os.environ["MASHGATE_API_KEY"],
)
```

The constructor is keyword-only and also accepts `access_token` (for end-user
auth), `timeout` (seconds, default `30.0`), and `headers`. The client supports
the context-manager protocol, so it closes its connection pool for you:

```python
with MashgateClient(base_url=..., api_key=...) as mg:
    ...
```

## Your first call

Create a hosted checkout session and redirect your customer to it. Note Python
takes `currency` plus a `line_items` list of dicts:

```python
import os
from mashgate import MashgateClient

mg = MashgateClient(
    base_url=os.environ["MASHGATE_BASE_URL"],
    api_key=os.environ["MASHGATE_API_KEY"],
)

session = mg.checkout.create_session(
    currency="UZS",
    success_url="https://example.com/success?session={sessionId}",
    cancel_url="https://example.com/cancel",
    line_items=[
        {"name": "Pro plan", "quantity": 1, "unitPrice": {"amount": "150000.00", "currency": "UZS"}},
    ],
)

print("redirect customer to:", session["checkoutUrl"])
```

Resource methods return parsed JSON as `dict`s, so read fields by key —
`session["sessionId"]`, `session["status"]`, `session["checkoutUrl"]`,
`session["totalAmount"]`, `session["expiresAt"]`.

### Or: a single payment

```python
payment = mg.payments.create(amount="150000.00", currency="UZS", order_id="order-001")
print(payment["paymentId"], payment["status"])
```

### Or: log in an end user

`auth.login` stores the returned access token on the client automatically, so
subsequent calls are sent as the authenticated user:

```python
data = mg.auth.login(email="user@example.com", password="secret")
# mg now sends Authorization: Bearer <accessToken>
profile = mg.auth.get_profile()
```

### Errors

Non-2xx responses raise `MashgateError` carrying `status` and `code`:

```python
from mashgate import MashgateError

try:
    mg.checkout.get_session("missing")
except MashgateError as e:
    print(e.status, e.code, e)
```

### Verify webhooks

```python
from mashgate import verify_webhook_signature

ok = verify_webhook_signature(
    body=raw_request_body,
    signature=request.headers["X-Mashgate-Signature"],
    secret=os.environ["WEBHOOK_SECRET"],
)
```

## Next steps

- [Building a vertical](../guides/building-a-vertical.md) — the end-to-end path:
  provision a tenant, wire modules, react to events, deploy.
- [Best practices](../best-practices.md) — idempotency, money/ledger as source of
  truth, multi-tenancy, webhook handling, error handling.
- [Service catalog](../modules/service-catalog.md) — the full module / RPC
  reference.
- Need a module the Python SDK doesn't expose yet? Reach for the
  [Go](./go.md) or [TypeScript](./typescript.md) quickstart.
