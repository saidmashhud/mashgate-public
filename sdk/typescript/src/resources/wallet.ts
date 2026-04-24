import type { MashgateClient } from "../client.js";
import type {
  AddPaymentMethodRequest,
  PaymentMethod,
  WalletBalance,
  WalletMovementsResponse,
} from "../types.js";

export class WalletResource {
  constructor(private readonly client: MashgateClient) {}

  async addPaymentMethod(data: AddPaymentMethodRequest): Promise<PaymentMethod> {
    return this.client.request<PaymentMethod>("POST", "/v1/wallet/payment-methods", { body: data });
  }

  async listPaymentMethods(): Promise<{ paymentMethods: PaymentMethod[] }> {
    return this.client.request("GET", "/v1/wallet/payment-methods");
  }

  async removePaymentMethod(paymentMethodId: string): Promise<void> {
    return this.client.request<void>("DELETE", `/v1/wallet/payment-methods/${paymentMethodId}`);
  }

  async setDefaultPaymentMethod(paymentMethodId: string): Promise<void> {
    return this.client.request<void>("POST", `/v1/wallet/payment-methods/${paymentMethodId}/default`);
  }

  async getBalance(currency?: string): Promise<WalletBalance> {
    return this.client.request<WalletBalance>("GET", "/v1/wallet/balance", {
      query: { currency },
    });
  }

  async listMovements(options?: {
    currency?: string;
    page?: number;
    pageSize?: number;
  }): Promise<WalletMovementsResponse> {
    return this.client.request<WalletMovementsResponse>("GET", "/v1/wallet/movements", {
      query: {
        currency: options?.currency,
        page: options?.page,
        page_size: options?.pageSize,
      },
    });
  }
}
