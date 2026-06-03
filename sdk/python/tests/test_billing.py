"""Tests for the BillingResource — platform BillingService.

Mocks the gateway via ``respx`` (httpx route mocker), so we exercise the
real ``MashgateClient.request`` path including header / query / body
serialisation, but without a network round-trip.

The point of these tests is to PIN the REST paths to the proto contract.
Each ``respx`` route is registered at exactly one path; if the resource
drifts to a different path, respx finds no matching route and raises (or
the per-route ``route.called`` assertion fails). The regression-critical
subscription/credit paths additionally register the *wrong* legacy path
and assert it was NOT hit, so a silent re-route fails loudly.
"""

from __future__ import annotations

import json

import httpx
import pytest
import respx

from mashgate import MashgateClient


BASE = "https://api.mashgate.uz"


@pytest.fixture
def client() -> MashgateClient:
    return MashgateClient(base_url=BASE, api_key="mg_test_key")


# ── Plans ─────────────────────────────────────────────────────────────


@respx.mock
def test_list_plans_gets_plans_path(client):
    route = respx.get(f"{BASE}/v1/billing/plans").mock(
        return_value=httpx.Response(200, json={"plans": [{"id": "pro"}]})
    )
    out = client.billing.list_plans()
    assert out["plans"][0]["id"] == "pro"
    assert route.called
    assert route.calls.last.request.method == "GET"
    assert route.calls.last.request.url.path == "/v1/billing/plans"


@respx.mock
def test_get_plan_gets_plan_by_id_path(client):
    route = respx.get(f"{BASE}/v1/billing/plans/pro").mock(
        return_value=httpx.Response(200, json={"id": "pro", "name": "Pro"})
    )
    out = client.billing.get_plan("pro")
    assert out["id"] == "pro"
    assert route.called
    assert route.calls.last.request.method == "GET"
    assert route.calls.last.request.url.path == "/v1/billing/plans/pro"


# ── Subscription ───────────────────────────────────────────────────────


@respx.mock
def test_get_subscription_gets_subscription_path(client):
    route = respx.get(f"{BASE}/v1/billing/subscription").mock(
        return_value=httpx.Response(200, json={"plan_id": "pro", "status": "active"})
    )
    out = client.billing.get_subscription()
    assert out["plan_id"] == "pro"
    assert route.called
    assert route.calls.last.request.method == "GET"
    assert route.calls.last.request.url.path == "/v1/billing/subscription"


@respx.mock
def test_change_plan_posts_subscription_change_path(client):
    """Regression guard: ChangePlan -> POST /v1/billing/subscription/change."""
    route = respx.post(f"{BASE}/v1/billing/subscription/change").mock(
        return_value=httpx.Response(200, json={"plan_id": "enterprise"})
    )
    # Register the legacy/wrong path so a drift back to it fails loudly
    # instead of silently matching a different route.
    wrong = respx.post(f"{BASE}/v1/billing/subscription").mock(
        return_value=httpx.Response(200, json={})
    )

    out = client.billing.change_plan(plan_id="enterprise", immediate=True)
    assert out["plan_id"] == "enterprise"

    assert route.called
    assert not wrong.called
    assert route.calls.last.request.method == "POST"
    assert route.calls.last.request.url.path == "/v1/billing/subscription/change"

    sent = json.loads(route.calls.last.request.content)
    assert sent["planId"] == "enterprise"
    assert sent["immediate"] is True


@respx.mock
def test_change_plan_omits_immediate_when_none(client):
    route = respx.post(f"{BASE}/v1/billing/subscription/change").mock(
        return_value=httpx.Response(200, json={"plan_id": "enterprise"})
    )
    client.billing.change_plan(plan_id="enterprise")
    sent = json.loads(route.calls.last.request.content)
    assert sent == {"planId": "enterprise"}
    assert "immediate" not in sent


@respx.mock
def test_cancel_plan_posts_subscription_cancel_path(client):
    route = respx.post(f"{BASE}/v1/billing/subscription/cancel").mock(
        return_value=httpx.Response(200, json={"status": "canceled"})
    )
    out = client.billing.cancel_plan(reason="too-expensive", immediate=False)
    assert out["status"] == "canceled"
    assert route.called
    assert route.calls.last.request.method == "POST"
    assert route.calls.last.request.url.path == "/v1/billing/subscription/cancel"

    sent = json.loads(route.calls.last.request.content)
    assert sent["reason"] == "too-expensive"
    assert sent["immediate"] is False


@respx.mock
def test_cancel_plan_sends_no_body_when_no_args(client):
    """No reason/immediate -> body is None (request sends no JSON payload)."""
    route = respx.post(f"{BASE}/v1/billing/subscription/cancel").mock(
        return_value=httpx.Response(200, json={"status": "canceled"})
    )
    client.billing.cancel_plan()
    assert route.called
    assert route.calls.last.request.content == b""


@respx.mock
def test_preview_plan_change_posts_subscription_preview_path(client):
    """Regression guard: PreviewPlanChange -> POST /v1/billing/subscription/preview."""
    route = respx.post(f"{BASE}/v1/billing/subscription/preview").mock(
        return_value=httpx.Response(200, json={"proration": "12.50"})
    )
    # Guard against drift to the apply path.
    wrong = respx.post(f"{BASE}/v1/billing/subscription/change").mock(
        return_value=httpx.Response(200, json={})
    )

    out = client.billing.preview_plan_change(plan_id="enterprise")
    assert out["proration"] == "12.50"

    assert route.called
    assert not wrong.called
    assert route.calls.last.request.method == "POST"
    assert route.calls.last.request.url.path == "/v1/billing/subscription/preview"

    sent = json.loads(route.calls.last.request.content)
    assert sent == {"planId": "enterprise"}


# ── Payment methods ─────────────────────────────────────────────────────


@respx.mock
def test_list_payment_methods_gets_payment_methods_path(client):
    route = respx.get(f"{BASE}/v1/billing/payment-methods").mock(
        return_value=httpx.Response(200, json={"payment_methods": []})
    )
    out = client.billing.list_payment_methods()
    assert out == {"payment_methods": []}
    assert route.called
    assert route.calls.last.request.method == "GET"
    assert route.calls.last.request.url.path == "/v1/billing/payment-methods"


@respx.mock
def test_add_payment_method_posts_payment_methods_path(client):
    route = respx.post(f"{BASE}/v1/billing/payment-methods").mock(
        return_value=httpx.Response(200, json={"id": "pm-1"})
    )
    out = client.billing.add_payment_method(
        token="tok_visa",
        provider="stripe",
        brand="visa",
        last4="4242",
        exp_month=12,
        exp_year=2030,
        set_default=True,
    )
    assert out["id"] == "pm-1"
    assert route.called
    assert route.calls.last.request.method == "POST"
    assert route.calls.last.request.url.path == "/v1/billing/payment-methods"

    sent = json.loads(route.calls.last.request.content)
    assert sent["token"] == "tok_visa"
    assert sent["provider"] == "stripe"
    assert sent["brand"] == "visa"
    assert sent["last4"] == "4242"
    assert sent["expMonth"] == 12
    assert sent["expYear"] == 2030
    assert sent["setDefault"] is True


@respx.mock
def test_add_payment_method_omits_unset_optionals(client):
    route = respx.post(f"{BASE}/v1/billing/payment-methods").mock(
        return_value=httpx.Response(200, json={"id": "pm-1"})
    )
    client.billing.add_payment_method(token="tok_only")
    sent = json.loads(route.calls.last.request.content)
    assert sent == {"token": "tok_only"}


@respx.mock
def test_set_default_payment_method_posts_default_path(client):
    route = respx.post(f"{BASE}/v1/billing/payment-methods/pm-1/default").mock(
        return_value=httpx.Response(200, json={"id": "pm-1", "default": True})
    )
    out = client.billing.set_default_payment_method("pm-1")
    assert out["default"] is True
    assert route.called
    assert route.calls.last.request.method == "POST"
    assert (
        route.calls.last.request.url.path == "/v1/billing/payment-methods/pm-1/default"
    )


@respx.mock
def test_remove_payment_method_deletes_payment_method_path(client):
    route = respx.delete(f"{BASE}/v1/billing/payment-methods/pm-1").mock(
        return_value=httpx.Response(200, json={"deleted": True})
    )
    out = client.billing.remove_payment_method("pm-1")
    assert out["deleted"] is True
    assert route.called
    assert route.calls.last.request.method == "DELETE"
    assert route.calls.last.request.url.path == "/v1/billing/payment-methods/pm-1"


# ── Invoices ─────────────────────────────────────────────────────────────


@respx.mock
def test_list_invoices_gets_invoices_path(client):
    route = respx.get(f"{BASE}/v1/billing/invoices").mock(
        return_value=httpx.Response(200, json={"invoices": [{"id": "inv-1"}]})
    )
    out = client.billing.list_invoices()
    assert out["invoices"][0]["id"] == "inv-1"
    assert route.called
    assert route.calls.last.request.method == "GET"
    assert route.calls.last.request.url.path == "/v1/billing/invoices"


@respx.mock
def test_get_invoice_gets_invoice_by_id_path(client):
    route = respx.get(f"{BASE}/v1/billing/invoices/inv-1").mock(
        return_value=httpx.Response(200, json={"id": "inv-1", "status": "open"})
    )
    out = client.billing.get_invoice("inv-1")
    assert out["id"] == "inv-1"
    assert route.called
    assert route.calls.last.request.method == "GET"
    assert route.calls.last.request.url.path == "/v1/billing/invoices/inv-1"


@respx.mock
def test_pay_invoice_posts_invoice_pay_path(client):
    route = respx.post(f"{BASE}/v1/billing/invoices/inv-1/pay").mock(
        return_value=httpx.Response(200, json={"id": "inv-1", "status": "paid"})
    )
    out = client.billing.pay_invoice("inv-1")
    assert out["status"] == "paid"
    assert route.called
    assert route.calls.last.request.method == "POST"
    assert route.calls.last.request.url.path == "/v1/billing/invoices/inv-1/pay"


# ── Credits ──────────────────────────────────────────────────────────────


@respx.mock
def test_get_credit_balance_gets_credits_path(client):
    """Regression guard: GetCreditBalance -> GET /v1/billing/credits."""
    route = respx.get(f"{BASE}/v1/billing/credits").mock(
        return_value=httpx.Response(200, json={"balance": "100.00"})
    )
    # Guard against drift to the redeem sub-path / a pluralised variant.
    wrong = respx.get(f"{BASE}/v1/billing/credits/balance").mock(
        return_value=httpx.Response(200, json={})
    )

    out = client.billing.get_credit_balance()
    assert out["balance"] == "100.00"

    assert route.called
    assert not wrong.called
    assert route.calls.last.request.method == "GET"
    assert route.calls.last.request.url.path == "/v1/billing/credits"


@respx.mock
def test_redeem_promo_code_posts_credits_redeem_path(client):
    """Regression guard: RedeemPromoCode -> POST /v1/billing/credits/redeem."""
    route = respx.post(f"{BASE}/v1/billing/credits/redeem").mock(
        return_value=httpx.Response(200, json={"credited": "25.00"})
    )
    # Guard against drift to a top-level /promo path or the balance path.
    wrong = respx.post(f"{BASE}/v1/billing/credits").mock(
        return_value=httpx.Response(200, json={})
    )

    out = client.billing.redeem_promo_code("WELCOME25")
    assert out["credited"] == "25.00"

    assert route.called
    assert not wrong.called
    assert route.calls.last.request.method == "POST"
    assert route.calls.last.request.url.path == "/v1/billing/credits/redeem"

    sent = json.loads(route.calls.last.request.content)
    assert sent == {"code": "WELCOME25"}
