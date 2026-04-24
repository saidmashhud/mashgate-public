import type { MashgateClient } from "../client.js";
import type {
  CreateCheckoutSessionRequest,
  CheckoutSession,
  CompleteCheckoutRequest,
} from "../types.js";

export class CheckoutResource {
  constructor(private readonly client: MashgateClient) {}

  async createSession(data: CreateCheckoutSessionRequest): Promise<CheckoutSession> {
    return this.client.request<CheckoutSession>("POST", "/v1/checkout/sessions", { body: data });
  }

  async getSession(sessionId: string): Promise<CheckoutSession> {
    return this.client.request<CheckoutSession>("GET", `/v1/checkout/sessions/${sessionId}`);
  }

  async completeSession(data: CompleteCheckoutRequest): Promise<{ success: boolean; redirectUrl?: string }> {
    return this.client.request("POST", `/v1/checkout/sessions/${data.sessionId}/complete`, {
      body: data,
    });
  }
}
