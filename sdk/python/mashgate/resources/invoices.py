"""Invoices resource.

Merchant invoices, mirroring the invoice-service REST API exposed via the
gateway.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class InvoicesResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create(
        self,
        *,
        tenant_id: str,
        amount: int,
        currency: str,
        customer_id: str | None = None,
        payment_id: str | None = None,
        subscription_id: str | None = None,
        line_items: list[dict[str, Any]] | None = None,
        due_date: str | None = None,
    ) -> dict[str, Any]:
        """Create a new invoice."""
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "amount": amount,
            "currency": currency,
        }
        if customer_id is not None:
            body["customerId"] = customer_id
        if payment_id is not None:
            body["paymentId"] = payment_id
        if subscription_id is not None:
            body["subscriptionId"] = subscription_id
        if line_items is not None:
            body["lineItems"] = line_items
        if due_date is not None:
            body["dueDate"] = due_date
        return self._c.request("POST", "/v1/invoices", body=body)

    def list(self, tenant_id: str, *, status: str | None = None) -> dict[str, Any]:
        """Return invoices for a tenant, optionally filtered by status."""
        return self._c.request(
            "GET",
            "/v1/invoices",
            query={"tenantId": tenant_id, "status": status},
        )

    def get(self, invoice_id: str) -> dict[str, Any]:
        """Return a single invoice by ID."""
        return self._c.request("GET", f"/v1/invoices/{invoice_id}")

    def get_pdf_url(self, invoice_id: str) -> str:
        """Return the relative URL for the invoice PDF.

        This does not perform a request — it just builds the path. Combine
        with the client's base URL to fetch the rendered PDF.
        """
        return f"/v1/invoices/{invoice_id}/pdf"

    def send(self, invoice_id: str) -> dict[str, Any]:
        """Email the invoice to the customer via notify-service.

        :returns: ``{"sent": bool}``.
        """
        return self._c.request("POST", f"/v1/invoices/{invoice_id}/send")

    def void(self, invoice_id: str) -> dict[str, Any]:
        """Void an invoice."""
        return self._c.request("POST", f"/v1/invoices/{invoice_id}/void")
