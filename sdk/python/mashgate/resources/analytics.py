"""Analytics resource (read-only).

Mirrors ``analytics.v1.AnalyticsService`` exposed over the gateway as REST.
Surface matches the Go SDK (``sdk/go/analytics.go``) and TypeScript SDK
(``sdk/typescript/src/resources/analytics.ts``) method-for-method.

All endpoints are GET with ``tenantId`` + ``period`` query params; time
series additionally take ``granularity``. ``period`` is one of
``"1d" | "7d" | "30d" | "90d" | "365d" | "custom"``; ``granularity`` is
one of ``"hour" | "day" | "week" | "month"``.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class AnalyticsResource:
    """Payment + customer analytics client."""

    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Payment analytics ─────────────────────────────────────────────

    def get_payment_metrics(self, tenant_id: str, period: str) -> dict[str, Any]:
        """Aggregate payment volume / count / avg ticket for the period."""
        return self._c.request(
            "GET",
            "/v1/analytics/payments/metrics",
            query={"tenantId": tenant_id, "period": period},
        )

    def get_volume_time_series(
        self, tenant_id: str, period: str, granularity: str
    ) -> dict[str, Any]:
        """Payment volume bucketed by ``granularity`` over the period."""
        return self._c.request(
            "GET",
            "/v1/analytics/payments/volume",
            query={
                "tenantId": tenant_id,
                "period": period,
                "granularity": granularity,
            },
        )

    def get_transaction_count(
        self, tenant_id: str, period: str, granularity: str
    ) -> dict[str, Any]:
        """Transaction count bucketed by ``granularity`` over the period."""
        return self._c.request(
            "GET",
            "/v1/analytics/payments/transactions",
            query={
                "tenantId": tenant_id,
                "period": period,
                "granularity": granularity,
            },
        )

    def get_payment_method_breakdown(
        self, tenant_id: str, period: str
    ) -> dict[str, Any]:
        """Share of card / wallet / bank / local over the period."""
        return self._c.request(
            "GET",
            "/v1/analytics/payments/methods",
            query={"tenantId": tenant_id, "period": period},
        )

    def get_geo_distribution(self, tenant_id: str, period: str) -> dict[str, Any]:
        """Volume / count grouped by customer country."""
        return self._c.request(
            "GET",
            "/v1/analytics/payments/geo",
            query={"tenantId": tenant_id, "period": period},
        )

    def get_failure_analysis(self, tenant_id: str, period: str) -> dict[str, Any]:
        """Failed-payment breakdown by reason code over the period."""
        return self._c.request(
            "GET",
            "/v1/analytics/payments/failures",
            query={"tenantId": tenant_id, "period": period},
        )

    # ── Customer analytics ────────────────────────────────────────────

    def get_customer_metrics(self, tenant_id: str, period: str) -> dict[str, Any]:
        """Customer cohort metrics over the period (new / repeat / churned)."""
        return self._c.request(
            "GET",
            "/v1/analytics/customers/metrics",
            query={"tenantId": tenant_id, "period": period},
        )

    def get_cohort_analysis(self, tenant_id: str, period: str) -> dict[str, Any]:
        """Retention / revenue cohort matrix."""
        return self._c.request(
            "GET",
            "/v1/analytics/customers/cohorts",
            query={"tenantId": tenant_id, "period": period},
        )

    def get_customer_segments(self, tenant_id: str) -> dict[str, Any]:
        """Customer segments (VIP / casual / churned / new)."""
        return self._c.request(
            "GET",
            "/v1/analytics/customers/segments",
            query={"tenantId": tenant_id},
        )

    def get_top_customers(
        self, tenant_id: str, period: str, limit: int | None = None
    ) -> dict[str, Any]:
        """Top customers by lifetime value (default backend limit 10)."""
        return self._c.request(
            "GET",
            "/v1/analytics/customers/top",
            query={"tenantId": tenant_id, "period": period, "limit": limit},
        )
