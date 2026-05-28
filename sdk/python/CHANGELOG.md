# Changelog — Python SDK

`mashgate` on PyPI.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) ·
Versioning: [SemVer](https://semver.org/spec/v2.0.0.html).

Aggregate changelog for all languages: [`../../CHANGELOG.md`](../../CHANGELOG.md).

---

## [0.7.0] — 2026-05-19

### Added — EVM chain provider Phase 1

7 new ERC-20 mint constants под `Mint.*` для Ethereum / BSC / Polygon /
Base — USDT и USDC contract addresses on mainnet. `Network.ETHEREUM`
etc. already existed; server now accepts them.

Server-side: wallet-crypto adds `evm` (EIP-55 address) + `evm_tx`
(RLP + EIP-1559 build) modules. chain-rpc gains `BuildEvmTransferTx`
RPC. ledger-core handlers derive via BIP-32 secp256k1 на coin 60,
sign keccak256(signing_payload), pack signed RLP, broadcast.

Phase 1 limits: sponsor wallets (gasless) — Solana only; ERC-20 decimals
assumed = 6 (USDT/USDC); status sync worker — Solana only (Phase 2).

2 new pytest+respx tests (ETH USDT withdraw + EVM mint literals
pinned).

---

## [0.6.0] — 2026-05-19

### Added — TRON chain provider Phase 1

- New `Mint.USDT_TRON_MAINNET = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"` —
  TRC-20 USDT contract address. The `Mint` enum now carries both Solana
  SPL mints and TRON TRC-20 contracts; ledger-core selects interpretation
  by wallet network.
- `Network.TRON` already existed; server now actually accepts it.
- 1 new pytest+respx test for TRON USDT withdraw.

Server-side: secp256k1 derivation + TRON address builder в wallet-crypto;
chain-rpc gains `BuildTronTransferTx`; ledger-core wallets с
`network=TRON` derive via BIP-32 на coin type 195, sign SHA-256(raw_data)
locally and broadcast through chain-rpc.

Phase 1 limits: sponsor wallets (gasless) — Solana only; TRON status-sync
worker — Phase 2.

---

## [0.5.0] — 2026-05-19

### Added — `withdraw(..., sponsor_wallet_id=...)` (gasless withdrawals)

- New optional `sponsor_wallet_id` keyword arg on
  `WalletAdminResource.withdraw`. When set, the platform sponsor wallet
  pays the chain fee + ATA rent instead of the source.
- 1 new pytest+respx test (15 total).

Server-side: dual-signer Solana tx (sponsor first, source second). Phase
1 only — fee debit from sponsor's off-chain balance deferred to
status-sync worker (Phase 2).

---

## [0.4.0] — 2026-05-19

### Added — `WalletAdminResource.import_chain` (BIP-39 recovery)

- `client.wallet_admin.import_chain(*, subject_id, subject_type, currency,
  network, mnemonic, idempotency_key=None)` →
  `POST /v1/wallets/chain/import`. Returns
  `{"wallet": ..., "was_existing": bool, "recovered_at": str}`.
- 2 new pytest+respx tests (fresh import + recovery).

Server-side guarantees: mnemonic_hash UNIQUE per tenant. Re-import same
phrase для одного subject = idempotent recovery с
`was_existing=True`. Re-import под другим subject_id → `403
PERMISSION_DENIED` (credential-stuffing protection). Mnemonic
SHA-256-hashed before storage; plaintext never persisted.

Errors mapped from gRPC status:
- `400 INVALID_ARGUMENT` — mnemonic fails BIP-39 checksum, network/currency missing.
- `403 PERMISSION_DENIED` — cross-subject mnemonic reuse.
- `412 FAILED_PRECONDITION` — `WALLET_ENCRYPTION_KEY` not configured server-side.

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
