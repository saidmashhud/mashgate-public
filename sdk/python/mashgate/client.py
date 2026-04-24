"""Mashgate API client."""

from __future__ import annotations

from typing import Any

import httpx

from mashgate.errors import MashgateError
from mashgate.resources.auth import AuthResource
from mashgate.resources.payments import PaymentsResource
from mashgate.resources.checkout import CheckoutResource
from mashgate.resources.wallet import WalletResource
from mashgate.resources.risk import RiskResource
from mashgate.resources.webhooks import WebhooksResource
from mashgate.resources.developer import DeveloperResource
from mashgate.resources.settings import SettingsResource
from mashgate.resources.chat import ChatResource
from mashgate.resources.notify import NotifyResource
from mashgate.resources.storage import StorageResource
from mashgate.resources.logs import LogsResource
from mashgate.resources.flags import FlagsResource


class MashgateClient:
    """Main entry-point for the Mashgate Payment Gateway API.

    Usage::

        client = MashgateClient(
            base_url="https://sandbox.mashgate.uz",
            api_key="sk_test_...",
        )
        payment = client.payments.create(amount="10000", currency="UZS")
    """

    def __init__(
        self,
        *,
        base_url: str,
        api_key: str | None = None,
        access_token: str | None = None,
        timeout: float = 30.0,
        headers: dict[str, str] | None = None,
    ) -> None:
        self._base_url = base_url.rstrip("/")
        self._api_key = api_key
        self._access_token = access_token

        default_headers = {"Content-Type": "application/json"}
        if headers:
            default_headers.update(headers)

        self._http = httpx.Client(
            base_url=self._base_url,
            headers=default_headers,
            timeout=timeout,
        )

        # Resource namespaces
        self.auth = AuthResource(self)
        self.payments = PaymentsResource(self)
        self.checkout = CheckoutResource(self)
        self.wallet = WalletResource(self)
        self.risk = RiskResource(self)
        self.webhooks = WebhooksResource(self)
        self.developer = DeveloperResource(self)
        self.settings = SettingsResource(self)
        self.chat = ChatResource(self)
        self.notify = NotifyResource(self)
        self.storage = StorageResource(self)
        self.logs = LogsResource(self)
        self.flags = FlagsResource(self)

    # ── Token management ──────────────────────────────────────────────

    def set_access_token(self, token: str) -> None:
        self._access_token = token

    # ── Internal request helper ───────────────────────────────────────

    def request(
        self,
        method: str,
        path: str,
        *,
        body: dict[str, Any] | None = None,
        query: dict[str, Any] | None = None,
        extra_headers: dict[str, str] | None = None,
    ) -> Any:
        headers: dict[str, str] = {}

        if self._access_token:
            headers["Authorization"] = f"Bearer {self._access_token}"
        elif self._api_key:
            headers["X-API-Key"] = self._api_key

        if extra_headers:
            headers.update(extra_headers)

        # Strip None values from query params
        params = {k: v for k, v in (query or {}).items() if v is not None} or None

        try:
            resp = self._http.request(
                method,
                path,
                json=body,
                params=params,
                headers=headers,
            )
        except httpx.TimeoutException:
            raise MashgateError(
                "Request timed out",
                status=408,
                code="request_timeout",
                retryable=True,
            )
        except httpx.ConnectError:
            raise MashgateError(
                "Connection failed",
                status=0,
                code="network_error",
                retryable=True,
            )

        if resp.status_code == 204:
            return None

        if resp.status_code >= 400:
            try:
                data = resp.json()
            except Exception:
                data = {}
            raise MashgateError(
                data.get("message", resp.text),
                status=resp.status_code,
                code=data.get("code", "api_error"),
                retryable=resp.status_code == 429 or resp.status_code >= 500,
                details=data,
            )

        return resp.json()

    def close(self) -> None:
        self._http.close()

    def __enter__(self) -> MashgateClient:
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()
