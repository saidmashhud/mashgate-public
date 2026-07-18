# Exchange Pack

The Exchange Pack is an invite-only custodial spot API. It exposes three
markets (`BTC/USDT`, `ETH/USDT`, `SOL/USDT`) and keeps authoritative balances
in Mashgate ledger-core. Product databases must not mirror a writable balance.

## Authentication and ownership

Use both a tenant API key and an end-user access token. The token determines
tenant, subject, and Exchange account. Order, deposit, withdrawal, merchant,
or wallet ownership must never come from a frontend-supplied account id.

Funding and trading require an allowlisted account with completed KYC and 2FA.
An account suspended by compliance cannot place orders or request withdrawals.

## Commands

The API supports `LIMIT GTC`, `LIMIT IOC`, `MARKET IOC`, `post_only`, partial
fills, and cancel. Every order, cancel, and withdrawal requires an
`Idempotency-Key`. The scope is tenant + subject + operation + key. Repeating
an identical request returns the original result; changing the payload with
the same key returns `409`.

See `examples/exchange/` for Go, TypeScript, and Python clients.

## Funding lifecycle

Testnet rails cover BTC native, ETH mainnet-shaped testnet, SOL, and USDT
TRC-20-shaped testnet. Deposits move from detected to confirming to credited.
Reorged credits enter `reorg_review` and suspend the account for operator
reconciliation.

A withdrawal reserves funds first, then passes compliance and optional
maker-checker review before custody signing and broadcast. Rejection releases
the hold. Clients must display the server status, not infer completion from an
HTTP acceptance response.

## HookLine events

- `exchange.order.*`
- `exchange.trade.*`
- `exchange.deposit.*`
- `exchange.withdrawal.*`
- `exchange.market.*`

Delivery is at least once. Persist each envelope in an inbox as `received`,
mark `processing` before work, and write `processed` only in or after the domain
commit. Failed events remain retryable and visible to the DLQ process.

## Production boundary

The current pack is testnet alpha. Real-money use is prohibited until all
asset parsers/checkpoints and reorg handling pass recovery drills, custody uses
an approved MPC/HSM adapter with hot/cold separation and treasury caps, and
legal/compliance approval is recorded.
