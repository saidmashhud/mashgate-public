"""Chain resource — crypto rails (mgChain).

Mirrors ``ChainService`` (``chain.proto``) exposed over the gateway as REST.

Backend: chain-rpc (Rust) + chain-indexer (Rust) + mgchain-orchestrator
(Scala/ZIO). Tenant isolation is enforced via the JWT — ChainService never
trusts a body-supplied ``tenant_id``.

Two surfaces are exposed here, matching the Go + TypeScript SDKs:

- High-level product surface (``/v1/chain/wallets``, ``/v1/chain/payments``,
  swaps, escrow, on/off-ramp, compliance, gas, rates, batch payouts) —
  mirrors the TypeScript ``ChainResource``.
- Low-level address/transaction/block surface (``/v1/chain/addresses``,
  ``/v1/chain/transactions``, ``/v1/chain/networks``) — mirrors the Go
  ``ChainClient``.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class ChainResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Wallets ───────────────────────────────────────────────────────

    def create_wallet(
        self,
        *,
        tenant_id: str,
        user_id: str,
        wallet_type: str,
        networks: list[str],
        label: str | None = None,
    ) -> dict[str, Any]:
        """Create a multi-network custodial crypto wallet for a user."""
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "userId": user_id,
            "walletType": wallet_type,
            "networks": networks,
        }
        if label is not None:
            body["label"] = label
        return self._c.request("POST", "/v1/chain/wallets", body=body)

    def get_wallet(self, wallet_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/chain/wallets/{wallet_id}")

    def get_wallet_balance(self, wallet_id: str) -> dict[str, Any]:
        """Return ``{"balances": [AssetBalance, ...]}`` for the wallet."""
        return self._c.request("GET", f"/v1/chain/wallets/{wallet_id}/balance")

    # ── Crypto payments ───────────────────────────────────────────────

    def pay(
        self,
        *,
        amount: str,
        asset: str,
        network: str,
        order_id: str | None = None,
        customer_id: str | None = None,
        destination_address: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Create a crypto payment intent (returns a deposit address to fund)."""
        body: dict[str, Any] = {"amount": amount, "asset": asset, "network": network}
        if order_id is not None:
            body["orderId"] = order_id
        if customer_id is not None:
            body["customerId"] = customer_id
        if destination_address is not None:
            body["destinationAddress"] = destination_address
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/chain/payments", body=body)

    def get_crypto_payment(self, payment_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/chain/payments/{payment_id}")

    # ── Swaps ─────────────────────────────────────────────────────────

    def swap(
        self,
        *,
        from_amount: str,
        from_asset: str,
        from_network: str,
        to_asset: str,
        to_network: str,
        slippage_bps: str | None = None,
        destination_address: str | None = None,
    ) -> dict[str, Any]:
        """Swap one asset for another (cross-asset / cross-network)."""
        body: dict[str, Any] = {
            "fromAmount": from_amount,
            "fromAsset": from_asset,
            "fromNetwork": from_network,
            "toAsset": to_asset,
            "toNetwork": to_network,
        }
        if slippage_bps is not None:
            body["slippageBps"] = slippage_bps
        if destination_address is not None:
            body["destinationAddress"] = destination_address
        return self._c.request("POST", "/v1/chain/swaps", body=body)

    def get_swap(self, swap_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/chain/swaps/{swap_id}")

    # ── Escrow ────────────────────────────────────────────────────────

    def create_escrow(
        self,
        *,
        amount: str,
        asset: str,
        network: str,
        payer_address: str,
        payee_address: str,
        arbiter_address: str | None = None,
        release_after: int | None = None,
        expires_at: int | None = None,
        use_case: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Open an on-chain escrow between a payer and payee."""
        body: dict[str, Any] = {
            "amount": amount,
            "asset": asset,
            "network": network,
            "payerAddress": payer_address,
            "payeeAddress": payee_address,
        }
        if arbiter_address is not None:
            body["arbiterAddress"] = arbiter_address
        if release_after is not None:
            body["releaseAfter"] = release_after
        if expires_at is not None:
            body["expiresAt"] = expires_at
        if use_case is not None:
            body["useCase"] = use_case
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/chain/escrows", body=body)

    def get_escrow(self, escrow_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/chain/escrows/{escrow_id}")

    def release_escrow(self, escrow_id: str, *, released_by: str) -> dict[str, Any]:
        """Release escrowed funds to the payee."""
        return self._c.request(
            "POST",
            f"/v1/chain/escrows/{escrow_id}/release",
            body={"releasedBy": released_by},
        )

    def dispute_escrow(self, escrow_id: str, *, reason: str) -> dict[str, Any]:
        """Raise a dispute on an escrow, suspending automatic release."""
        return self._c.request(
            "POST",
            f"/v1/chain/escrows/{escrow_id}/dispute",
            body={"reason": reason},
        )

    # ── On-ramp / off-ramp ────────────────────────────────────────────

    def on_ramp(
        self,
        *,
        fiat_amount: str,
        fiat_currency: str,
        target_asset: str,
        target_network: str,
        destination_address: str,
        provider: str | None = None,
    ) -> dict[str, Any]:
        """Buy crypto with fiat (fiat → crypto)."""
        body: dict[str, Any] = {
            "fiatAmount": fiat_amount,
            "fiatCurrency": fiat_currency,
            "targetAsset": target_asset,
            "targetNetwork": target_network,
            "destinationAddress": destination_address,
        }
        if provider is not None:
            body["provider"] = provider
        return self._c.request("POST", "/v1/chain/on-ramp", body=body)

    def off_ramp(
        self,
        *,
        crypto_amount: str,
        crypto_asset: str,
        crypto_network: str,
        target_currency: str,
        payout_method: str,
        payout_details: str,
        provider: str | None = None,
    ) -> dict[str, Any]:
        """Sell crypto for fiat (crypto → fiat)."""
        body: dict[str, Any] = {
            "cryptoAmount": crypto_amount,
            "cryptoAsset": crypto_asset,
            "cryptoNetwork": crypto_network,
            "targetCurrency": target_currency,
            "payoutMethod": payout_method,
            "payoutDetails": payout_details,
        }
        if provider is not None:
            body["provider"] = provider
        return self._c.request("POST", "/v1/chain/off-ramp", body=body)

    # ── Compliance ────────────────────────────────────────────────────

    def screen_address(self, address: str, network: str) -> dict[str, Any]:
        """Screen an address against sanction / risk lists."""
        return self._c.request(
            "POST",
            "/v1/chain/compliance/screen-address",
            body={"address": address, "network": network},
        )

    def screen_transaction(self, tx_hash: str, network: str) -> dict[str, Any]:
        """Screen a transaction against sanction / risk lists."""
        return self._c.request(
            "POST",
            "/v1/chain/compliance/screen-tx",
            body={"txHash": tx_hash, "network": network},
        )

    # ── Gas & rates ───────────────────────────────────────────────────

    def gas_estimate(self, network: str) -> dict[str, Any]:
        """Return the current gas estimate for a network."""
        return self._c.request("GET", f"/v1/chain/gas/{network}")

    def exchange_rate(self, from_asset: str, to_asset: str) -> dict[str, Any]:
        """Return the exchange rate from one asset to another."""
        return self._c.request("GET", f"/v1/chain/rates/{from_asset}/{to_asset}")

    # ── Batch payouts ─────────────────────────────────────────────────

    def batch_payout(
        self,
        *,
        asset: str,
        network: str,
        recipients: list[dict[str, Any]],
    ) -> dict[str, Any]:
        """Pay many recipients in a single batch.

        :param recipients: list of ``{"address": str, "amount": str,
            "reference"?: str}`` entries.
        """
        return self._c.request(
            "POST",
            "/v1/chain/payouts",
            body={"asset": asset, "network": network, "recipients": recipients},
        )

    # ── Address management (low-level, mirrors Go ChainClient) ────────

    def create_address(
        self,
        *,
        tenant_id: str,
        network: str,
    ) -> dict[str, Any]:
        """Derive a new deposit address for the tenant on the given network."""
        return self._c.request(
            "POST",
            "/v1/chain/addresses",
            body={"tenantId": tenant_id, "network": network},
        )

    def list_addresses(self, *, tenant_id: str, network: str | None = None) -> dict[str, Any]:
        """Return all tenant addresses, optionally filtered by network."""
        return self._c.request(
            "GET",
            "/v1/chain/addresses",
            query={"tenantId": tenant_id, "network": network},
        )

    def get_address(self, address_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/chain/addresses/{address_id}")

    def get_balance(self, address_id: str) -> dict[str, Any]:
        """Return the balance for an address on its network."""
        return self._c.request("GET", f"/v1/chain/addresses/{address_id}/balance")

    # ── Transactions (low-level, mirrors Go ChainClient) ──────────────

    def list_transactions(
        self,
        *,
        tenant_id: str,
        address_id: str | None = None,
        page: int | None = None,
        page_size: int | None = None,
    ) -> dict[str, Any]:
        """Return transaction history for a tenant (optionally per address)."""
        return self._c.request(
            "GET",
            "/v1/chain/transactions",
            query={
                "tenantId": tenant_id,
                "addressId": address_id,
                "page": page,
                "pageSize": page_size,
            },
        )

    def get_transaction(self, tx_id: str) -> dict[str, Any]:
        """Retrieve a single transaction by id or hash."""
        return self._c.request("GET", f"/v1/chain/transactions/{tx_id}")

    def estimate_fee(
        self,
        *,
        network: str,
        from_address: str | None = None,
        to_address: str | None = None,
        amount: str | None = None,
        asset: str | None = None,
    ) -> dict[str, Any]:
        """Estimate the network fee for a hypothetical send."""
        body: dict[str, Any] = {"network": network}
        if from_address is not None:
            body["fromAddress"] = from_address
        if to_address is not None:
            body["toAddress"] = to_address
        if amount is not None:
            body["amount"] = amount
        if asset is not None:
            body["asset"] = asset
        return self._c.request("POST", "/v1/chain/fee-estimate", body=body)

    def send_transaction(
        self,
        *,
        network: str,
        from_address: str,
        to_address: str,
        amount: str,
        asset: str | None = None,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        """Submit a transaction to the network.

        Idempotent via the ``Idempotency-Key`` header.
        """
        body: dict[str, Any] = {
            "network": network,
            "fromAddress": from_address,
            "toAddress": to_address,
            "amount": amount,
        }
        if asset is not None:
            body["asset"] = asset
        headers: dict[str, str] = {}
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key
        return self._c.request(
            "POST",
            "/v1/chain/transactions/send",
            body=body,
            extra_headers=headers or None,
        )

    # ── Blocks (low-level, mirrors Go ChainClient) ────────────────────

    def get_block(self, network: str, hash_or_height: str) -> dict[str, Any]:
        """Retrieve a block by hash or height on a given network."""
        return self._c.request(
            "GET",
            f"/v1/chain/networks/{network}/blocks/{hash_or_height}",
        )

    def get_latest_block(self, network: str) -> dict[str, Any]:
        """Return the most recent block on a network."""
        return self.get_block(network, "latest")

    # ── Networks (low-level, mirrors Go ChainClient) ──────────────────

    def list_networks(self) -> dict[str, Any]:
        """Return the chains supported by this Mashgate instance."""
        return self._c.request("GET", "/v1/chain/networks")
