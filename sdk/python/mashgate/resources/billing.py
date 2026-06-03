"""Billing resource.

Mirrors ``BillingService`` from ``contracts/proto/v1/billing.proto`` —
platform subscription, plans, invoices, payment methods, and credits for
the current tenant. Auth is tenant-scoped; pass an admin JWT or
service-account API key on the parent client.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class BillingResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Plans ─────────────────────────────────────────────────────────

    def list_plans(self) -> dict[str, Any]:
        """Return all platform plans available to the current tenant."""
        return self._c.request("GET", "/v1/billing/plans")

    def get_plan(self, plan_id: str) -> dict[str, Any]:
        """Retrieve a platform plan by ID."""
        return self._c.request("GET", f"/v1/billing/plans/{plan_id}")

    # ── Subscription ──────────────────────────────────────────────────

    def get_subscription(self) -> dict[str, Any]:
        """Return the current tenant's active subscription."""
        return self._c.request("GET", "/v1/billing/subscription")

    def change_plan(self, *, plan_id: str, immediate: bool | None = None) -> dict[str, Any]:
        """Switch the tenant's subscription to a new plan."""
        body: dict[str, Any] = {"planId": plan_id}
        if immediate is not None:
            body["immediate"] = immediate
        return self._c.request("POST", "/v1/billing/subscription/change", body=body)

    def cancel_plan(
        self,
        *,
        reason: str | None = None,
        immediate: bool | None = None,
    ) -> dict[str, Any]:
        """Cancel the current subscription.

        Effective at period end unless ``immediate`` is set.
        """
        body: dict[str, Any] = {}
        if reason is not None:
            body["reason"] = reason
        if immediate is not None:
            body["immediate"] = immediate
        return self._c.request("POST", "/v1/billing/subscription/cancel", body=body or None)

    def preview_plan_change(self, *, plan_id: str) -> dict[str, Any]:
        """Compute prorated charges + effective dates for a hypothetical plan
        switch without applying it. Useful for confirmation UIs."""
        return self._c.request(
            "POST", "/v1/billing/subscription/preview", body={"planId": plan_id}
        )

    # ── Payment methods ───────────────────────────────────────────────

    def list_payment_methods(self) -> dict[str, Any]:
        """Return all billing payment methods registered for the tenant."""
        return self._c.request("GET", "/v1/billing/payment-methods")

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
        """Register a new payment method for the tenant."""
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
        return self._c.request("POST", "/v1/billing/payment-methods", body=body)

    def set_default_payment_method(self, method_id: str) -> dict[str, Any]:
        """Mark a method as the auto-billing default."""
        return self._c.request(
            "POST", f"/v1/billing/payment-methods/{method_id}/default"
        )

    def remove_payment_method(self, method_id: str) -> dict[str, Any]:
        """Delete a billing payment method."""
        return self._c.request("DELETE", f"/v1/billing/payment-methods/{method_id}")

    # ── Invoices ──────────────────────────────────────────────────────

    def list_invoices(self) -> dict[str, Any]:
        """Return all billing invoices for the tenant."""
        return self._c.request("GET", "/v1/billing/invoices")

    def get_invoice(self, invoice_id: str) -> dict[str, Any]:
        """Retrieve a single billing invoice by ID."""
        return self._c.request("GET", f"/v1/billing/invoices/{invoice_id}")

    def pay_invoice(self, invoice_id: str) -> dict[str, Any]:
        """Trigger an immediate payment attempt on an open invoice."""
        return self._c.request("POST", f"/v1/billing/invoices/{invoice_id}/pay")

    # ── Credits ───────────────────────────────────────────────────────

    def get_credit_balance(self) -> dict[str, Any]:
        """Return the tenant's current credit balance."""
        return self._c.request("GET", "/v1/billing/credits")

    def redeem_promo_code(self, code: str) -> dict[str, Any]:
        """Apply a promo code to the tenant."""
        return self._c.request("POST", "/v1/billing/credits/redeem", body={"code": code})
