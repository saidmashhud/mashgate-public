"""Guard resource — guard-service REST API.

Per-tenant rate limiting and IP blocklisting. Mirrors the guard-service
routes exposed over the gateway as REST.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class GuardResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def check(
        self,
        *,
        tenant_id: str,
        path: str,
        method: str,
        ip: str,
    ) -> dict[str, Any]:
        """Perform a rate-limit and IP-blocklist check.

        Returns ``{"allowed": bool, "remaining": int, "resetAt": int,
        "reason": str}``.
        """
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "path": path,
            "method": method,
            "ip": ip,
        }
        return self._c.request("POST", "/v1/guard/check", body=body)

    def upsert_rate_limit(
        self,
        *,
        tenant_id: str,
        path: str,
        method: str,
        rpm: int,
    ) -> dict[str, Any]:
        """Create or update a rate-limit config for a tenant."""
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "path": path,
            "method": method,
            "rpm": rpm,
        }
        return self._c.request("POST", "/v1/guard/rate-limits", body=body)

    def list_rate_limits(self, tenant_id: str) -> dict[str, Any]:
        """Return all rate-limit configs for a tenant."""
        return self._c.request(
            "GET", "/v1/guard/rate-limits", query={"tenantId": tenant_id}
        )

    def block_ip(
        self,
        *,
        tenant_id: str,
        ip: str,
        reason: str | None = None,
        expires_at: int | None = None,
    ) -> dict[str, Any]:
        """Add an IP to the tenant blocklist.

        :param expires_at: Optional epoch timestamp; omit for a permanent block.
        """
        body: dict[str, Any] = {"tenantId": tenant_id, "ip": ip}
        if reason is not None:
            body["reason"] = reason
        if expires_at is not None:
            body["expiresAt"] = expires_at
        return self._c.request("POST", "/v1/guard/blocklist/ips", body=body)

    def unblock_ip(self, tenant_id: str, ip: str) -> dict[str, Any]:
        """Remove an IP from the tenant blocklist.

        Returns ``{"success": bool}``.
        """
        return self._c.request(
            "DELETE",
            f"/v1/guard/blocklist/ips/{ip}",
            query={"tenantId": tenant_id},
        )

    def list_blocked_ips(self, tenant_id: str) -> dict[str, Any]:
        """Return all blocked IPs for a tenant."""
        return self._c.request(
            "GET", "/v1/guard/blocklist/ips", query={"tenantId": tenant_id}
        )
