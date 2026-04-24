"""Developer resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class DeveloperResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create_application(self, *, name: str, app_type: str, description: str | None = None) -> dict[str, Any]:
        body: dict[str, Any] = {"name": name, "appType": app_type}
        if description is not None:
            body["description"] = description
        return self._c.request("POST", "/v1/developer/applications", body=body)

    def list_applications(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/developer/applications")

    def get_application(self, app_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/developer/applications/{app_id}")

    def delete_application(self, app_id: str) -> None:
        self._c.request("DELETE", f"/v1/developer/applications/{app_id}")
