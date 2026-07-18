import { randomUUID } from "node:crypto";
import { MashgateClient } from "@mashgate/sdk";

const baseUrl = process.env.MASHGATE_API_URL;
const accessToken = process.env.MASHGATE_ACCESS_TOKEN;
if (!baseUrl || !accessToken) throw new Error("Set MASHGATE_API_URL and MASHGATE_ACCESS_TOKEN");

const client = new MashgateClient({
  baseUrl,
  apiKey: process.env.MASHGATE_API_KEY,
  accessToken,
});

const [{ markets }, { balances }] = await Promise.all([
  client.exchange.listMarkets(),
  client.exchange.listBalances(),
]);
console.log({ markets: markets.length, balances: balances.length });

if (process.env.PLACE_TEST_ORDER === "1") {
  const order = await client.exchange.placeOrder(
    {
      market: "BTC/USDT",
      side: "ORDER_SIDE_BUY",
      kind: "ORDER_KIND_LIMIT",
      timeInForce: "TIME_IN_FORCE_GTC",
      price: "1.00",
      quantity: "0.00001",
      postOnly: true,
    },
    `exchange-example-${randomUUID()}`,
  );
  console.log({ orderId: order.orderId, status: order.status });
}
