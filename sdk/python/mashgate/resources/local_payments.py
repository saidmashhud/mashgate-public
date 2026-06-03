"""Local payments resource — country-specific payment providers.

TJ: Tcell, Korti Milli, Alif, Eskhata. UZ: Click, Payme, Apelsin, Uzcard,
Humo, Oson. Mirrors ``LocalPaymentsService`` (``local_payments.proto``) over
the gateway as REST. ADR-0009 governs which providers are wired per Mashgate
deployment.

Two surfaces are exposed here, matching the Go + TypeScript SDKs:

- Card / wallet / refund / provider-config surface (``/v1/local/...``) —
  mirrors the TypeScript ``LocalPaymentsResource``.
- Initiate / confirm / status / cancel surface (``/v1/local-payments/...``) —
  mirrors the Go ``LocalPaymentsClient``.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class LocalPaymentsResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Card payments (Uzcard / Humo) — OTP flow ──────────────────────

    def pay_by_card(
        self,
        *,
        provider: str,
        card_number: str,
        expiry_date: str,
        amount: str,
        currency: str | None = None,
        order_id: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Start a card payment (Uzcard / Humo). Triggers an OTP challenge."""
        body: dict[str, Any] = {
            "provider": provider,
            "cardNumber": card_number,
            "expiryDate": expiry_date,
            "amount": amount,
        }
        if currency is not None:
            body["currency"] = currency
        if order_id is not None:
            body["orderId"] = order_id
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/local/card-payment", body=body)

    def confirm_otp(self, *, payment_id: str, otp_code: str) -> dict[str, Any]:
        """Confirm a card payment with the OTP code sent to the cardholder."""
        return self._c.request(
            "POST",
            f"/v1/local/card-payment/{payment_id}/confirm-otp",
            body={"otpCode": otp_code},
        )

    # ── Mobile wallet payments (Click / Payme / Oson) — redirect flow ─

    def pay_by_wallet(
        self,
        *,
        provider: str,
        amount: str,
        phone: str | None = None,
        currency: str | None = None,
        order_id: str | None = None,
        return_url: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Start a mobile-wallet payment. Returns a ``redirectUrl`` to send the user to."""
        body: dict[str, Any] = {"provider": provider, "amount": amount}
        if phone is not None:
            body["phone"] = phone
        if currency is not None:
            body["currency"] = currency
        if order_id is not None:
            body["orderId"] = order_id
        if return_url is not None:
            body["returnUrl"] = return_url
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/local/wallet-payment", body=body)

    # ── Status ────────────────────────────────────────────────────────

    def get_payment(self, payment_id: str) -> dict[str, Any]:
        """Return the current state of a local payment."""
        return self._c.request("GET", f"/v1/local/payments/{payment_id}")

    # ── Refund ────────────────────────────────────────────────────────

    def refund(
        self,
        *,
        payment_id: str,
        amount: str | None = None,
        reason: str | None = None,
    ) -> dict[str, Any]:
        """Refund a local payment (full when ``amount`` omitted, else partial)."""
        body: dict[str, Any] = {"paymentId": payment_id}
        if amount is not None:
            body["amount"] = amount
        if reason is not None:
            body["reason"] = reason
        return self._c.request(
            "POST",
            f"/v1/local/payments/{payment_id}/refund",
            body=body,
        )

    # ── Provider configuration ────────────────────────────────────────

    def list_providers(self) -> dict[str, Any]:
        """List configured local-payment providers for the tenant."""
        return self._c.request("GET", "/v1/local/providers")

    def upsert_provider(
        self,
        *,
        tenant_id: str,
        provider: str,
        merchant_id: str,
        service_id: str | None = None,
        callback_url: str | None = None,
        enabled: bool | None = None,
        credentials: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Create or update a provider configuration for the tenant."""
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "provider": provider,
            "merchantId": merchant_id,
        }
        if service_id is not None:
            body["serviceId"] = service_id
        if callback_url is not None:
            body["callbackUrl"] = callback_url
        if enabled is not None:
            body["enabled"] = enabled
        if credentials is not None:
            body["credentials"] = credentials
        return self._c.request("PUT", "/v1/local/providers", body=body)

    # ── Webhook callback (provider → us) ──────────────────────────────

    def handle_callback(
        self,
        *,
        provider: str,
        payload: str,
        signature: str | None = None,
    ) -> dict[str, Any]:
        """Forward a raw provider callback for verification and processing."""
        body: dict[str, Any] = {"provider": provider, "payload": payload}
        if signature is not None:
            body["signature"] = signature
        return self._c.request("POST", "/v1/local/callback", body=body)

    # ── Initiate / confirm / status / cancel (mirrors Go LocalPaymentsClient) ──

    def list_supported_methods(
        self,
        *,
        tenant_id: str,
        country: str | None = None,
    ) -> dict[str, Any]:
        """Return the providers available for the tenant's country.

        E.g. for a TJ tenant: ``[{id: "tcell-mobile", name: "Tcell Mobile
        Money"}, {id: "korti-milli", name: "Korti Milli"}, ...]``.
        """
        return self._c.request(
            "GET",
            "/v1/local-payments/methods",
            query={"tenantId": tenant_id, "country": country},
        )

    def initiate_payment(
        self,
        *,
        tenant_id: str,
        method_id: str,
        amount: str,
        currency: str | None = None,
        order_id: str | None = None,
        phone: str | None = None,
        return_url: str | None = None,
        metadata: dict[str, str] | None = None,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        """Start a local-rail payment flow.

        Returns a provider-specific ``next_step`` (URL to redirect, USSD code
        to dial, QR to scan). Idempotent via the ``Idempotency-Key`` header.
        """
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "methodId": method_id,
            "amount": amount,
        }
        if currency is not None:
            body["currency"] = currency
        if order_id is not None:
            body["orderId"] = order_id
        if phone is not None:
            body["phone"] = phone
        if return_url is not None:
            body["returnUrl"] = return_url
        if metadata is not None:
            body["metadata"] = metadata
        headers: dict[str, str] = {}
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key
        return self._c.request(
            "POST",
            "/v1/local-payments/initiate",
            body=body,
            extra_headers=headers or None,
        )

    def confirm_payment(
        self,
        payment_id: str,
        *,
        code: str | None = None,
        otp: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Submit a provider callback / confirmation (e.g. SMS OTP, USSD code)."""
        body: dict[str, Any] = {}
        if code is not None:
            body["code"] = code
        if otp is not None:
            body["otp"] = otp
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request(
            "POST",
            f"/v1/local-payments/{payment_id}/confirm",
            body=body or None,
        )

    def get_payment_status(self, payment_id: str) -> dict[str, Any]:
        """Return the current state of a local payment."""
        return self._c.request("GET", f"/v1/local-payments/{payment_id}")

    def cancel_payment(self, payment_id: str, *, reason: str | None = None) -> None:
        """Cancel a pending local payment (best-effort; depends on provider)."""
        body: dict[str, Any] = {}
        if reason is not None:
            body["reason"] = reason
        self._c.request(
            "POST",
            f"/v1/local-payments/{payment_id}/cancel",
            body=body or None,
        )

    def list_payments(
        self,
        *,
        tenant_id: str,
        page: int | None = None,
        page_size: int | None = None,
    ) -> dict[str, Any]:
        """Return local payment history for the tenant."""
        return self._c.request(
            "GET",
            "/v1/local-payments",
            query={"tenantId": tenant_id, "page": page, "pageSize": page_size},
        )
