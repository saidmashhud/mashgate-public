"""Logs resource (mgLogs)."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class LogsResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def audit(
        self,
        *,
        tenant_id: str,
        actor: str | None = None,
        action: str | None = None,
        page: int | None = None,
        page_size: int | None = None,
        from_ms: int | None = None,
        to_ms: int | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/logs/audit",
            query={
                "tenantId": tenant_id,
                "actor": actor,
                "action": action,
                "page": page,
                "pageSize": page_size,
                "fromMs": from_ms,
                "toMs": to_ms,
            },
        )

    def activity(
        self,
        *,
        tenant_id: str,
        log_type: str | None = None,
        page: int | None = None,
        page_size: int | None = None,
        from_ms: int | None = None,
        to_ms: int | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/logs/activity",
            query={
                "tenantId": tenant_id,
                "logType": log_type,
                "page": page,
                "pageSize": page_size,
                "fromMs": from_ms,
                "toMs": to_ms,
            },
        )

    def payments(
        self,
        *,
        tenant_id: str,
        status: str | None = None,
        page: int | None = None,
        page_size: int | None = None,
        from_ms: int | None = None,
        to_ms: int | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/logs/payments",
            query={
                "tenantId": tenant_id,
                "status": status,
                "page": page,
                "pageSize": page_size,
                "fromMs": from_ms,
                "toMs": to_ms,
            },
        )

    def webhooks(
        self,
        *,
        tenant_id: str,
        endpoint_id: str | None = None,
        page: int | None = None,
        page_size: int | None = None,
        from_ms: int | None = None,
        to_ms: int | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/logs/webhooks",
            query={
                "tenantId": tenant_id,
                "endpointId": endpoint_id,
                "page": page,
                "pageSize": page_size,
                "fromMs": from_ms,
                "toMs": to_ms,
            },
        )
