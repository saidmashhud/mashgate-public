# Changelog — TypeScript SDK

`@mashgate/sdk` on npm.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) ·
Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

Aggregate changelog for all languages: [`../../CHANGELOG.md`](../../CHANGELOG.md).

---

## [1.4.0] — 2026-05-19

### Added — `WalletAdminResource.importChain` (BIP-39 recovery с was_existing)

- `client.walletAdmin.importChain(req)` →
  `POST /v1/wallets/chain/import`. Returns
  `{ wallet, was_existing, recovered_at? }`.
- New interfaces: `ImportChainWalletRequest`, `ImportChainWalletResponse`.
- 2 new vitest mock-fetch tests (fresh import + recovery).

Server-side guarantees: mnemonic_hash UNIQUE per tenant. Re-import same
phrase для одного subject = idempotent recovery с
`was_existing=true`. Re-import под другим subject_id → `PERMISSION_DENIED`
(credential-stuffing protection). Mnemonic SHA-256-hashed before storage;
plaintext never persisted.

Errors mapped from gRPC status:
- `INVALID_ARGUMENT` — mnemonic fails BIP-39 checksum, network/currency missing.
- `PERMISSION_DENIED` — cross-subject mnemonic reuse.
- `FAILED_PRECONDITION` — `WALLET_ENCRYPTION_KEY` not configured server-side.

---

## [1.3.0] — 2026-05-15

### Added — `WalletAdminResource.transfer` (atomic inter-wallet movement)

Mirrors new `wallet.v1.WalletService.TransferBetweenWallets` RPC
(ledger-core handler landed in `mashgate@af67d653`, 2026-05-15).
Same-currency v1.

- `client.walletAdmin.transfer(fromWalletId, req)` →
  `POST /v1/wallets/{fromWalletId}/transfer`. Returns
  `{ transfer_id, debit, credit }` with both `WalletTransaction` rows.
- New types: `TransferBetweenWalletsRequest`, `TransferBetweenWalletsResponse`.
- 1 new vitest mock-fetch test.

Server-side guarantees (per-tenant atomic): balance delta on both wallets
+ two `wallet_transactions` rows + three outbox events
(`wallet.debit`, `wallet.credit`, `wallet.transfer`) commit in one
Postgres tx. Idempotency key namespaced per leg server-side.

Errors mapped from gRPC status:
- `INVALID_ARGUMENT` — same wallet IDs, currency mismatch, non-positive amount.
- `FAILED_PRECONDITION` — source/destination frozen, insufficient balance.
- `PERMISSION_DENIED` — wallets belong to different tenants.
- `NOT_FOUND` — wallet does not exist.

---

## [1.1.0] — 2026-04-27

### Added — `WalletAdminResource` (admin / merchant Wallet API)

Closes the SDK side of [ADR-0016 — mgCrypto BaaS surface](https://github.com/saidmashhud/mashgate/blob/main/docs/adr/0016-mgcrypto-baas-surface.md).

- **`client.walletAdmin`** — full `wallet.v1.WalletService`. Use with admin JWT or service-account API key. End-user wallet (saved cards, balance) is still on `client.wallet`.
- Methods: `create`, `createChain` (returns `{wallet, mnemonic}` — mnemonic shown once), `get`, `list` (with `cursor`/`limit`), `freeze(reason)`, `unfreeze(note)`, `credit`, `debit`, `withdraw` (now accepts `mint`), `listTransactions`, `getTransaction`, `depositAddress(walletId, network, mint?)`.
- Typed enums (const-as-object + literal union types):
  - `Currency` — UZS / KZT / KGS / TJS / RUB / USD / EUR / USDT / USDC / SOL / ETH / TRX / BNB / TON.
  - `Network` — Solana / Ethereum / Base / Polygon / BSC / Tron / TON.
  - `Mint` — `USDCSolanaMainnet`, `USDTSolanaMainnet`. Allows arbitrary base58 strings outside the whitelist (server validates).
  - `WalletType`, `WalletStatus`, `TransactionType`, `TransactionStatus`, `TransactionReason` — proto enum-strings.
- Domain types: `Wallet` (alias `AdminWallet` to avoid name clash with end-user wallet types), `WalletTransaction`, `DepositAddress`, plus request/response shapes.
- 8 vitest tests in `__tests__/walletAdmin.test.ts` exercising the full surface via mocked `fetch`.

### Fixed

- `__tests__/client.test.ts` mock `fetch` now returns `headers: new Headers()`. Previously failed against `response.headers.get("x-request-id")` in `client.ts:181`.

### Compatibility

- TypeScript strict mode: typed const-as-object enums accept untyped string literals — existing callers are not broken.
- Wire format: enums serialise as plain JSON strings (`"USDC"`, not `{value:"USDC"}`).

### Internal

- New `README.md` with on-chain (mgCrypto) examples.

---

## [1.0.0] — 2026-04-25

Initial public release after extraction from `github.com/saidmashhud/mashgate` monorepo. See aggregate [CHANGELOG.md](../../CHANGELOG.md) § v1.0.0.
