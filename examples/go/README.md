# Mashgate Go SDK — quickstart

A single runnable file (`main.go`) covering the three things every integration
needs: client init from env, creating a hosted checkout session + a single
payment, and a signature-verifying webhook receiver.

## Run

```sh
export MASHGATE_BASE_URL="https://api.mashgate.uz"   # or http://localhost:9661 for local dev
export MASHGATE_API_KEY="mg_test_..."

# Checkout + payment demo — prints the redirect URL and ids
go run .
```

```sh
# Webhook receiver — verifies signatures, listens on :8080/webhooks
export MASHGATE_WEBHOOK_SECRET="whsec_..."   # SigningSecret from your endpoint
go run . serve
```

`MASHGATE_API_URL` is accepted as an alias for `MASHGATE_BASE_URL`.

## How it links to the SDK

`go.mod` uses a `replace` directive to build against the SDK source in this repo
(`../../sdk/go`). When you copy this example into your own project, drop the
`replace` line and pull the SDK from GitHub instead:

```sh
go get github.com/saidmashhud/mashgate-public/sdk/go@main
```

```go
import mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
```

See [docs/getting-started/go.md](../../docs/getting-started/go.md) for the full
walkthrough.
