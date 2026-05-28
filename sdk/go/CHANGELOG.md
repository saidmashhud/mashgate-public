# Changelog — Go SDK

`github.com/saidmashhud/mashgate-public/sdk/go`

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) ·
Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

Aggregate changelog for all languages: [`../../CHANGELOG.md`](../../CHANGELOG.md).

---

## [v1.13.0] — 2026-05-19 — `sdk/go/v1.13.0`

### Added — EVM chain provider Phase 1 (Ethereum / BSC / Polygon / Base)

7 new ERC-20 mint constants:
- `MintUSDTEthereumMainnet` / `MintUSDCEthereumMainnet`
- `MintUSDTBscMainnet` / `MintUSDCBscMainnet`
- `MintUSDTPolygonMainnet` / `MintUSDCPolygonMainnet`
- `MintUSDCBaseMainnet`

`NetworkEthereum` / `NetworkBSC` / `NetworkPolygon` / `NetworkBase`
already existed; server-side ledger-core теперь actually accepts them.

Server-side (`mashgate@TBD`):
- `wallet-crypto` got `evm` module (keccak256 → EIP-55 checksum address)
  and `evm_tx` module (minimal RLP encoder + EIP-1559 build + ERC-20
  `transfer(address,uint256)` calldata).
- chain-rpc gains `BuildEvmTransferTx` RPC — fetches nonce + EIP-1559
  fees from the matching EthereumProvider (chain_id pinned at construction:
  Ethereum=1, BSC=56, Polygon=137, Base=8453), RLP-encodes the body,
  returns signing payload + context.
- ledger-core wallets с network=ETHEREUM/BSC/POLYGON/BASE derive via
  BIP-32 secp256k1 on coin type 60 (industry standard для all EVM
  chains), sign keccak256(signing_payload) с k256 ECDSA recoverable,
  pack the signed RLP locally (shared `wallet-crypto/evm_tx`), broadcast
  through chain-rpc's EthereumProvider.

Phase 1 limits:
- Sponsor wallets (gasless) — Solana only.
- ERC-20 decimals assumed = 6 (USDT/USDC). 18-decimal tokens (DAI, WETH)
  need a per-token decimals override — Phase 2.
- Status sync worker — Solana only. EVM confirmation detection via
  `eth_getTransactionReceipt` — Phase 2.

2 new httptest-mock tests (ETH USDT withdraw shape + all 7 EVM mint
literals pinned).

---

## [v1.12.0] — 2026-05-19 — `sdk/go/v1.12.0`

### Added — TRON chain provider Phase 1 (USDT-TRC20 + native TRX)

- New `MintUSDTTronMainnet = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"` — TRC-20
  USDT contract address. Reuses the existing `Mint` type alias because
  TRON withdrawals share the SDK shape with Solana SPL withdrawals.
- `NetworkTron` already shipped в earlier release — server-side
  ledger-core now actually accepts it instead of bailing with "only
  SOLANA supported".
- 2 new httptest-mock tests covering TRC-20 withdraw + wire-value
  assertions.

Server-side (`mashgate@TBD`):
- `wallet-crypto` got `slip10_secp256k1` (BIP-32 derive, returns 65-byte
  uncompressed pubkey) и `tron` (keccak256 → `0x41` prefix → base58check
  address derivation). Phantom-style derivation path `m/44'/195'/0'/0/0`.
- chain-rpc new `BuildTronTransferTx` RPC delegating to the TRON node's
  `/wallet/createtransaction` (native) or `/wallet/triggersmartcontract`
  (TRC-20). Returns raw_data_hex + tx JSON для local signing.
- ledger-core `handle_create_chain_wallet` / `handle_import_chain_wallet`
  now branch on network; TRON wallets derive via secp256k1, store the
  32-byte private key encrypted as before. `handle_on_chain_transfer`
  branches: secp256k1 ECDSA sign SHA-256(raw_data), inject signature
  into TRON tx JSON, broadcast.

Phase 1 limits:
- Sponsor wallets (gasless) supported для SOLANA only; passing
  `sponsor_wallet_id` on a TRON wallet returns an explicit error.
- Status sync worker still polls Solana-only — TRON confirmation
  detection deferred to Phase 2.

---

## [v1.11.0] — 2026-05-19 — `sdk/go/v1.11.0`

### Added — `WithdrawRequest.SponsorWalletID` (gasless withdrawals)

`WalletService.Withdraw` теперь принимает optional `SponsorWalletID`. When
set, the platform sponsor wallet pays the chain fee + any SPL ATA rent
instead of the source — letting a customer holding USDT but zero SOL still
move tokens off-chain.

- New field `WithdrawRequest.SponsorWalletID string`.
- 1 new httptest-mock test verifying field is forwarded в body.

Server-side guarantees (per-tenant):
- Sponsor must be active on-chain wallet in same tenant + network.
- Dual-signer Solana tx: sponsor signs first (key[0], pays fee), source
  signs second (key[1], authorises movement).
- Fee debit deferred to status-sync worker on confirmation (Phase 2 —
  current implementation broadcasts с sponsor as fee payer but doesn't yet
  debit sponsor's off-chain balance separately from source's).

---

## [v1.10.0] — 2026-05-19 — `sdk/go/v1.10.0`

### 💥 BREAKING — `ImportChain` return type now `*ImportChainWalletResponse`

Recovery flow polished. The RPC now surfaces a `was_existing` flag и
distinguishes a fresh import from a recovery (same mnemonic re-imported
для того же subject). Cross-subject mnemonic reuse within a tenant now
returns `403 PERMISSION_DENIED` (credential-stuffing protection).

**Migration**: callers that destructured `*Wallet` directly need
`resp.Wallet`. Example:

```go
// Before (v1.8.0/v1.9.0):
wallet, err := client.Wallet.ImportChain(ctx, req, idem)
fmt.Println(wallet.WalletID)

// After (v1.10.0):
resp, err := client.Wallet.ImportChain(ctx, req, idem)
if resp.WasExisting {
    log.Println("✓ wallet recovered")
} else {
    log.Println("✓ wallet imported")
}
fmt.Println(resp.Wallet.WalletID)
```

- New `ImportChainWalletResponse{ Wallet, WasExisting, RecoveredAt }`.
- 3 new httptest-mock tests (`FreshImport`, `Recovery`, `CrossSubject403`).

Server-side guarantees (per-tenant): mnemonic_hash UNIQUE lookup runs
before key derivation, so a denied import никогда не decrypts the seed.
Audit log records every fresh import (`wallet.imported`), recovery
(`wallet.imported_reused`), and denial (`wallet.import_denied`).

---

## [v1.9.0] — 2026-05-15 — `sdk/go/v1.9.0`

### Added — `WalletService.Transfer` (atomic inter-wallet movement)

Mirrors new `wallet.v1.WalletService.TransferBetweenWallets` RPC (proto +
ledger-core handler landed in `mashgate@af67d653`, 2026-05-15). Same-currency
v1: cross-currency FX out of scope.

- **Go (`sdk/go/fintech`)**:
  - `(*WalletService).Transfer(ctx, TransferRequest, idempotencyKey) (*TransferResponse, error)`
    → `POST /v1/wallets/{from_wallet_id}/transfer`.
  - New types `TransferRequest` + `TransferResponse` (mirror proto messages).
  - 3 new httptest-mock tests (`SendsExpectedShape`,
    `PropagatesServerError`, `MirrorsIdempotencyKeyIntoBody`).

Server-side guarantees (per-tenant atomic): both balances + two
`wallet_transactions` rows + three outbox events (`wallet.debit`,
`wallet.credit`, `wallet.transfer`) commit in one Postgres tx, or none
at all. Idempotency key is namespaced per leg server-side
(`<key>:debit` / `<key>:credit`) so the global UNIQUE on
`idempotency_key` covers both rows; replay returns the cached pair.

Errors mapped from gRPC status:
- `INVALID_ARGUMENT` — same wallet IDs, currency mismatch, non-positive amount.
- `FAILED_PRECONDITION` — source/destination frozen, insufficient balance.
- `PERMISSION_DENIED` — wallets belong to different tenants.
- `NOT_FOUND` — wallet does not exist.

---

## [v1.7.0] — 2026-05-12 — `sdk/go/v1.7.0`

### Added — eight resources closing TS-SDK gap

All resources mirror existing TS counterparts and use the same `ResourceClient` sub-client pattern as Billing.

- **`AnalyticsClient`** (10 RPCs) — payment metrics, volume time-series, transaction counts, payment-method/geo breakdowns, failure analysis, customer cohorts, segments, top customers.
- **`ChainClient`** (12 RPCs) — addresses CRUD, balance, transactions list/get, fee estimate, send, blocks, networks (subset of 21 from chain.proto).
- **`DeveloperClient`** (6 RPCs) — self-service portal: API keys CRUD, webhook endpoints listing, activity + integration health.
- **`LocalPaymentsClient`** (6 RPCs) — country-specific providers (TJ Tcell/Korti Milli/Alif, UZ Click/Payme): initiate, confirm, get status, cancel, list, supported methods.
- **`MeteringClient`** (3 RPCs) — usage event recording, list, summary. Idempotent.
- **`RiskClient`** (8 RPCs) — transaction assessment, assessments list/get, blocklist CRUD, rules read, identifier risk profile.
- **`SettingsClient`** (5 RPCs) — tenant config blob: get/list/set/delete + bulk patch.
- **`WalletAdminClient`** (7 RPCs) — privileged wallet ops: list, get, freeze, unfreeze, adjust-balance, close, audit-log. RBAC-gated.

### Removed

- `_todo_resources.go` — все resources implemented; documentation no longer needed.

### Migration notes

No breaking changes. Existing `client.Billing.*`, `client.Storage.*`, etc. continue working. New resources accessible via `client.Analytics.*`, `client.Chain.*`, etc.

### Tested

`go vet ./...` ✓  `go build ./...` ✓  `go test ./...` ✓ (both `sdk/go` and `sdk/go/fintech` packages).

---

## [v1.5.0] — 2026-05-12 — `sdk/go/v1.5.0`

### Added — `Billing` resource: 15 RPCs on `BillingClient` sub-client

Closes the largest contract-drift gap with TypeScript SDK. Wraps
`BillingService` from `contracts/proto/v1/billing.proto` (29 RPCs).

Methods on `client.Billing`:
- Plans: `ListPlans`, `GetPlan`
- Subscription: `GetSubscription`, `ChangePlan`, `CancelPlan`, `PreviewPlanChange`
- Payment methods: `ListPaymentMethods`, `AddPaymentMethod`, `SetDefaultPaymentMethod`, `RemovePaymentMethod`
- Invoices: `ListInvoices`, `GetInvoice`, `PayInvoice`
- Credit/promo: `GetCreditBalance`, `RedeemPromoCode`

New types in `models.go`: `BillingPlan`, `BillingSubscription`, `ChangePlanRequest`,
`CancelPlanRequest`, `PreviewPlanChangeRequest/Response`, `BillingPaymentMethod`,
`AddBillingPaymentMethodRequest`, `BillingInvoice`, `BillingInvoiceLine`,
`CreditBalance`, `RedeemPromoCodeResponse`.

### Documented — `_todo_resources.go`

Explicit TODO file marking 8 resources still missing vs TypeScript SDK:
`analytics`, `chain`, `developer`, `local_payments`, `metering`,
`risk`, `settings`, `wallet_admin`. Generated types already exist in
`_generated/types.gen.go` — only idiomatic Go wrappers pending.

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
