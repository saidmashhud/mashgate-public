"""Payments resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class PaymentsResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create(
        self,
        *,
        amount: str,
        currency: str,
        order_id: str | None = None,
        auto_capture: bool | None = None,
        idempotency_key: str | None = None,
        payment_method_token: str | None = None,
        payment_method_type: str | None = None,
        payment_method_brand: str | None = None,
        payment_method_last4: str | None = None,
        payment_method_bin: str | None = None,
        payment_method_provider: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"amount": amount, "currency": currency}
        if order_id is not None:
            body["orderId"] = order_id
        if auto_capture is not None:
            body["autoCapture"] = auto_capture
        if payment_method_token is not None:
            body["paymentMethodToken"] = payment_method_token
        if payment_method_type is not None:
            body["paymentMethodType"] = payment_method_type
        if payment_method_brand is not None:
            body["paymentMethodBrand"] = payment_method_brand
        if payment_method_last4 is not None:
            body["paymentMethodLast4"] = payment_method_last4
        if payment_method_bin is not None:
            body["paymentMethodBin"] = payment_method_bin
        if payment_method_provider is not None:
            body["paymentMethodProvider"] = payment_method_provider
        if metadata is not None:
            body["metadata"] = metadata

        headers: dict[str, str] = {}
        if idempotency_key:
            headers["X-Idempotency-Key"] = idempotency_key

        return self._c.request("POST", "/v1/payments", body=body, extra_headers=headers or None)

    def get(self, payment_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/payments/{payment_id}")

    def list(self, *, page: int | None = None, page_size: int | None = None) -> dict[str, Any]:
        return self._c.request("GET", "/v1/payments", query={"page": page, "page_size": page_size})

    def authorize(self, payment_id: str, *, idempotency_key: str | None = None) -> dict[str, Any]:
        headers: dict[str, str] = {}
        if idempotency_key:
            headers["X-Idempotency-Key"] = idempotency_key
        return self._c.request("POST", f"/v1/payments/{payment_id}/authorize", extra_headers=headers or None)

    def capture(
        self,
        payment_id: str,
        *,
        amount: str | None = None,
        currency: str | None = None,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {}
        if amount is not None:
            body["amount"] = amount
        if currency is not None:
            body["currency"] = currency
        headers: dict[str, str] = {}
        if idempotency_key:
            headers["X-Idempotency-Key"] = idempotency_key
        return self._c.request(
            "POST", f"/v1/payments/{payment_id}/capture", body=body or None, extra_headers=headers or None
        )

    def void(self, payment_id: str, *, idempotency_key: str | None = None) -> dict[str, Any]:
        headers: dict[str, str] = {}
        if idempotency_key:
            headers["X-Idempotency-Key"] = idempotency_key
        return self._c.request("POST", f"/v1/payments/{payment_id}/void", extra_headers=headers or None)

    def refund(
        self,
        payment_id: str,
        *,
        amount: str,
        currency: str | None = None,
        reason: str | None = None,
        note: str | None = None,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"amount": amount}
        if currency is not None:
            body["currency"] = currency
        if reason is not None:
            body["reason"] = reason
        if note is not None:
            body["note"] = note
        headers: dict[str, str] = {}
        if idempotency_key:
            headers["X-Idempotency-Key"] = idempotency_key
        return self._c.request(
            "POST", f"/v1/payments/{payment_id}/refund", body=body, extra_headers=headers or None
        )
