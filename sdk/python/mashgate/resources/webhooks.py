"""Webhooks resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class WebhooksResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create_endpoint(self, *, url: str, description: str | None = None, event_types: list[str] | None = None) -> dict[str, Any]:
        body: dict[str, Any] = {"url": url}
        if description is not None:
            body["description"] = description
        if event_types is not None:
            body["event_types"] = event_types
        return self._c.request("POST", "/v1/events/endpoints", body=body)

    def list_endpoints(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/events/endpoints")

    def get_endpoint(self, endpoint_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/events/endpoints/{endpoint_id}")

    def update_endpoint(self, endpoint_id: str, **kwargs: Any) -> dict[str, Any]:
        body: dict[str, Any] = {}
        for key in ("url", "description", "status"):
            if key in kwargs:
                body[key] = kwargs[key]
        if "event_types" in kwargs:
            body["event_types"] = kwargs["event_types"]
        return self._c.request("PUT", f"/v1/events/endpoints/{endpoint_id}", body=body)

    def delete_endpoint(self, endpoint_id: str) -> None:
        self._c.request("DELETE", f"/v1/events/endpoints/{endpoint_id}")

    def test_endpoint(self, endpoint_id: str) -> dict[str, Any]:
        return self._c.request("POST", f"/v1/events/endpoints/{endpoint_id}/test")

    def list_deliveries(self, endpoint_id: str, *, page: int | None = None, page_size: int | None = None) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/events/deliveries",
            query={"endpoint_id": endpoint_id, "page": page, "page_size": page_size},
        )

    def retry_delivery(self, endpoint_id: str, delivery_id: str) -> dict[str, Any]:
        _ = endpoint_id  # kept for API compatibility
        return self._c.request("POST", f"/v1/events/deliveries/{delivery_id}/retry")
