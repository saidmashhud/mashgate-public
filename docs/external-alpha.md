# External Alpha Pack

This is the handoff contract for invited developers building on Mashgate with
the public SDKs and HookLine webhooks.

## What Is Included

- Go SDK: `github.com/saidmashhud/mashgate-public/sdk/go`
- TypeScript SDK: `@mashgate/sdk`
- Python SDK: `mashgate`
- Runnable Go, TypeScript, and Python quickstarts
- Runnable Go, TypeScript, and Python Exchange examples
- Pinned contract snapshot in `contracts-sync/manifests/active.yaml`
- HookLine webhook delivery and verifier SDKs

The platform itself is still invite-only. Tenant provisioning, API keys, and
webhook endpoint secrets are issued by the Mashgate operator.

## Required Gate

Run the external alpha gate from the repo root:

```bash
bash scripts/external-alpha-smoke.sh
```

It validates:

- Go SDK tests and Go quickstart compilation.
- TypeScript SDK install/build/tests and TypeScript quickstart compilation.
- Python SDK clean venv install/tests and Python quickstart syntax.
- Pinned contract snapshot and generated Go/TypeScript artifacts.
- Vendored Exchange proto/event snapshot and parity checks.
- HookLine Go/JS/Python webhook verifier gate from the sibling HookLine repo.
- This document and the webhook module docs are present.

If HookLine is not checked out beside this repo, either clone it or run with:

```bash
REQUIRE_HOOKLINE=0 bash scripts/external-alpha-smoke.sh
```

That is acceptable for SDK-only diagnosis, but not for an external handoff.

## Optional Full Contract Sync

The default gate only verifies that the pinned snapshot and generated artifacts
exist. Full regeneration is intentionally opt-in because it requires the private
Mashgate core repo and codegen tools:

```bash
RUN_CONTRACT_SYNC=1 \
MASHGATE_SRC=/path/to/mashgate \
bash scripts/external-alpha-smoke.sh
```

Run the general generator only in an isolated checkout. Sync the additive
Exchange proto and event schemas without changing the core checkout:

```bash
MASHGATE_SRC=/path/to/mashgate \
contracts-sync/scripts/sync-exchange.sh
CHECK_ONLY=1 MASHGATE_SRC=/path/to/mashgate \
contracts-sync/scripts/sync-exchange.sh
```

## Developer Handoff Checklist

Before handing the pack to an external developer:

- `scripts/external-alpha-smoke.sh` passes with HookLine enabled.
- `docs/compatibility-matrix.md` matches `contracts-sync/manifests/active.yaml`.
- The developer has a sandbox tenant id and scoped API key.
- The developer has a HookLine endpoint secret and knows who can rotate it.
- Webhook handlers verify `x-hl-signature`, dedupe event ids, and return 2xx
  quickly.
- Money-moving calls use stable idempotency keys.
- Exchange consumers pass no caller-selected account ownership and implement
  the inbox lifecycle described in `docs/modules/exchange.md`.

## Alpha Limits

- This is not self-service yet: tenant creation and key rotation are operator
  assisted.
- Package publishing is still per-language and may lag the repository state.
- Contract sync reads from the private core repo; external contributors should
  not edit generated files directly.
- HookLine delivery is at least once, not exactly once. Consumers must dedupe.

## GO / NO-GO

GO for invite-only alpha if the gate passes and the handoff checklist is
complete.

NO-GO if any SDK cannot install in a clean environment, HookLine verifier tests
fail, or the consumer cannot describe idempotent webhook handling.
