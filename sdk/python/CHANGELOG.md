# Changelog — Python SDK

`mashgate` on PyPI.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) ·
Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

Aggregate changelog for all languages: [`../../CHANGELOG.md`](../../CHANGELOG.md).

---

## [0.3.0] — 2026-05-15

### Added — `WalletAdminResource.transfer` (atomic inter-wallet movement)

Mirrors new `wallet.v1.WalletService.TransferBetweenWallets` RPC
(ledger-core handler landed в `mashgate@af67d653`, 2026-05-15).
Same-currency v1.

- `client.wallet_admin.transfer(from_wallet_id, *, to_wallet_id, amount, ...)`
  → `POST /v1/wallets/{from_wallet_id}/transfer`. Returns
  `dict[str, Any]` with `transfer_id`, `debit`, `credit` keys.
- 2 new pytest+respx tests (`test_transfer_posts_to_from_wallet_path_with_body`,
  `test_transfer_strips_optional_fields_when_absent`).

Server-side guarantees (per-tenant atomic): balance delta on both wallets
+ two `wallet_transactions` rows + three outbox events
(`wallet.debit`, `wallet.credit`, `wallet.transfer`) commit in one
Postgres tx. Idempotency key namespaced per leg server-side.

Errors mapped from gRPC status:
- `400 INVALID_ARGUMENT` — same wallet IDs, currency mismatch, non-positive amount.
- `403 PERMISSION_DENIED` — wallets belong to different tenants.
- `404 NOT_FOUND` — wallet does not exist.
- `412 FAILED_PRECONDITION` — source/destination frozen, insufficient balance.

---

## [0.2.0] — 2026-04-27

### Added — `WalletAdminResource` (admin / merchant Wallet API)

Closes the SDK side of [ADR-0016 — mgCrypto BaaS surface](https://github.com/saidmashhud/mashgate/blob/main/docs/adr/0016-mgcrypto-baas-surface.md).

- **`client.wallet_admin`** — full `wallet.v1.WalletService`. Use with admin JWT or service-account API key. End-user wallet operations (saved cards, balance) are still on `client.wallet`.
- Methods (keyword-only args, type hints, docstrings):
  - `create`, `create_chain` (returns `{"wallet": {...}, "mnemonic": "..."}` — mnemonic shown once), `get`, `list`.
  - `freeze(wallet_id, *, reason)` / `unfreeze(wallet_id, *, note="")`.
  - `credit`, `debit`, `withdraw` (now accepts `mint`).
  - `list_transactions`, `get_transaction`.
  - `deposit_address(wallet_id, *, network, mint=None)`.
- Typed enums (`(str, Enum)` — JSON-serialise transparently to their string values):
  - `Currency` — UZS / KZT / KGS / TJS / RUB / USD / EUR / USDT / USDC / SOL / ETH / TRX / BNB / TON.
  - `Network` — SOLANA / ETHEREUM / BASE / POLYGON / BSC / TRON / TON.
  - `Mint` — `USDC_SOLANA_MAINNET`, `USDT_SOLANA_MAINNET` (arbitrary base58 strings also accepted).
  - `WalletType`, `WalletStatus`, `TransactionType`, `TransactionStatus`, `TransactionReason` — proto enum-strings.
- All public types exported at package root: `from mashgate import Currency, Network, Mint, WalletAdminResource, …`.
- **First test suite** for the SDK — 10 pytest+respx tests in `tests/test_wallet_admin.py`.

### Compatibility

- `(str, Enum)` inheritance keeps callers using plain strings (`"USDC"`) source-compatible — `Currency.USDC == "USDC"` is `True`.
- Requires Python ≥ 3.9 (per `pyproject.toml`); `(str, Enum)` is the pre-3.11 idiom for `StrEnum` semantics.

### Internal

- New `README.md` with on-chain (mgCrypto) examples.

---

## [0.1.0] — 2026-04-25

Initial public release after extraction from `github.com/saidmashhud/mashgate` monorepo. See aggregate [CHANGELOG.md](../../CHANGELOG.md) § v1.0.0.
