import type { MashgateClient } from "../client.js";
import type { PaymentLink } from "../types.js";

export interface CreatePaymentLinkRequest {
  tenantId: string;
  amount: number;
  currency: string;
  description?: string;
  expiresAt?: string;
}

export class PaymentLinksResource {
  constructor(private readonly client: MashgateClient) {}

  async create(data: CreatePaymentLinkRequest): Promise<PaymentLink> {
    return this.client.request<PaymentLink>("POST", "/v1/payment-links", { body: data });
  }

  async list(tenantId: string): Promise<PaymentLink[]> {
    return this.client.request<PaymentLink[]>("GET", "/v1/payment-links", { query: { tenantId } });
  }

  async get(id: string): Promise<PaymentLink> {
    return this.client.request<PaymentLink>("GET", `/v1/payment-links/${id}`);
  }
}
