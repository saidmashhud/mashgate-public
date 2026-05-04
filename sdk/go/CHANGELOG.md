# Changelog — Go SDK

`github.com/saidmashhud/mashgate-public/sdk/go`

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) ·
Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

Aggregate changelog for all languages: [`../../CHANGELOG.md`](../../CHANGELOG.md).

---

## [v1.3.0] — 2026-05-04 — `sdk/go/v1.3.0`

### Added — `auth` resource: Register, SendOtp, VerifyOtp + extended TokenPair

Closes proper phone-auth path для downstream verticals (qr-app first consumer).
Replaces ad-hoc `pseudoEmail/pseudoPass + raw HTTP` workaround в qr-app
backend.

- `(*Client).Register(ctx, RegisterRequest) (*RegisterResponse, error)` —
  `POST /v1/auth/register`. New types `RegisterRequest` (snake_case body
  fields) + `RegisterResponse`. **SECURITY NOTE** (см. models.go): Mashgate
  upstream `/v1/auth/register` сейчас не валидирует `role` field — known
  upstream bug, будет закрыто whitelist'ом. Use `Role: "merchant"` для
  customer flows.
- `(*Client).SendOtp(ctx, SendOtpRequest) error` — `POST /v1/auth/otp/send`.
  New type `SendOtpRequest{UserID|Phone, Purpose}`. Purpose ∈ {"login",
  "password_reset", "phone_verify"}.
- `(*Client).VerifyOtp(ctx, VerifyOtpRequest) (bool, error)` —
  `POST /v1/auth/otp/verify`. New type `VerifyOtpRequest{UserID, Code, Purpose}`.
- **`TokenPair` extended** — added optional `ExpiresAt int64` + `User *AuthUserInfo`
  (UserID/Email/FullName/TenantID/Role/Roles). Backward-compatible: existing
  consumers ignore new fields; populated only when upstream response includes
  them. New type `AuthUserInfo`.
- 4 httptest-mock unit tests: Login parses User/ExpiresAt, Register sends
  snake_case body, SendOtp posts expected body, VerifyOtp returns valid bool.

---

## [v1.2.0] — 2026-04-30 — `sdk/go/v1.2.0`

### Added — `iam` resource: ListTenants

Closes Phase C of [ADR-0020 — Tenant identity SoT в Mashgate IAM](https://github.com/saidmashhud/mashgate/blob/main/docs/adr/0020-tenant-identity-sot.md)
на стороне Go SDK. Mirrors TypeScript SDK `iam.listTenants()` shipped в v1.1.0.

- New method `(*Client).ListTenants(ctx, opts *ListTenantsOptions) ([]Tenant, error)`.
  Calls `GET /v1/iam/tenants`. Used by downstream verticals (mail, kiro, grid, crm)
  для cold-start backfill перед subscription к Kafka `tenant-events` topic.
- New types: `Tenant` (clean non-pointer struct — TenantID/Code/Name/Mode/Status/PlanID/Metadata/UserCount/timestamps),
  `ListTenantsOptions` (Status/Search/Page/PageSize/SortBy/SortOrder filters mirroring proto `ListTenantsRequest`).
- 3 httptest-mock unit tests (`iam_test.go`): no-options / all-options
  query construction / empty response.

### Compatibility

Source-compatible. Pure additive — existing `ListAPIKeys`, `CreateAPIKey`,
`DeleteAPIKey`, `CheckPermission` unchanged.

---

## [v1.1.0] — 2026-04-27 — `sdk/go/v1.1.0`

### Added — `fintech` package (admin / merchant Wallet API)

Closes the SDK side of [ADR-0016 — mgCrypto BaaS surface](https://github.com/saidmashhud/mashgate/blob/main/docs/adr/0016-mgcrypto-baas-surface.md).

- New methods on `*WalletService`:
  - `CreateChain(ctx, req, idempotencyKey)` — generate non-custodial on-chain wallet (BIP-39 mnemonic returned **once** in response; surface to user, never persist). L1 of ADR-0016.
  - `Freeze(ctx, walletID, reason)` / `Unfreeze(ctx, walletID, note)` — compliance / fraud trigger.
  - `GetTransaction(ctx, walletID, transactionID)` — single-tx lookup.
- Extended methods:
  - `DepositAddress(ctx, walletID, network Network, mint Mint)` — pass non-empty `mint` (SPL token base58) to receive the Associated Token Account derived from `(wallet_owner, mint)`. Empty `mint` → native asset (wallet owner address). L3 of ADR-0016.
  - `WithdrawRequest.Mint` field — replaces the legacy `mint=...;` description hack. L2 of ADR-0016.
- New typed string aliases in `fintech/types.go`:
  - `Currency` — fiat (UZS, KZT, KGS, TJS, RUB, USD, EUR) + crypto tickers (USDT, USDC, SOL, ETH, TRX, BNB, TON).
  - `Network` — `NetworkSolana`, `NetworkEthereum`, `NetworkBase`, `NetworkPolygon`, `NetworkBSC`, `NetworkTron`, `NetworkTON`.
  - `Mint` — `MintUSDCSolanaMainnet`, `MintUSDTSolanaMainnet`.
  - `Stringer` impl for log formatters.
- 9 new httptest-based unit tests in `fintech/wallet_test.go` covering:
  - createChain shape + idempotency-key header.
  - depositAddress mint pass / omit.
  - withdraw mint in body (regression-guard against the legacy description hack).
  - freeze / unfreeze endpoints.
  - getTransaction.
  - list cursor + limit.
  - 4xx → `*APIError` mapping.
  - typed-constant wire format (no `{value: "..."}`).

### Compatibility

- Untyped string literals (`"USDC"`, `"SOLANA"`) remain assignable to typed aliases — existing call sites compile unchanged.
- Direct assignment from `var s string` requires an explicit cast: `fintech.Network(s)`.
- Wire format unchanged — JSON fields still emit plain strings.

### Internal

- `mashgate-public/sdk/go/README.md` § Wallet (admin/merchant) section added with on-chain examples.

---

## [v1.0.0] — 2026-04-25 — `sdk/go/v1.0.0`

Initial public release after extraction from `github.com/saidmashhud/mashgate` monorepo. See aggregate [CHANGELOG.md](../../CHANGELOG.md) § v1.0.0.
