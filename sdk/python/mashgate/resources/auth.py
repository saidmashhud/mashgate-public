"""Auth resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class AuthResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def register(
        self, *, email: str, password: str, tenant_id: str, full_name: str | None = None, role: str | None = None
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"email": email, "password": password, "tenantId": tenant_id}
        if full_name:
            body["fullName"] = full_name
        if role:
            body["role"] = role
        return self._c.request("POST", "/v1/auth/register", body=body)

    def login(self, *, email: str, password: str) -> dict[str, Any]:
        data = self._c.request("POST", "/v1/auth/login", body={"email": email, "password": password})
        if "accessToken" in data:
            self._c.set_access_token(data["accessToken"])
        return data

    def refresh_token(self, refresh_token: str) -> dict[str, Any]:
        data = self._c.request("POST", "/v1/auth/refresh-token", body={"refreshToken": refresh_token})
        if "accessToken" in data:
            self._c.set_access_token(data["accessToken"])
        return data

    def logout(self) -> None:
        self._c.request("POST", "/v1/auth/logout")
        self._c.set_access_token("")

    def get_profile(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/auth/profile")

    def get_capabilities(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/auth/capabilities")
