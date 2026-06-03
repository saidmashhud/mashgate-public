"""Payment links resource.

Shareable payment URLs, mirroring the payment-link endpoints on the
checkout-service exposed via the gateway.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class PaymentLinksResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create(
        self,
        *,
        tenant_id: str,
        amount: int,
        currency: str,
        description: str | None = None,
        expires_at: str | None = None,
    ) -> dict[str, Any]:
        """Create a new payment link."""
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "amount": amount,
            "currency": currency,
        }
        if description is not None:
            body["description"] = description
        if expires_at is not None:
            body["expiresAt"] = expires_at
        return self._c.request("POST", "/v1/payment-links", body=body)

    def list(self, tenant_id: str) -> dict[str, Any]:
        """Return all payment links for a tenant."""
        return self._c.request("GET", "/v1/payment-links", query={"tenantId": tenant_id})

    def get(self, link_id: str) -> dict[str, Any]:
        """Return a payment link by ID."""
        return self._c.request("GET", f"/v1/payment-links/{link_id}")
