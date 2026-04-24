"""Checkout resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class CheckoutResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create_session(
        self,
        *,
        success_url: str,
        cancel_url: str,
        line_items: list[dict[str, Any]],
        currency: str,
        expires_in_minutes: int | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "successUrl": success_url,
            "cancelUrl": cancel_url,
            "lineItems": line_items,
            "currency": currency,
        }
        if expires_in_minutes is not None:
            body["expiresInMinutes"] = expires_in_minutes
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/checkout/sessions", body=body)

    def get_session(self, session_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/checkout/sessions/{session_id}")

    def complete_session(
        self,
        session_id: str,
        *,
        payment_method_token: str,
        payment_method_type: str = "card",
        payment_method_brand: str | None = None,
        payment_method_last4: str | None = None,
        wallet_provider: str | None = None,
        wallet_phone: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "sessionId": session_id,
            "paymentMethodToken": payment_method_token,
            "paymentMethodType": payment_method_type,
        }
        if payment_method_brand is not None:
            body["paymentMethodBrand"] = payment_method_brand
        if payment_method_last4 is not None:
            body["paymentMethodLast4"] = payment_method_last4
        if wallet_provider is not None:
            body["walletProvider"] = wallet_provider
        if wallet_phone is not None:
            body["walletPhone"] = wallet_phone
        return self._c.request("POST", f"/v1/checkout/sessions/{session_id}/complete", body=body)
