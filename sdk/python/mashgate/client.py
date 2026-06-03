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
from mashgate.resources.wallet_admin import WalletAdminResource
from mashgate.resources.billing import BillingResource
from mashgate.resources.subscriptions import SubscriptionsResource
from mashgate.resources.invoices import InvoicesResource
from mashgate.resources.payment_links import PaymentLinksResource
from mashgate.resources.iam import IamResource
from mashgate.resources.analytics import AnalyticsResource
from mashgate.resources.metering import MeteringResource
from mashgate.resources.mail import MailResource
from mashgate.resources.guard import GuardResource
from mashgate.resources.chain import ChainResource
from mashgate.resources.local_payments import LocalPaymentsResource


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
        # Admin/merchant-side WalletService — full wallet.v1.WalletService.
        # End-user wallet ops (saved cards, balance) live on `self.wallet`.
        self.wallet_admin = WalletAdminResource(self)
        # Platform billing (subscription, plans, credits) for the tenant.
        self.billing = BillingResource(self)
        # Recurring billing plans + customer subscriptions.
        self.subscriptions = SubscriptionsResource(self)
        # Merchant invoices.
        self.invoices = InvoicesResource(self)
        # Shareable payment links.
        self.payment_links = PaymentLinksResource(self)
        # IAM (mgID) — tenants, roles, groups, policies, API keys, scopes.
        self.iam = IamResource(self)
        # Payment + customer analytics (read-only).
        self.analytics = AnalyticsResource(self)
        # Usage metering + quota status.
        self.metering = MeteringResource(self)
        # Mail (mgMail) — mailboxes, messages, domains, DKIM.
        self.mail = MailResource(self)
        # Guard — per-tenant rate limiting + IP blocklisting.
        self.guard = GuardResource(self)
        # Chain (mgChain) — crypto rails: wallets, payments, swaps, escrow.
        self.chain = ChainResource(self)
        # Local payments — country-specific providers (TJ/UZ).
        self.local_payments = LocalPaymentsResource(self)

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
