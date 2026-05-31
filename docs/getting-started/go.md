# Getting started — Go

The Go SDK is the most complete Mashgate client: payments, checkout, identity,
wallets, webhooks, and the full BaaS suite (chat, notify, storage, flags, logs,
guard, invoices, subscriptions, analytics, and more), plus a `fintech` subpackage
for KYC / compliance / merchant / wallet-admin flows.

## Install

No package registry — import straight from GitHub. Requires **Go 1.22+**.

```sh
go get github.com/saidmashhud/mashgate-public/sdk/go@main
```

```go
import mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
```

## Prerequisites

Everything in Mashgate is multi-tenant. Before your first call you need a
**tenant** and an **API key** from the Mashgate console — see
[Building a vertical → Step 0](../guides/building-a-vertical.md#step-0--provision-a-tenant).

Keys are environment-scoped: `mg_test_...` for sandbox, `mg_live_...` for
production. Keep the key out of source — read it from the environment:

```sh
export MASHGATE_BASE_URL="https://api.mashgate.uz"   # or http://localhost:9661 for local dev
export MASHGATE_API_KEY="mg_test_..."
```

## Initialize the client

Construct a client once and reuse it (it holds a pooled `*http.Client`):

```go
import (
    "os"

    mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
)

client := mashgate.New(
    os.Getenv("MASHGATE_BASE_URL"),
    os.Getenv("MASHGATE_API_KEY"),
)
```

Prefer functional options? `NewClient` defaults the base URL to
`https://api.mashgate.uz` and lets you tune timeouts and retries:

```go
import "time"

client := mashgate.NewClient(os.Getenv("MASHGATE_API_KEY"),
    mashgate.WithTimeout(10*time.Second),
    mashgate.WithMaxRetries(3),
)
```

## Your first call

Create a hosted checkout session and redirect your customer to the returned URL.
`CreateCheckout` auto-generates an idempotency key when you don't supply one, so
a retried request never double-charges.

```go
package main

import (
    "context"
    "log"
    "os"

    mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
)

func main() {
    client := mashgate.New(
        os.Getenv("MASHGATE_BASE_URL"),
        os.Getenv("MASHGATE_API_KEY"),
    )

    session, err := client.CreateCheckout(context.Background(), mashgate.CreateCheckoutRequest{
        TotalAmount:   mashgate.Money{Amount: "150000.00", Currency: "UZS"},
        SuccessURL:    "https://example.com/success?session={sessionId}",
        CancelURL:     "https://example.com/cancel",
        CustomerEmail: "customer@example.com",
        Items: []mashgate.LineItem{
            {
                Name:      "Pro plan",
                Quantity:  1,
                UnitPrice: mashgate.Money{Amount: "150000.00", Currency: "UZS"},
            },
        },
    })
    if err != nil {
        log.Fatalf("create checkout: %v", err)
    }

    log.Printf("redirect customer to: %s", session.CheckoutURL)
}
```

The returned `*mashgate.CheckoutSession`:

```go
type CheckoutSession struct {
    SessionID     string        // "cs_..."
    Status        string        // pending | completed | expired | cancelled
    TotalAmount   mashgate.Money
    CheckoutURL   string        // redirect the customer here
    SuccessURL    string
    CancelURL     string
    CustomerEmail string
    ExpiresAt     int64         // unix seconds
    CreatedAt     int64
}
```

### Or: start a phone-OTP flow

```go
err := client.SendOtp(context.Background(), mashgate.SendOtpRequest{
    Phone:   "+998901234567",
    Purpose: "login", // "login" | "password_reset" | "phone_verify"
})
// ... user enters the code ...
ok, err := client.VerifyOtp(context.Background(), mashgate.VerifyOtpRequest{
    Phone:   "+998901234567",
    Code:    "123456",
    Purpose: "login",
})
```

### Errors

Every call returns a typed `*mashgate.Error` carrying the machine-readable
`Code`, the `RequestID` (include it in support requests), and a `DocURL()`:

```go
import "errors"

session, err := client.CreateCheckout(ctx, req)
if err != nil {
    var mgErr *mashgate.Error
    if errors.As(err, &mgErr) {
        log.Printf("code=%s request_id=%s doc=%s", mgErr.Code, mgErr.RequestID, mgErr.DocURL())
    }
}
```

## Next steps

- [Building a vertical](../guides/building-a-vertical.md) — the end-to-end path:
  provision a tenant, wire modules, react to events, deploy.
- [Best practices](../best-practices.md) — idempotency, money/ledger as source of
  truth, multi-tenancy, webhook handling, error handling.
- [Service catalog](../modules/service-catalog.md) — the full module / RPC
  reference.
- Fintech Pack — for KYC, compliance, merchant, and admin wallet flows, import
  `github.com/saidmashhud/mashgate-public/sdk/go/fintech` and construct it with
  `fintech.New(baseURL, tenantID, apiKey)`. See the
  [Go SDK README](../../sdk/go/README.md) for examples.
