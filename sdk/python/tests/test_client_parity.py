"""Structural parity test — guards the Python SDK against drift from the Go/TS SDKs.

The three SDKs are generated from one source of truth and must expose the same
25 resource namespaces. A constructed :class:`MashgateClient` is the canonical
surface, so we assert every namespace is wired, non-``None``, and an instance
of a ``*Resource`` class. This fails loudly if a resource is dropped, renamed,
left unwired in ``client.py``, or swapped for a non-resource value.

No network is touched — construction alone wires every namespace.
"""

from __future__ import annotations

import importlib

import pytest

from mashgate import (
    MashgateClient,
    MashgateError,
    verify_webhook_signature,
)


BASE = "https://api.mashgate.uz"


# The 25 resource namespaces that MUST exist on the client, in lockstep with
# the Go and TS SDKs. Maps the attribute name -> the wired class name.
EXPECTED_RESOURCES: dict[str, str] = {
    "auth": "AuthResource",
    "payments": "PaymentsResource",
    "checkout": "CheckoutResource",
    "wallet": "WalletResource",
    "risk": "RiskResource",
    "webhooks": "WebhooksResource",
    "developer": "DeveloperResource",
    "settings": "SettingsResource",
    "chat": "ChatResource",
    "notify": "NotifyResource",
    "storage": "StorageResource",
    "logs": "LogsResource",
    "flags": "FlagsResource",
    "wallet_admin": "WalletAdminResource",
    "billing": "BillingResource",
    "subscriptions": "SubscriptionsResource",
    "invoices": "InvoicesResource",
    "payment_links": "PaymentLinksResource",
    "iam": "IamResource",
    "analytics": "AnalyticsResource",
    "metering": "MeteringResource",
    "mail": "MailResource",
    "guard": "GuardResource",
    "chain": "ChainResource",
    "local_payments": "LocalPaymentsResource",
}


@pytest.fixture
def client() -> MashgateClient:
    return MashgateClient(base_url=BASE, api_key="mg_test_key")


def test_expected_resource_count_is_25():
    """Tripwire — if the parity set itself shrinks/grows, the suite must notice."""
    assert len(EXPECTED_RESOURCES) == 25
    # No accidental duplicate keys collapsing the dict.
    assert len(set(EXPECTED_RESOURCES)) == 25


@pytest.mark.parametrize("attr", sorted(EXPECTED_RESOURCES))
def test_resource_attribute_present_and_non_none(client, attr):
    assert hasattr(client, attr), f"client is missing resource namespace: {attr!r}"
    resource = getattr(client, attr)
    assert resource is not None, f"client.{attr} is None — resource not wired"


@pytest.mark.parametrize(
    "attr,class_name", sorted(EXPECTED_RESOURCES.items())
)
def test_resource_is_wired_to_expected_class(client, attr, class_name):
    resource = getattr(client, attr)
    # The concrete class wired in client.py must match the parity contract …
    assert type(resource).__name__ == class_name, (
        f"client.{attr} is a {type(resource).__name__}, expected {class_name}"
    )
    # … and, defensively, every namespace must be a *Resource (no value/dict swap).
    assert type(resource).__name__.endswith("Resource"), (
        f"client.{attr} class {type(resource).__name__!r} does not end with 'Resource'"
    )


def test_no_unexpected_resource_namespaces(client):
    """Catch additions that land in client.py but never made it into the
    parity contract (and thus were never reconciled with Go/TS)."""
    # *Resource-typed public attributes actually attached to the instance.
    wired = {
        name
        for name in vars(client)
        if not name.startswith("_")
        and type(getattr(client, name)).__name__.endswith("Resource")
    }
    assert wired == set(EXPECTED_RESOURCES), (
        "resource namespaces on the client diverge from the parity contract: "
        f"unexpected={sorted(wired - set(EXPECTED_RESOURCES))} "
        f"missing={sorted(set(EXPECTED_RESOURCES) - wired)}"
    )


def test_top_level_exports_present():
    """Public API surface — these three names anchor the package's contract."""
    mod = importlib.import_module("mashgate")
    for name in ("MashgateClient", "MashgateError", "verify_webhook_signature"):
        assert hasattr(mod, name), f"mashgate package does not export {name!r}"
        assert name in mod.__all__, f"{name!r} missing from mashgate.__all__"

    # The imported references are the real objects, not None placeholders.
    assert MashgateClient is mod.MashgateClient
    assert MashgateError is mod.MashgateError
    assert verify_webhook_signature is mod.verify_webhook_signature


def test_mashgate_error_is_exception_subclass():
    assert isinstance(MashgateError, type)
    assert issubclass(MashgateError, Exception)


def test_verify_webhook_signature_is_callable():
    assert callable(verify_webhook_signature)


def test_representative_methods_exist_and_are_callable(client):
    """Spot-check that the newer resources expose a real, callable method —
    not just a bare class. ``billing.list_plans`` is the canonical example."""
    assert callable(client.billing.list_plans), (
        "client.billing.list_plans must be a callable method"
    )

    # These namespaces must each be a live resource object (not None / not a
    # class) so callers can reach their methods.
    for attr in (
        "subscriptions",
        "invoices",
        "payment_links",
        "iam",
        "metering",
        "local_payments",
    ):
        resource = getattr(client, attr)
        assert resource is not None
        # An instance, not the class object itself.
        assert not isinstance(resource, type), (
            f"client.{attr} is a class, expected an instance"
        )
        assert type(resource).__name__.endswith("Resource")
