# Changelog

All notable changes to Mashgate SDK are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning: per-language [SemVer](https://semver.org/spec/v2.0.0.html).

Per-language changelogs under `sdk/{go,typescript,python}/CHANGELOG.md`.
Top-level entry is the aggregate snapshot.

---

## [Unreleased]

(empty)

---

## [v1.1.0] — 2026-04-27

### Added — Wallet admin / on-chain (mgCrypto BaaS surface)

Closes the D-stream of [ADR-0016 — mgCrypto BaaS surface через ledger-core и chain-rpc](https://github.com/saidmashhud/mashgate/blob/main/docs/adr/0016-mgcrypto-baas-surface.md)
on the SDK side. All three language SDKs gained the full
`wallet.v1.WalletService` surface, typed Currency / Network / Mint
constants, and the new fields needed for SPL token deposits and
withdrawals.

- **Go (`sdk/go/fintech`)**:
  - New methods on `WalletService`: `CreateChain` (BIP-39 mnemonic on-chain wallet, L1), `Freeze`/`Unfreeze`, `GetTransaction`.
  - Extended methods: `DepositAddress(ctx, walletID, network Network, mint Mint)` — pass non-empty mint to receive the SPL Associated Token Account (L3); `WithdrawRequest.Mint` field replaces the legacy `mint=...;` description hack (L2).
  - New `fintech/types.go`: typed `Currency`, `Network`, `Mint` aliases with constants (`CurrencyUSDC`, `NetworkSolana`, `MintUSDCSolanaMainnet`, …). Wire format unchanged — string literals stay assignable.
  - 9 httptest-mock unit tests added.

- **TypeScript (`@mashgate/sdk`)**:
  - New `WalletAdminResource`, exposed as `client.walletAdmin` — same surface as Go fintech (full `wallet.v1.WalletService`).
  - `Currency` / `Network` / `Mint` / proto enum-strings as const-as-object + literal-union types — IDE autocomplete + compile-time check, plain strings on the wire.
  - 8 vitest mock-fetch tests added.
  - Pre-existing `client.test.ts` mocks fixed to include `Headers` (was failing on `response.headers.get()`).

- **Python (`mashgate-sdk`)**:
  - New `WalletAdminResource`, exposed as `client.wallet_admin` — same surface.
  - `(str, Enum)`-based typed constants: `Currency`, `Network`, `Mint`, plus `WalletStatus`, `WalletType`, `TransactionType`, `TransactionStatus`, `TransactionReason`. Inherit from `str` so json.dumps emits plain strings and `==` against plain strings holds.
  - First test suite for the SDK — 10 pytest+respx tests under `tests/`.

- **Cursor pagination on list-RPC**: `walletAdmin.list` and
  `walletAdmin.listTransactions` now propagate the opaque
  `next_cursor` returned by `ledger-core` — closes L6 of ADR-0016 on
  the consumer side.

- **READMEs**: Go `sdk/go/README.md` admin/merchant section; new
  `sdk/typescript/README.md`; new `sdk/python/README.md`. All include
  on-chain (mgCrypto) examples.

### Compatibility

- **Go**: callers using string literals (`"USDC"`, `"SOLANA"`) remain
  source-compatible — typed aliases auto-accept untyped literals.
  Direct assignment from `var s string` requires an explicit cast
  (`fintech.Network(s)`).
- **TS**: `Currency` / `Network` / `Mint` are TypeScript const objects
  + literal unions — wire format identical to plain strings; existing
  call sites compile unchanged.
- **Python**: `Currency.USDC` and `"USDC"` are interchangeable
  thanks to `(str, Enum)` inheritance. Existing call sites work
  unchanged.

### Notes

- **Kiro migration completed in-tree** — `kiro/backend/go.mod` uses a
  local `replace` directive against this repo until `v1.1.0` is
  tagged. Three call sites in `kiro/backend/internal/domain/{payment,settlement}.go`
  updated for the new `DepositAddress` signature and typed
  `Currency`/`Network`. After tag, the `replace` can be removed.

---

## [v1.0.0] — 2026-04-25

### Added
- Initial repo extraction from `github.com/saidmashhud/mashgate` monorepo.
- Go SDK: migrated `mashgate/sdk/go/` → `mashgate-public/sdk/go/`. Module path is now `github.com/saidmashhud/mashgate-public/sdk/go`.
- Go SDK: new **Fintech Pack** subpackage under `sdk/go/fintech/` — types + clients for `kyc`, `compliance`, `merchant`, `wallet` services. Sourced from Kiro's hand-rolled client to avoid divergence.
- TypeScript SDK: migrated `mashgate/sdk/typescript/` → `mashgate-public/sdk/typescript/`. npm package `@mashgate/sdk`.
- Python SDK: migrated `mashgate/sdk/python/` → `mashgate-public/sdk/python/`. PyPI package `mashgate-sdk`.
- Repository scaffolding: README, LICENSE (Apache 2.0), CONTRIBUTING, ROADMAP, docs/, examples/, contracts-sync/, tests/, tooling/, .github/ workflows placeholders.

### Changed
- All Go imports rewritten to `github.com/saidmashhud/mashgate-public/sdk/go`.

### Planned — v1.0.0
- Coordinated release (`sdk/go/v1.0.0`, `sdk/typescript/v1.0.0`, `sdk/python/v1.0.0`).
- Compatibility matrix published.
- Migration guide for existing consumers (in-tree SDK, Kiro hand-rolled).
- Compatibility bridge in core monorepo `mashgate/sdk/` re-exporting from public packages with deprecation notices; grace window ≥6 months.
