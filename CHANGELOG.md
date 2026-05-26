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

## [2026-05-19] — fee abstraction Phase 1 (sponsored withdrawals)

- **`sdk/go/v1.11.0`** — new `WithdrawRequest.SponsorWalletID`.
- **`@mashgate/sdk@1.5.0`** — new `InitiateWithdrawalRequest.sponsor_wallet_id`.
- **`mashgate@0.5.0`** — new `withdraw(..., sponsor_wallet_id=...)`.

Server (`mashgate@TBD`): dual-signer Solana txs. New
`chain.internal.v1.BuildSolanaTransferTxRequest.fee_payer_pubkey_base58`
selects между legacy single-signer (source pays fee) и sponsor variant
(sponsor signs first, pays fee + ATA rent; source signs second,
authorises movement). 4 new builders в `wallet/solana_tx.rs` (SOL + SPL,
sponsor flavour each).

Phase 1 limits: sponsor pays the fee on-chain but не получает явный
off-chain debit movement yet — that's Phase 2 (status-sync worker captures
actual fee_lamports on confirmation and emits a separate sponsor-debit
movement).

Per-language details:
- [`sdk/go/CHANGELOG.md`](sdk/go/CHANGELOG.md#v1110--2026-05-19--sdkgov1110).
- [`sdk/typescript/CHANGELOG.md`](sdk/typescript/CHANGELOG.md#150--2026-05-19).
- [`sdk/python/CHANGELOG.md`](sdk/python/CHANGELOG.md#050--2026-05-19).

---

## [2026-05-19] — wallet recovery polish (mnemonic_hash + was_existing + cross-subject deny)

- **`sdk/go/v1.10.0`** — 💥 BREAKING — `(*WalletService).ImportChain` returns
  `*ImportChainWalletResponse` instead of `*Wallet`.
- **`@mashgate/sdk@1.4.0`** — new `client.walletAdmin.importChain(req)`.
- **`mashgate@0.4.0`** — new `client.wallet_admin.import_chain(*, ...)`.

Server (`mashgate@TBD`): `wallet.v1.WalletService.ImportChainWallet` polished:

- Pre-flight lookup by `(tenant_id, mnemonic_hash)` before deriving keys —
  cross-subject mnemonic reuse returns `PERMISSION_DENIED` immediately, без
  exposing seed material через side channels.
- Idempotent recovery: re-import same phrase для того же subject = same
  wallet с `was_existing=true`.
- Audit log entry on every call (`wallet.imported`, `wallet.imported_reused`,
  `wallet.import_denied`).
- New `ImportChainWalletResponse { wallet, was_existing, recovered_at }`.

Per-language details:
- [`sdk/go/CHANGELOG.md`](sdk/go/CHANGELOG.md#v1100--2026-05-19--sdkgov1100).
- [`sdk/typescript/CHANGELOG.md`](sdk/typescript/CHANGELOG.md#140--2026-05-19).
- [`sdk/python/CHANGELOG.md`](sdk/python/CHANGELOG.md#040--2026-05-19).

---

## [2026-05-15] — wallet.transfer (atomic inter-wallet movement)

- **`sdk/go/v1.9.0`** — `(*WalletService).Transfer(ctx, TransferRequest, idempotencyKey)`.
- **`@mashgate/sdk@1.3.0`** — `client.walletAdmin.transfer(fromWalletId, req)`.
- **`mashgate@0.3.0`** — `client.wallet_admin.transfer(from_wallet_id, *, ...)`.

All three SDKs mirror new `wallet.v1.WalletService.TransferBetweenWallets`
RPC (ledger-core handler в `mashgate@af67d653`, same date). Same-currency
v1 — cross-currency FX is out of scope until a future Convert RPC.

Server commits one Postgres tx: balance delta on both wallets + two
`wallet_transactions` rows (debit on source, credit on dest, both
referencing the synthetic `transfer_id`) + three outbox events
(`wallet.debit`, `wallet.credit`, `wallet.transfer`).

Per-language details:
- [`sdk/go/CHANGELOG.md`](sdk/go/CHANGELOG.md#v190--2026-05-15--sdkgov190).
- [`sdk/typescript/CHANGELOG.md`](sdk/typescript/CHANGELOG.md#130--2026-05-15).
- [`sdk/python/CHANGELOG.md`](sdk/python/CHANGELOG.md#030--2026-05-15).

---

## [sdk/go/v1.3.0] — 2026-05-04

### Added — Go SDK: `auth.Register` + `auth.SendOtp` + `auth.VerifyOtp` + extended `TokenPair`

Unblocks proper phone-auth path в downstream verticals — qr-app первый
консьюмер. Replaces ad-hoc `pseudoEmail/pseudoPass + raw HTTP` workaround.

- **Go (`sdk/go`)**:
  - `(*Client).Register(ctx, RegisterRequest) (*RegisterResponse, error)` →
    `POST /v1/auth/register`.
  - `(*Client).SendOtp(ctx, SendOtpRequest) error` → `POST /v1/auth/otp/send`.
  - `(*Client).VerifyOtp(ctx, VerifyOtpRequest) (bool, error)` →
    `POST /v1/auth/otp/verify`.
  - `TokenPair` extended с `ExpiresAt int64` + `User *AuthUserInfo` (optional;
    backward-compat).
  - 4 new httptest-mock tests.

См. подробно: [`sdk/go/CHANGELOG.md`](sdk/go/CHANGELOG.md#v130--2026-05-04--sdkgov130).

**SECURITY NOTE**: Mashgate upstream `/v1/auth/register` сейчас принимает
`role` field без sanitization — known upstream bug, fix запланирован
(whitelist допустимых ролей при self-service signup). Используйте
`Role: "merchant"` для customer flows; admin-grade роли пусть выдаёт
admin через `IamService.AssignRole`.

---

## [sdk/go/v1.2.0] — 2026-04-30

### Added — Go SDK only: `iam.ListTenants`

Closes Phase C of [ADR-0020 — Tenant identity SoT в Mashgate IAM](https://github.com/saidmashhud/mashgate/blob/main/docs/adr/0020-tenant-identity-sot.md)
на стороне Go SDK (TypeScript SDK уже имел эту функциональность в v1.1.0).

- **Go (`sdk/go`)**:
  - New `(*Client).ListTenants(ctx, opts *ListTenantsOptions) ([]Tenant, error)` —
    `GET /v1/iam/tenants`. Used by downstream verticals для cold-start backfill
    перед subscription к Kafka `tenant-events` (см. ADR-0020 Phase B.3).
  - New types `Tenant` + `ListTenantsOptions` (Status/Search/Page/PageSize/SortBy/SortOrder).
  - 3 httptest-mock unit tests added.

### Compatibility

Pure additive. Existing IAM methods (`ListAPIKeys`, `CreateAPIKey`,
`DeleteAPIKey`, `CheckPermission`) unchanged.

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
