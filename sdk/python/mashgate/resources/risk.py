"""Risk resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class RiskResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def assess(self, *, amount: str, currency: str, customer_id: str | None = None, metadata: dict[str, str] | None = None) -> dict[str, Any]:
        body: dict[str, Any] = {"amount": amount, "currency": currency}
        if customer_id is not None:
            body["customerId"] = customer_id
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/risk/assess/payment", body=body)

    def investigate(self, *, payment_id: str) -> dict[str, Any]:
        return self._c.request("POST", "/v1/risk/investigate", body={"paymentId": payment_id})

    # ── Rules ─────────────────────────────────────────────────────────

    def list_rules(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/risk/rules")

    def create_rule(self, *, name: str, rule_type: str, condition: str, weight: float, action: str, description: str | None = None) -> dict[str, Any]:
        body: dict[str, Any] = {"name": name, "ruleType": rule_type, "condition": condition, "weight": weight, "action": action}
        if description is not None:
            body["description"] = description
        return self._c.request("POST", "/v1/risk/rules", body=body)

    def get_rule(self, rule_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/risk/rules/{rule_id}")

    def update_rule(self, rule_id: str, **kwargs: Any) -> dict[str, Any]:
        body: dict[str, Any] = {}
        key_map = {"name": "name", "description": "description", "condition": "condition", "weight": "weight", "action": "action", "enabled": "enabled"}
        for py_key, api_key in key_map.items():
            if py_key in kwargs:
                body[api_key] = kwargs[py_key]
        return self._c.request("PUT", f"/v1/risk/rules/{rule_id}", body=body)

    def delete_rule(self, rule_id: str) -> None:
        self._c.request("DELETE", f"/v1/risk/rules/{rule_id}")

    # ── Blocklist ─────────────────────────────────────────────────────

    def list_blocklist(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/risk/blocklist")

    def add_blocklist_entry(self, *, entry_type: str, value: str, reason: str | None = None, expires_in_seconds: int | None = None) -> dict[str, Any]:
        body: dict[str, Any] = {"entryType": entry_type, "value": value}
        if reason is not None:
            body["reason"] = reason
        if expires_in_seconds is not None:
            body["expiresInSeconds"] = expires_in_seconds
        return self._c.request("POST", "/v1/risk/blocklist", body=body)

    def remove_blocklist_entry(self, entry_id: str) -> None:
        self._c.request("DELETE", f"/v1/risk/blocklist/{entry_id}")
