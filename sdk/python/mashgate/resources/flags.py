"""Flags resource (mgFlags)."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class FlagsResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create(
        self,
        *,
        tenant_id: str,
        key: str,
        flag_type: str = "boolean",
        enabled: bool = False,
        rollout_percentage: int | None = None,
        description: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "key": key,
            "flagType": flag_type,
            "enabled": enabled,
        }
        if rollout_percentage is not None:
            body["rolloutPercentage"] = rollout_percentage
        if description is not None:
            body["description"] = description
        return self._c.request("POST", "/v1/flags", body=body)

    def list(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", "/v1/flags", query={"tenantId": tenant_id})

    def get(self, flag_key: str, *, tenant_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET", f"/v1/flags/{flag_key}", query={"tenantId": tenant_id}
        )

    def update(
        self,
        flag_key: str,
        *,
        tenant_id: str,
        enabled: bool | None = None,
        rollout_percentage: int | None = None,
        description: str | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id}
        if enabled is not None:
            body["enabled"] = enabled
        if rollout_percentage is not None:
            body["rolloutPercentage"] = rollout_percentage
        if description is not None:
            body["description"] = description
        return self._c.request("PUT", f"/v1/flags/{flag_key}", body=body)

    def evaluate(
        self,
        *,
        tenant_id: str,
        flag_key: str,
        entity_id: str | None = None,
        context: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id, "flagKey": flag_key}
        if entity_id is not None:
            body["entityId"] = entity_id
        if context is not None:
            body["context"] = context
        return self._c.request("POST", "/v1/flags/evaluate", body=body)
