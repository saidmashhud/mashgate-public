# Exchange Testnet Runbook

## Preconditions

- Separate Exchange tenant, OIDC audience, database, HookLine subscription,
  custody secret, and operator account.
- Software custody adapter only; no production keys or real funds.
- `contracts-sync/manifests/exchange-alpha.yaml` points to the deployed core
  contract ref and `sync-exchange.sh` passes in check-only mode.

## Smoke sequence

1. Authenticate an allowlisted, KYC-approved account and complete 2FA.
2. Resolve one deposit address per supported rail.
3. Ingest a test deposit through detection, confirmations, and one ledger
   credit. Replay the same observation and verify no second credit.
4. Place and cancel LIMIT orders, execute partial and complete fills, and replay
   each command with its original idempotency key.
5. Request a withdrawal, verify the ledger hold, escalate/resolve compliance,
   approve, sign, broadcast, and reconcile.
6. Replay HookLine events in reordered and duplicate form. Verify one domain
   mutation per event id.
7. Restart trading-core from its journal, restore databases into disposable
   instances, and run ledger/custody/deposit/withdrawal/trade reconcilers.

## Stop conditions

Stop the market maker and block all money mutations on stale reference prices,
an unexplained balance, an unsettled fill, an orphan hold, a reorg review, a
stuck withdrawal, failed reconciliation, or cross-tenant authorization result.

The MM kill endpoint is internal and requires its dedicated admin bearer. A
kill must cancel all open MM orders before the service is considered stopped.

## Real-money no-go

Do not promote the testnet software custody adapter. MPC/HSM custody, hot/cold
treasury policy, maker-checker withdrawal approvals, legal review, and a signed
recovery report are independent mandatory gates.
