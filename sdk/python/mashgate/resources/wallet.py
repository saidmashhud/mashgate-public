"""Wallet resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class WalletResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def add_payment_method(
        self,
        *,
        token: str,
        provider: str | None = None,
        brand: str | None = None,
        last4: str | None = None,
        exp_month: int | None = None,
        exp_year: int | None = None,
        set_default: bool | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"token": token}
        if provider is not None:
            body["provider"] = provider
        if brand is not None:
            body["brand"] = brand
        if last4 is not None:
            body["last4"] = last4
        if exp_month is not None:
            body["expMonth"] = exp_month
        if exp_year is not None:
            body["expYear"] = exp_year
        if set_default is not None:
            body["setDefault"] = set_default
        return self._c.request("POST", "/v1/wallet/payment-methods", body=body)

    def list_payment_methods(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/wallet/payment-methods")

    def remove_payment_method(self, payment_method_id: str) -> None:
        self._c.request("DELETE", f"/v1/wallet/payment-methods/{payment_method_id}")

    def set_default_payment_method(self, payment_method_id: str) -> dict[str, Any]:
        return self._c.request("PUT", f"/v1/wallet/payment-methods/{payment_method_id}/default")

    def get_balance(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/wallet/balance")

    def list_movements(self, *, page: int | None = None, page_size: int | None = None) -> dict[str, Any]:
        return self._c.request("GET", "/v1/wallet/movements", query={"page": page, "page_size": page_size})

    # ── Wallet provider payment helpers (Click / Payme / Oson) ──

    def pay_with_click(
        self,
        session_id: str,
        *,
        phone: str,
        amount: str | None = None,
    ) -> dict[str, Any]:
        """Initiate a Click wallet payment via a checkout session."""
        return self._c.checkout.complete_session(
            session_id,
            payment_method_token=f"wallet_click_{session_id}",
            payment_method_type="wallet",
            payment_method_brand="click",
            wallet_provider="click",
            wallet_phone=phone,
        )

    def pay_with_payme(
        self,
        session_id: str,
        *,
        phone: str,
        amount: str | None = None,
    ) -> dict[str, Any]:
        """Initiate a Payme wallet payment via a checkout session."""
        return self._c.checkout.complete_session(
            session_id,
            payment_method_token=f"wallet_payme_{session_id}",
            payment_method_type="wallet",
            payment_method_brand="payme",
            wallet_provider="payme",
            wallet_phone=phone,
        )

    def pay_with_oson(
        self,
        session_id: str,
        *,
        phone: str,
        amount: str | None = None,
    ) -> dict[str, Any]:
        """Initiate an Oson wallet payment via a checkout session."""
        return self._c.checkout.complete_session(
            session_id,
            payment_method_token=f"wallet_oson_{session_id}",
            payment_method_type="wallet",
            payment_method_brand="oson",
            wallet_provider="oson",
            wallet_phone=phone,
        )
