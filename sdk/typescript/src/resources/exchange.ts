import type { MashgateClient } from "../client.js";

export type ExchangeOrderSide = "ORDER_SIDE_BUY" | "ORDER_SIDE_SELL";
export type ExchangeOrderKind = "ORDER_KIND_LIMIT" | "ORDER_KIND_MARKET";
export type ExchangeTimeInForce = "TIME_IN_FORCE_GTC" | "TIME_IN_FORCE_IOC";
export type ExchangeOrderStatus =
  | "ORDER_STATUS_PENDING_RESERVE"
  | "ORDER_STATUS_ACCEPTED"
  | "ORDER_STATUS_PARTIALLY_FILLED"
  | "ORDER_STATUS_FILLED"
  | "ORDER_STATUS_CANCELLED"
  | "ORDER_STATUS_REJECTED";

export interface ExchangeMarket {
  market: string;
  baseAsset: string;
  quoteAsset: string;
  priceScale: number;
  quantityScale: number;
  minQuantity: string;
  minNotional: string;
  status: string;
}

export interface ExchangePriceLevel {
  price: string;
  quantity: string;
  orderCount: number;
}

export interface ExchangeOrderBook {
  market: string;
  sequence: string;
  bids: ExchangePriceLevel[];
  asks: ExchangePriceLevel[];
}

export interface ExchangeTrade {
  tradeId: string;
  market: string;
  price: string;
  quantity: string;
  quoteQuantity: string;
  takerSide: ExchangeOrderSide;
  executedAt: string;
}

export interface ExchangeBalance {
  asset: string;
  available: string;
  held: string;
  total: string;
}

export interface ExchangeOrder {
  orderId: string;
  accountId: string;
  market: string;
  side: ExchangeOrderSide;
  kind: ExchangeOrderKind;
  timeInForce: ExchangeTimeInForce;
  price?: string;
  quantity: string;
  filledQuantity: string;
  remainingQuantity: string;
  status: ExchangeOrderStatus;
  postOnly: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ExchangeDepositAddress {
  asset: string;
  network: string;
  address: string;
  memo?: string;
}

export interface ExchangeDeposit {
  depositId: string;
  asset: string;
  network: string;
  amount: string;
  address: string;
  txHash: string;
  confirmations: number;
  status: string;
  detectedAt: string;
  creditedAt?: string;
}

export interface ExchangeWithdrawal {
  withdrawalId: string;
  asset: string;
  network: string;
  amount: string;
  destination: string;
  status: string;
  txHash?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PlaceExchangeOrder {
  market: string;
  side: ExchangeOrderSide;
  kind: ExchangeOrderKind;
  timeInForce: ExchangeTimeInForce;
  price?: string;
  quantity: string;
  postOnly?: boolean;
}

export interface ExchangeOrderListQuery {
  market?: string;
  status?: ExchangeOrderStatus;
  limit?: number;
  cursor?: string;
}

export interface ExchangeTransferListQuery {
  asset?: string;
  status?: string;
  limit?: number;
  cursor?: string;
}

export class ExchangeResource {
  constructor(private readonly client: MashgateClient) {}

  async listMarkets(): Promise<{ markets: ExchangeMarket[] }> {
    return this.client.request("GET", "/v1/exchange/markets");
  }

  async getOrderBook(market: string, depth = 50): Promise<ExchangeOrderBook> {
    return this.client.request("GET", "/v1/exchange/order-book", {
      query: { market, depth },
    });
  }

  async listTrades(
    market: string,
    options: { limit?: number; cursor?: string } = {},
  ): Promise<{ trades: ExchangeTrade[]; nextCursor?: string }> {
    return this.client.request("GET", "/v1/exchange/trades", {
      query: { market, ...options },
    });
  }

  async listBalances(): Promise<{ balances: ExchangeBalance[] }> {
    return this.client.request("GET", "/v1/exchange/balances");
  }

  async placeOrder(data: PlaceExchangeOrder, idempotencyKey: string): Promise<ExchangeOrder> {
    return this.client.request("POST", "/v1/exchange/orders", {
      body: data,
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  async listOrders(
    query: ExchangeOrderListQuery = {},
  ): Promise<{ orders: ExchangeOrder[]; nextCursor?: string }> {
    return this.client.request("GET", "/v1/exchange/orders", {
      query: {
        market: query.market,
        status: query.status,
        limit: query.limit,
        cursor: query.cursor,
      },
    });
  }

  async getOrder(orderId: string): Promise<ExchangeOrder> {
    return this.client.request("GET", `/v1/exchange/orders/${encodeURIComponent(orderId)}`);
  }

  async cancelOrder(orderId: string, idempotencyKey: string): Promise<ExchangeOrder> {
    return this.client.request(
      "POST",
      `/v1/exchange/orders/${encodeURIComponent(orderId)}/cancel`,
      { body: {}, headers: { "Idempotency-Key": idempotencyKey } },
    );
  }

  async getDepositAddress(asset: string, network: string): Promise<ExchangeDepositAddress> {
    return this.client.request("GET", "/v1/exchange/deposits/address", {
      query: { asset, network },
    });
  }

  async listDeposits(
    query: ExchangeTransferListQuery = {},
  ): Promise<{ deposits: ExchangeDeposit[]; nextCursor?: string }> {
    return this.client.request("GET", "/v1/exchange/deposits", {
      query: {
        asset: query.asset,
        status: query.status,
        limit: query.limit,
        cursor: query.cursor,
      },
    });
  }

  async requestWithdrawal(
    data: { asset: string; network: string; amount: string; destination: string },
    idempotencyKey: string,
  ): Promise<ExchangeWithdrawal> {
    return this.client.request("POST", "/v1/exchange/withdrawals", {
      body: data,
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  async listWithdrawals(
    query: ExchangeTransferListQuery = {},
  ): Promise<{ withdrawals: ExchangeWithdrawal[]; nextCursor?: string }> {
    return this.client.request("GET", "/v1/exchange/withdrawals", {
      query: {
        asset: query.asset,
        status: query.status,
        limit: query.limit,
        cursor: query.cursor,
      },
    });
  }

  async getWithdrawal(withdrawalId: string): Promise<ExchangeWithdrawal> {
    return this.client.request(
      "GET",
      `/v1/exchange/withdrawals/${encodeURIComponent(withdrawalId)}`,
    );
  }
}
