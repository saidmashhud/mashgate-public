"""Subscriptions resource.

Recurring billing plans and customer subscriptions, mirroring the
subscription-service REST API exposed via the gateway.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class SubscriptionsResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Plans ─────────────────────────────────────────────────────────

    def create_plan(
        self,
        *,
        tenant_id: str,
        name: str,
        amount: int,
        currency: str,
        interval: str,
        trial_days: int | None = None,
    ) -> dict[str, Any]:
        """Create a new subscription plan.

        :param interval: One of ``"monthly"``, ``"yearly"``, ``"weekly"``.
        """
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "name": name,
            "amount": amount,
            "currency": currency,
            "interval": interval,
        }
        if trial_days is not None:
            body["trialDays"] = trial_days
        return self._c.request("POST", "/v1/subscriptions/plans", body=body)

    def list_plans(self, tenant_id: str) -> dict[str, Any]:
        """Return all subscription plans for a tenant."""
        return self._c.request(
            "GET", "/v1/subscriptions/plans", query={"tenantId": tenant_id}
        )

    # ── Subscriptions ─────────────────────────────────────────────────

    def create(
        self,
        *,
        tenant_id: str,
        customer_id: str,
        plan_id: str,
        payment_method_token: str,
    ) -> dict[str, Any]:
        """Subscribe a customer to a plan."""
        return self._c.request(
            "POST",
            "/v1/subscriptions",
            body={
                "tenantId": tenant_id,
                "customerId": customer_id,
                "planId": plan_id,
                "paymentMethodToken": payment_method_token,
            },
        )

    def list(self, tenant_id: str) -> dict[str, Any]:
        """Return all subscriptions for a tenant."""
        return self._c.request("GET", "/v1/subscriptions", query={"tenantId": tenant_id})

    def cancel(self, subscription_id: str) -> dict[str, Any]:
        """Cancel a subscription."""
        return self._c.request("POST", f"/v1/subscriptions/{subscription_id}/cancel")

    def pause(self, subscription_id: str) -> dict[str, Any]:
        """Pause an active subscription."""
        return self._c.request("POST", f"/v1/subscriptions/{subscription_id}/pause")

    def resume(self, subscription_id: str) -> dict[str, Any]:
        """Resume a paused subscription."""
        return self._c.request("POST", f"/v1/subscriptions/{subscription_id}/resume")
