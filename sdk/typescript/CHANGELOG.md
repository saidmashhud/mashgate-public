# Changelog — TypeScript SDK

`@mashgate/sdk` on npm.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) ·
Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

Aggregate changelog for all languages: [`../../CHANGELOG.md`](../../CHANGELOG.md).

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
