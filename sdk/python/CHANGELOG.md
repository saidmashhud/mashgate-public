# Changelog — Python SDK

`mashgate` on PyPI.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) ·
Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

Aggregate changelog for all languages: [`../../CHANGELOG.md`](../../CHANGELOG.md).

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
