import os
import uuid

from mashgate import MashgateClient


base_url = os.environ.get("MASHGATE_API_URL")
access_token = os.environ.get("MASHGATE_ACCESS_TOKEN")
if not base_url or not access_token:
    raise SystemExit("Set MASHGATE_API_URL and MASHGATE_ACCESS_TOKEN")

with MashgateClient(
    base_url=base_url,
    api_key=os.environ.get("MASHGATE_API_KEY"),
    access_token=access_token,
) as client:
    markets = client.exchange.list_markets()["markets"]
    balances = client.exchange.list_balances()["balances"]
    print({"markets": len(markets), "balances": len(balances)})

    if os.environ.get("PLACE_TEST_ORDER") == "1":
        order = client.exchange.place_order(
            market="BTC/USDT",
            side="ORDER_SIDE_BUY",
            kind="ORDER_KIND_LIMIT",
            time_in_force="TIME_IN_FORCE_GTC",
            price="1.00",
            quantity="0.00001",
            post_only=True,
            idempotency_key=f"exchange-example-{uuid.uuid4()}",
        )
        print({"orderId": order["orderId"], "status": order["status"]})
