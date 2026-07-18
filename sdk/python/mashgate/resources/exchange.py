"""Custodial spot Exchange API.

The access token determines tenant, subject, and exchange account. No method
accepts an account or wallet owner identifier.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any
from urllib.parse import quote

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class ExchangeResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def list_markets(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/exchange/markets")

    def get_order_book(self, market: str, *, depth: int = 50) -> dict[str, Any]:
        return self._c.request(
            "GET", "/v1/exchange/order-book", query={"market": market, "depth": depth}
        )

    def list_trades(
        self, market: str, *, limit: int = 50, cursor: str | None = None
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/exchange/trades",
            query={"market": market, "limit": limit, "cursor": cursor},
        )

    def list_balances(self) -> dict[str, Any]:
        return self._c.request("GET", "/v1/exchange/balances")

    def place_order(
        self,
        *,
        market: str,
        side: str,
        kind: str,
        time_in_force: str,
        quantity: str,
        idempotency_key: str,
        price: str | None = None,
        post_only: bool = False,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "market": market,
            "side": side,
            "kind": kind,
            "timeInForce": time_in_force,
            "quantity": quantity,
            "postOnly": post_only,
        }
        if price is not None:
            body["price"] = price
        return self._c.request(
            "POST",
            "/v1/exchange/orders",
            body=body,
            extra_headers={"Idempotency-Key": idempotency_key},
        )

    def list_orders(
        self,
        *,
        market: str | None = None,
        status: str | None = None,
        limit: int = 50,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/exchange/orders",
            query={"market": market, "status": status, "limit": limit, "cursor": cursor},
        )

    def get_order(self, order_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/exchange/orders/{quote(order_id, safe='')}")

    def cancel_order(self, order_id: str, *, idempotency_key: str) -> dict[str, Any]:
        return self._c.request(
            "POST",
            f"/v1/exchange/orders/{quote(order_id, safe='')}/cancel",
            body={},
            extra_headers={"Idempotency-Key": idempotency_key},
        )

    def get_deposit_address(self, *, asset: str, network: str) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/exchange/deposits/address",
            query={"asset": asset, "network": network},
        )

    def list_deposits(
        self,
        *,
        asset: str | None = None,
        status: str | None = None,
        limit: int = 50,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/exchange/deposits",
            query={"asset": asset, "status": status, "limit": limit, "cursor": cursor},
        )

    def request_withdrawal(
        self,
        *,
        asset: str,
        network: str,
        amount: str,
        destination: str,
        idempotency_key: str,
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/exchange/withdrawals",
            body={
                "asset": asset,
                "network": network,
                "amount": amount,
                "destination": destination,
            },
            extra_headers={"Idempotency-Key": idempotency_key},
        )

    def list_withdrawals(
        self,
        *,
        asset: str | None = None,
        status: str | None = None,
        limit: int = 50,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/exchange/withdrawals",
            query={"asset": asset, "status": status, "limit": limit, "cursor": cursor},
        )

    def get_withdrawal(self, withdrawal_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET", f"/v1/exchange/withdrawals/{quote(withdrawal_id, safe='')}"
        )
