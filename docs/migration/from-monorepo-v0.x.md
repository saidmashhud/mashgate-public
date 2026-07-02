# Migration - from the In-Tree SDK

Older Mashgate consumers imported SDK code from the private `mashgate` monorepo.
New integrations should depend on `mashgate-public` instead. The public repo is
the release and compatibility boundary for Go, TypeScript, and Python.

## Go

Before:

```go
import mashgate "github.com/saidmashhud/mashgate/sdk/go"
```

After:

```go
import mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
```

For tenant-scoped KYC, compliance, merchant, or wallet-ledger flows:

```go
import mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
import "github.com/saidmashhud/mashgate-public/sdk/go/fintech"

mg := mashgate.NewWithTenant(baseURL, tenantID, apiKey)
check, err := mg.KYC.Request(ctx, fintech.RequestCheckRequest{...}, idempotencyKey)
```

Run:

```sh
go get github.com/saidmashhud/mashgate-public/sdk/go@latest
go mod tidy
go test ./...
```

## TypeScript

Before:

```ts
import { MashgateClient } from "@mashgate/internal-sdk";
```

After:

```ts
import { MashgateClient } from "@mashgate/sdk";
```

Run:

```sh
npm install @mashgate/sdk
npm run build
npm test
```

## Python

Before:

```python
from mashgate_sdk import Client
```

After:

```python
from mashgate import MashgateClient
```

Run:

```sh
pip install mashgate
pytest
```

## Checks before deleting the old dependency

- Auth/login flow still stores the returned access token where your app expects.
- API-key calls send `X-API-Key`, not `Authorization: Bearer`.
- Mutating calls pass an idempotency key.
- Webhook handlers verify the raw body before JSON parsing.
- Generated/copy-pasted DTO packages have no remaining imports.
