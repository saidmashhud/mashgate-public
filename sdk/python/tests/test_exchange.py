from __future__ import annotations

import json

import httpx

from mashgate import MashgateClient


def test_exchange_place_order_forwards_subject_token_and_idempotency():
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path == "/v1/exchange/orders"
        assert request.headers["Authorization"] == "Bearer user-token"
        assert request.headers["Idempotency-Key"] == "order-1"
        body = json.loads(request.content)
        assert "accountId" not in body
        return httpx.Response(
            200,
            json={"orderId": "order-id", "market": "BTC/USDT", "status": "ORDER_STATUS_ACCEPTED"},
        )

    client = MashgateClient(base_url="https://example.test", access_token="user-token")
    client._http.close()
    client._http = httpx.Client(
        base_url="https://example.test", transport=httpx.MockTransport(handler)
    )
    result = client.exchange.place_order(
        market="BTC/USDT",
        side="ORDER_SIDE_BUY",
        kind="ORDER_KIND_LIMIT",
        time_in_force="TIME_IN_FORCE_GTC",
        price="50000",
        quantity="0.001",
        idempotency_key="order-1",
    )
    assert result["market"] == "BTC/USDT"
    client.close()
