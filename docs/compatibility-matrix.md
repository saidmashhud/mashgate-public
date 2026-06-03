# Compatibility matrix

SDKs version **independently per language** (semver). Each SDK release pins to a
Mashgate contract snapshot (`contracts-sync/manifests/`); any SDK on a given
contract snapshot is compatible with any Mashgate platform version that serves
that contract major.

| SDK | Package | Min runtime | Status | Contract snapshot |
|-----|---------|-------------|--------|-------------------|
| Go | `github.com/saidmashhud/mashgate-public/sdk/go` | Go 1.22 | stable (v0.x) | v1 |
| TypeScript | `@mashgate/sdk` (npm) | Node 18 | stable (v0.x) | v1 |
| Python | `mashgate` (PyPI) | Python 3.9+ | stable (v0.x) | v1 |

## Module coverage

Go and TypeScript expose the full v1 surface (25 resources). Python currently
ships a core subset (auth, payments, wallet, wallet_admin, checkout, webhooks,
notify, storage, chat, flags, logs, risk, developer, settings); the remaining
resources (billing, subscriptions, invoices, iam, analytics, metering, mail,
guard, chain, local_payments, payment_links) are being ported — track the
Python SDK README for current coverage.

## Contract snapshots

A platform contract major (`v1`) is forward-compatible: new RPCs/fields are
additive. An SDK pinned to a `v1` snapshot keeps working as the platform adds
`v1` RPCs; it only needs an upgrade to call the newly-added ones. A platform
`v2` would ship alongside `v1` (no hard break) until `v1` is deprecated.
