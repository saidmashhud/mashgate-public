"""Metering resource (usage tracking input for billing).

Mirrors ``metering.v1.MeteringService`` exposed over the gateway as REST.

Surface unions the Go SDK (``sdk/go/metering.go``) and the TypeScript SDK
(``sdk/typescript/src/resources/metering.ts``):

- Raw usage events: :meth:`record_usage`, :meth:`list_usage`,
  :meth:`get_usage_summary` (Go — drives invoice line-item drilldown).
- Aggregated quota dashboards: :meth:`get_usage_resource_summary`,
  :meth:`get_usage_time_series`, :meth:`get_quota_status` (TS — drives the
  control-plane usage UI).

Verticals shouldn't call :meth:`record_usage` directly for metered actions
— the platform records billable events automatically on Payment / Storage /
Chain RPCs. This client exists for custom-metering use cases (e.g. AI tokens
consumed).
"""

from __future__ import annotations

import uuid
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class MeteringResource:
    """Usage + quota client."""

    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Raw usage events (Go surface) ─────────────────────────────────

    def record_usage(
        self,
        *,
        tenant_id: str,
        meter_code: str,
        quantity: float,
        reference_id: str | None = None,
        timestamp: str | None = None,
        metadata: dict[str, str] | None = None,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        """Emit a usage event. Idempotent via ``idempotency_key``.

        ``meter_code`` is one of the platform-known meters (e.g. ``"api_call"``,
        ``"storage_gb_hour"``, ``"chain_tx_bytes"``) OR a tenant-custom meter
        registered via the control-plane.

        If ``idempotency_key`` is omitted a random UUID is generated, matching
        the Go SDK behaviour.
        """
        key = idempotency_key or str(uuid.uuid4())
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "meterCode": meter_code,
            "quantity": quantity,
        }
        if reference_id is not None:
            body["referenceId"] = reference_id
        if timestamp is not None:
            body["timestamp"] = timestamp
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request(
            "POST",
            "/v1/metering/usage",
            body=body,
            extra_headers={"Idempotency-Key": key},
        )

    def list_usage(
        self,
        *,
        tenant_id: str,
        meter_code: str | None = None,
        from_: str | None = None,
        to: str | None = None,
        page_size: int | None = None,
    ) -> dict[str, Any]:
        """Raw usage records over a time window for a tenant.

        Used for invoice line-item drilldown. ``from_`` / ``to`` are RFC 3339
        timestamps.
        """
        return self._c.request(
            "GET",
            "/v1/metering/usage",
            query={
                "tenantId": tenant_id,
                "meterCode": meter_code,
                "from": from_,
                "to": to,
                "pageSize": page_size,
            },
        )

    def get_usage_summary(
        self, tenant_id: str, from_: str, to: str
    ) -> dict[str, Any]:
        """Aggregated usage by meter for a billing period.

        Drives invoice line items. ``from_`` / ``to`` are RFC 3339 timestamps.
        Mirrors the Go SDK ``GetUsageSummary`` (hits ``/v1/metering/summary``).
        """
        return self._c.request(
            "GET",
            "/v1/metering/summary",
            query={"tenantId": tenant_id, "from": from_, "to": to},
        )

    # ── Aggregated quota dashboards (TS surface) ──────────────────────

    def get_usage_resource_summary(self, tenant_id: str) -> dict[str, Any]:
        """Per-resource usage summary for the current billing period.

        Mirrors the TS SDK ``getUsageSummary`` — hits ``/v1/metering/usage``
        with only ``tenantId`` and returns ``{resources, periodStart,
        periodEnd}`` where each resource carries current / limit / projected /
        overage. Resource codes: ``API_CALLS``, ``WEBHOOK_DELIVERIES``,
        ``STORAGE_BYTES``, ``CHAT_MESSAGES``, ``SMS_SENT``, ``EMAIL_SENT``,
        ``TEAM_MEMBERS``.
        """
        return self._c.request(
            "GET", "/v1/metering/usage", query={"tenantId": tenant_id}
        )

    def get_usage_time_series(
        self, tenant_id: str, resource: str, start: str, end: str
    ) -> dict[str, Any]:
        """Time series of usage for a single ``resource`` over ``[start, end]``.

        ``start`` / ``end`` are RFC 3339 timestamps; ``resource`` is one of the
        usage resource codes (e.g. ``API_CALLS``).
        """
        return self._c.request(
            "GET",
            f"/v1/metering/usage/{resource}/timeseries",
            query={"tenantId": tenant_id, "start": start, "end": end},
        )

    def get_quota_status(self, tenant_id: str) -> dict[str, Any]:
        """Quota status across all resources for a tenant.

        Returns ``{quotas, anyExceeded, anyNearLimit}`` where each quota
        carries its limit, enforcement (``SOFT`` | ``HARD``) and the alert
        thresholds reached.
        """
        return self._c.request(
            "GET", "/v1/metering/quota", query={"tenantId": tenant_id}
        )
