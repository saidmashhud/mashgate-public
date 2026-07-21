from __future__ import annotations

import json

import httpx
import pytest

from mashgate import (
    AlertCategory,
    AlertSeverity,
    KycCheckType,
    KycSubjectType,
    MashgateClient,
    MashgateError,
    MerchantType,
)


def client_for(handler) -> MashgateClient:
    client = MashgateClient(
        base_url="https://example.test",
        api_key="mg_test_key",
        tenant_id="tenant-1",
    )
    client._http.close()
    client._http = httpx.Client(
        base_url="https://example.test", transport=httpx.MockTransport(handler)
    )
    return client


def test_kyc_request_includes_tenant_and_idempotency():
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path == "/v1/kyc/checks"
        assert request.headers["X-Tenant-ID"] == "tenant-1"
        assert request.headers["Idempotency-Key"] == "kyc-1"
        body = json.loads(request.content)
        assert body["tenantId"] == "tenant-1"
        assert body["subjectType"] == "KYC_SUBJECT_INDIVIDUAL"
        return httpx.Response(200, json={"check": {"checkId": "check-1"}})

    client = client_for(handler)
    result = client.kyc.request_check(
        subject_id="user-1",
        subject_type=KycSubjectType.INDIVIDUAL,
        check_type=KycCheckType.FULL,
        idempotency_key="kyc-1",
    )
    assert result["check"]["checkId"] == "check-1"
    client.close()


def test_compliance_raise_and_resolve_wire_contract():
    calls: list[tuple[str, dict]] = []

    def handler(request: httpx.Request) -> httpx.Response:
        calls.append((request.url.path, json.loads(request.content)))
        return httpx.Response(200, json={"alertId": "alert-1"})

    client = client_for(handler)
    client.compliance.raise_alert(
        subject_id="user-1",
        subject_type="user",
        category=AlertCategory.AML,
        severity=AlertSeverity.HIGH,
        source="screening",
        source_ref="screen-1",
        description="review",
        idempotency_key="alert-1",
    )
    client.compliance.resolve("alert-1", "cleared")
    assert calls[0][1]["tenantId"] == "tenant-1"
    assert calls[1][0].endswith("/alert-1/resolve")
    assert calls[1][1]["tenant_id"] == "tenant-1"
    client.close()


def test_merchant_onboard_and_suspend_wire_contract():
    calls: list[tuple[str, dict, str | None]] = []

    def handler(request: httpx.Request) -> httpx.Response:
        calls.append(
            (
                request.url.path,
                json.loads(request.content),
                request.headers.get("Idempotency-Key"),
            )
        )
        return httpx.Response(200, json={"merchantId": "merchant-1"})

    client = client_for(handler)
    client.merchant.onboard(
        subject_id="business-1",
        merchant_type=MerchantType.BUSINESS,
        display_name="Merchant",
        legal_name="Merchant LLC",
        registration_number="123",
        country_code="TJ",
        config={},
        idempotency_key="merchant-1",
    )
    client.merchant.suspend("merchant-1", "review")
    assert calls[0][1]["tenantId"] == "tenant-1"
    assert calls[0][2] == "merchant-1"
    assert calls[1][1]["suspend_reason"] == "review"
    client.close()


def test_fintech_resources_fail_fast_without_tenant():
    client = MashgateClient(base_url="https://example.test", api_key="mg_test_key")
    with pytest.raises(MashgateError) as exc:
        client.kyc.list()
    assert exc.value.code == "tenant_id_required"
    client.close()
