import type { MashgateClient } from "../client.js";
import type { Plan, Subscription } from "../types.js";

export interface CreatePlanRequest {
  tenantId: string;
  name: string;
  amount: number;
  currency: string;
  interval: "monthly" | "yearly" | "weekly";
  trialDays?: number;
}

export interface CreateSubscriptionRequest {
  tenantId: string;
  customerId: string;
  planId: string;
  paymentMethodToken: string;
}

export class SubscriptionsResource {
  constructor(private readonly client: MashgateClient) {}

  async createPlan(data: CreatePlanRequest): Promise<Plan> {
    return this.client.request<Plan>("POST", "/v1/subscriptions/plans", { body: data });
  }

  async listPlans(tenantId: string): Promise<Plan[]> {
    return this.client.request<Plan[]>("GET", "/v1/subscriptions/plans", { query: { tenantId } });
  }

  async create(data: CreateSubscriptionRequest): Promise<Subscription> {
    return this.client.request<Subscription>("POST", "/v1/subscriptions", { body: data });
  }

  async list(tenantId: string): Promise<Subscription[]> {
    return this.client.request<Subscription[]>("GET", "/v1/subscriptions", { query: { tenantId } });
  }

  async cancel(id: string): Promise<Subscription> {
    return this.client.request<Subscription>("POST", `/v1/subscriptions/${id}/cancel`);
  }

  async pause(id: string): Promise<Subscription> {
    return this.client.request<Subscription>("POST", `/v1/subscriptions/${id}/pause`);
  }

  async resume(id: string): Promise<Subscription> {
    return this.client.request<Subscription>("POST", `/v1/subscriptions/${id}/resume`);
  }
}
