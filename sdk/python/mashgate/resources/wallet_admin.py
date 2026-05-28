"""Admin/merchant-side WalletService client.

Mirrors ``wallet.v1.WalletService`` from
``mashgate/contracts/proto/v1/wallet.proto``, exposed over the gateway as
REST via google.api.http transcoding. End-user wallet operations
(saved cards, balance, movements) live in :class:`WalletResource`.

Auth here is tenant-scoped — pass an admin JWT or service-account API
key when constructing the client.
"""

from __future__ import annotations

from enum import Enum
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


# ── Typed string enums ──────────────────────────────────────────────────
#
# We inherit from ``str`` so members serialise transparently to JSON
# (json.dumps treats them as their string value), and so callers may pass
# either ``Currency.USDC`` or the plain string ``"USDC"`` — both are
# accepted by the server, and our resource methods cast to ``str(...)``
# at the wire boundary so internal None-stripping works either way.


class Currency(str, Enum):
    """Unit of account a wallet holds.

    Fiat values follow ISO 4217. Crypto values use the ticker symbol
    used across the Mashgate stack.
    """

    # Fiat
    UZS = "UZS"
    KZT = "KZT"
    KGS = "KGS"
    TJS = "TJS"
    RUB = "RUB"
    USD = "USD"
    EUR = "EUR"
    # Crypto / stablecoin tickers
    USDT = "USDT"
    USDC = "USDC"
    SOL = "SOL"
    ETH = "ETH"
    TRX = "TRX"
    BNB = "BNB"
    TON = "TON"


class Network(str, Enum):
    """Blockchain network for crypto wallet operations."""

    SOLANA = "SOLANA"
    ETHEREUM = "ETHEREUM"
    BASE = "BASE"
    POLYGON = "POLYGON"
    BSC = "BSC"
    TRON = "TRON"
    TON = "TON"


class Mint(str, Enum):
    """Token contract / mint addresses on the supported networks.

    Empty mint (``""``) means native asset transfer path (SOL on Solana,
    TRX on Tron, ...). Members listed here are well-known mainnet
    contracts; for tokens not in this enum pass the plain string.

    Solana mints are SPL token mint addresses (base58). TRON mints are
    TRC-20 contract addresses (base58check). ledger-core picks the right
    interpretation based on the wallet's ``network`` column.
    """

    USDC_SOLANA_MAINNET = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
    USDT_SOLANA_MAINNET = "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
    USDT_TRON_MAINNET = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
    # EVM mainnet token contracts (0x-hex).
    USDT_ETHEREUM_MAINNET = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
    USDC_ETHEREUM_MAINNET = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
    USDT_BSC_MAINNET = "0x55d398326f99059fF775485246999027B3197955"
    USDC_BSC_MAINNET = "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"
    USDT_POLYGON_MAINNET = "0xc2132D05D31c914a87C6611C10748AEb04B58e8F"
    USDC_POLYGON_MAINNET = "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359"
    USDC_BASE_MAINNET = "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"


class WalletType(str, Enum):
    FIAT = "WALLET_TYPE_FIAT"
    CRYPTO = "WALLET_TYPE_CRYPTO"
    OMNIBUS = "WALLET_TYPE_OMNIBUS"


class WalletStatus(str, Enum):
    ACTIVE = "WALLET_STATUS_ACTIVE"
    FROZEN = "WALLET_STATUS_FROZEN"
    CLOSED = "WALLET_STATUS_CLOSED"


class TransactionType(str, Enum):
    CREDIT = "TRANSACTION_TYPE_CREDIT"
    DEBIT = "TRANSACTION_TYPE_DEBIT"


class TransactionStatus(str, Enum):
    PENDING = "TRANSACTION_STATUS_PENDING"
    SETTLED = "TRANSACTION_STATUS_SETTLED"
    FAILED = "TRANSACTION_STATUS_FAILED"
    REVERSED = "TRANSACTION_STATUS_REVERSED"


class TransactionReason(str, Enum):
    DEPOSIT = "TRANSACTION_REASON_DEPOSIT"
    WITHDRAWAL = "TRANSACTION_REASON_WITHDRAWAL"
    PAYMENT = "TRANSACTION_REASON_PAYMENT"
    REFUND = "TRANSACTION_REASON_REFUND"
    SETTLEMENT = "TRANSACTION_REASON_SETTLEMENT"
    FEE = "TRANSACTION_REASON_FEE"
    ADJUSTMENT = "TRANSACTION_REASON_ADJUSTMENT"
    CONVERSION = "TRANSACTION_REASON_CONVERSION"


def _opt_str(v: Any) -> str | None:
    """Coerce enum / None / str to plain string. Returns None for None."""
    if v is None:
        return None
    return str(v.value) if isinstance(v, Enum) else str(v)


class WalletAdminResource:
    """Admin/merchant client for the full ``wallet.v1.WalletService``.

    Auth: tenant-scoped. Requires an admin JWT (preferred) or service
    account API key on the parent client.
    """

    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Off-chain wallet ──────────────────────────────────────────────

    def create(
        self,
        *,
        subject_id: str,
        subject_type: str,
        wallet_type: WalletType | str,
        currency: Currency | str,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "subject_id": subject_id,
            "subject_type": subject_type,
            "wallet_type": _opt_str(wallet_type),
            "currency": _opt_str(currency),
        }
        if idempotency_key:
            body["idempotency_key"] = idempotency_key
        return self._c.request("POST", "/v1/wallets", body=body)

    # ── On-chain wallet (L1 of ADR-0016) ──────────────────────────────

    def create_chain(
        self,
        *,
        subject_id: str,
        subject_type: str,
        currency: Currency | str,
        network: Network | str,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        """Generate a non-custodial on-chain wallet.

        Returns ``{"wallet": {...}, "mnemonic": "..."}``. The ``mnemonic``
        is shown to the end user ONCE — surface it immediately and never
        persist it server-side. Currently SOLANA only.
        """
        body: dict[str, Any] = {
            "subject_id": subject_id,
            "subject_type": subject_type,
            "currency": _opt_str(currency),
            "network": _opt_str(network),
        }
        if idempotency_key:
            body["idempotency_key"] = idempotency_key
        return self._c.request("POST", "/v1/wallets/chain", body=body)

    def import_chain(
        self,
        *,
        subject_id: str,
        subject_type: str,
        currency: Currency | str,
        network: Network | str,
        mnemonic: str,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        """Recover / import a non-custodial wallet from a BIP-39 mnemonic.

        Returns ``{"wallet": {...}, "was_existing": bool, "recovered_at": str}``.

        - Re-importing the same phrase for the same ``subject_id`` returns
          the existing wallet с ``was_existing=True`` (idempotent recovery).
        - Re-importing under a different ``subject_id`` within the tenant
          raises ``403 PERMISSION_DENIED`` — credential-stuffing protection.

        :param mnemonic: BIP-39 phrase (12 or 24 words). **Clear from caller
            memory immediately after this call** — phrase touches process
            memory briefly but is never persisted plaintext server-side
            (SHA-256 hash для dedup only).

        :raises mashgate.errors.APIError:
            - ``400 INVALID_ARGUMENT`` — mnemonic fails BIP-39 checksum,
              network/currency missing.
            - ``403 PERMISSION_DENIED`` — same mnemonic_hash already
              belongs to a different subject in this tenant.
            - ``412 FAILED_PRECONDITION`` — ``WALLET_ENCRYPTION_KEY`` not
              configured server-side.
        """
        body: dict[str, Any] = {
            "subject_id": subject_id,
            "subject_type": subject_type,
            "currency": _opt_str(currency),
            "network": _opt_str(network),
            "mnemonic": mnemonic,
        }
        if idempotency_key:
            body["idempotency_key"] = idempotency_key
        return self._c.request("POST", "/v1/wallets/chain/import", body=body)

    # ── Read ──────────────────────────────────────────────────────────

    def get(self, wallet_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/wallets/{wallet_id}")

    def list(
        self,
        *,
        subject_id: str | None = None,
        status: WalletStatus | str | None = None,
        wallet_type: WalletType | str | None = None,
        limit: int | None = None,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/wallets",
            query={
                "subject_id": subject_id,
                "status": _opt_str(status),
                "wallet_type": _opt_str(wallet_type),
                "limit": limit,
                "cursor": cursor,
            },
        )

    # ── Compliance ────────────────────────────────────────────────────

    def freeze(self, wallet_id: str, *, reason: str) -> dict[str, Any]:
        """Halt all debits on a wallet (compliance / fraud trigger)."""
        return self._c.request(
            "POST",
            f"/v1/wallets/{wallet_id}/freeze",
            body={"freeze_reason": reason},
        )

    def unfreeze(self, wallet_id: str, *, note: str = "") -> dict[str, Any]:
        return self._c.request(
            "POST",
            f"/v1/wallets/{wallet_id}/unfreeze",
            body={"note": note},
        )

    # ── Money movements ───────────────────────────────────────────────

    def credit(
        self,
        wallet_id: str,
        *,
        amount: str,
        reason: TransactionReason | str,
        reference_id: str | None = None,
        reference_type: str | None = None,
        description: str | None = None,
        idempotency_key: str | None = None,
        merchant_id: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"amount": amount, "reason": _opt_str(reason)}
        for k, v in [
            ("reference_id", reference_id),
            ("reference_type", reference_type),
            ("description", description),
            ("idempotency_key", idempotency_key),
            ("merchant_id", merchant_id),
        ]:
            if v is not None:
                body[k] = v
        return self._c.request("POST", f"/v1/wallets/{wallet_id}/credit", body=body)

    def debit(
        self,
        wallet_id: str,
        *,
        amount: str,
        reason: TransactionReason | str,
        reference_id: str | None = None,
        reference_type: str | None = None,
        description: str | None = None,
        idempotency_key: str | None = None,
        merchant_id: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"amount": amount, "reason": _opt_str(reason)}
        for k, v in [
            ("reference_id", reference_id),
            ("reference_type", reference_type),
            ("description", description),
            ("idempotency_key", idempotency_key),
            ("merchant_id", merchant_id),
        ]:
            if v is not None:
                body[k] = v
        return self._c.request("POST", f"/v1/wallets/{wallet_id}/debit", body=body)

    def transfer(
        self,
        from_wallet_id: str,
        *,
        to_wallet_id: str,
        amount: str,
        reason: TransactionReason | str | None = None,
        description: str | None = None,
        idempotency_key: str | None = None,
        merchant_id: str | None = None,
        note: str | None = None,
    ) -> dict[str, Any]:
        """Atomically transfer ``amount`` between two wallets in the same tenant.

        Mirrors ``wallet.v1.WalletService.TransferBetweenWallets``.
        Same-currency only in v1; cross-currency FX is out of scope.

        The server commits a single Postgres transaction: balance delta on
        both wallets + two ``wallet_transactions`` rows (debit on source,
        credit on destination, both referencing the synthetic
        ``transfer_id``) + three outbox events (``wallet.debit``,
        ``wallet.credit``, ``wallet.transfer``). Either the entire transfer
        commits or none of it.

        :param from_wallet_id: Source wallet UUID.
        :param to_wallet_id: Destination wallet UUID. Must be distinct.
        :param amount: Decimal string. Must be > 0.
        :param reason: Defaults to ``TRANSACTION_REASON_ADJUSTMENT`` if omitted.
        :param idempotency_key: Optional. Replay-safe; the server namespaces
            the key per leg internally so the global ``wallet_transactions``
            UNIQUE index covers both rows.
        :param merchant_id: Surfaced in the paired ``wallet.{credit,debit}``
            envelope events. Empty = those events emit without ``merchant_id``
            and consumers requiring it drop them (money state still commits).
        :param note: Optional free-text attached to the ``wallet.transfer``
            envelope event.

        :returns: ``{"transfer_id": str, "debit": WalletTransaction,
            "credit": WalletTransaction}``.

        :raises mashgate.errors.APIError: server-side rejection — distinct
            HTTP statuses surface the gRPC status:

            - ``400 INVALID_ARGUMENT`` — same wallet IDs, currency mismatch,
              non-positive amount.
            - ``403 PERMISSION_DENIED`` — wallets belong to different tenants.
            - ``404 NOT_FOUND`` — wallet does not exist.
            - ``412 FAILED_PRECONDITION`` — source/destination frozen,
              insufficient balance.
        """
        body: dict[str, Any] = {
            "to_wallet_id": to_wallet_id,
            "amount": amount,
        }
        for k, v in [
            ("reason", _opt_str(reason)),
            ("description", description),
            ("idempotency_key", idempotency_key),
            ("merchant_id", merchant_id),
            ("note", note),
        ]:
            if v is not None and v != "":
                body[k] = v
        return self._c.request(
            "POST", f"/v1/wallets/{from_wallet_id}/transfer", body=body
        )

    def withdraw(
        self,
        wallet_id: str,
        *,
        amount: str,
        destination_type: str,
        destination_id: str,
        network: Network | str | None = None,
        mint: Mint | str | None = None,
        description: str | None = None,
        idempotency_key: str | None = None,
        sponsor_wallet_id: str | None = None,
    ) -> dict[str, Any]:
        """Initiate a withdrawal.

        Pass ``mint`` for SPL tokens, leave None for native SOL. L2 of
        ADR-0016 — the ``mint`` field replaces the legacy ``mint=...;``
        prefix in ``description``.

        :param sponsor_wallet_id: Optional sponsor wallet UUID. When set,
            the platform sponsor pays the chain fee + any SPL ATA rent
            instead of the source — used for gasless withdrawals so a
            customer holding USDT but zero SOL can still move tokens.
            Sponsor must be in the same tenant + network, active, on-chain.
        """
        body: dict[str, Any] = {
            "amount": amount,
            "destination_type": destination_type,
            "destination_id": destination_id,
        }
        for k, v in [
            ("network", _opt_str(network)),
            ("mint", _opt_str(mint)),
            ("description", description),
            ("idempotency_key", idempotency_key),
            ("sponsor_wallet_id", sponsor_wallet_id),
        ]:
            if v is not None and v != "":
                body[k] = v
        return self._c.request("POST", f"/v1/wallets/{wallet_id}/withdraw", body=body)

    # ── Transactions ──────────────────────────────────────────────────

    def list_transactions(
        self,
        wallet_id: str,
        *,
        type: TransactionType | str | None = None,
        status: TransactionStatus | str | None = None,
        reason: TransactionReason | str | None = None,
        limit: int | None = None,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            f"/v1/wallets/{wallet_id}/transactions",
            query={
                "type": _opt_str(type),
                "status": _opt_str(status),
                "reason": _opt_str(reason),
                "limit": limit,
                "cursor": cursor,
            },
        )

    def get_transaction(self, wallet_id: str, transaction_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET",
            f"/v1/wallets/{wallet_id}/transactions/{transaction_id}",
        )

    # ── Deposit address (L3 of ADR-0016) ─────────────────────────────

    def deposit_address(
        self,
        wallet_id: str,
        *,
        network: Network | str,
        mint: Mint | str | None = None,
    ) -> dict[str, Any]:
        """Resolve the on-chain deposit target for a wallet.

        Pass a non-empty ``mint`` (SPL token mint) on Solana to get the
        Associated Token Account derived from ``(wallet_owner, mint)``.
        Empty / None ``mint`` returns the wallet owner address (native
        asset path — SOL etc.).
        """
        mint_str = _opt_str(mint)
        return self._c.request(
            "GET",
            f"/v1/wallets/{wallet_id}/deposit-address",
            query={
                "network": _opt_str(network),
                "mint": mint_str if mint_str else None,
            },
        )
