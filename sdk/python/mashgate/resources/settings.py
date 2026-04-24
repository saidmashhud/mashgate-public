"""Settings resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class SettingsResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def get(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/settings")

    def update(self, **kwargs: Any) -> dict[str, Any]:
        body: dict[str, Any] = {}
        key_map = {
            "refund_enabled": "refundEnabled",
            "max_refund_amount": "maxRefundAmount",
            "max_refunds_per_day": "maxRefundsPerDay",
            "auto_refund_enabled": "autoRefundEnabled",
            "policies_json": "policiesJson",
        }
        for py_key, api_key in key_map.items():
            if py_key in kwargs:
                body[api_key] = kwargs[py_key]
        return self._c.request("PUT", "/v1/settings", body=body)
