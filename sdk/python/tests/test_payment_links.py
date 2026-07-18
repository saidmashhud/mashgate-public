from __future__ import annotations

import json

import httpx

from mashgate import MashgateClient


def test_payment_links_use_structural_merchant_ownership():
    seen = 0

    def handler(request: httpx.Request) -> httpx.Response:
        nonlocal seen
        seen += 1
        if seen == 1:
            body = json.loads(request.content)
            assert body["merchantId"] == "merchant-1"
            assert body["description"] == "invoice 42"
            return httpx.Response(201, json={"id": "link-1", **body})
        assert request.url.params["tenantId"] == "tenant-1"
        assert request.url.params["merchantId"] == "merchant-1"
        return httpx.Response(200, json=[])

    client = MashgateClient(base_url="https://example.test")
    client._http.close()
    client._http = httpx.Client(
        base_url="https://example.test", transport=httpx.MockTransport(handler)
    )
    link = client.payment_links.create(
        tenant_id="tenant-1",
        merchant_id="merchant-1",
        amount=42,
        currency="USDT",
        description="invoice 42",
    )
    assert link["merchantId"] == "merchant-1"
    client.payment_links.list("tenant-1", "merchant-1")
    client.close()
