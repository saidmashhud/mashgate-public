import type { MashgateClient } from "../client.js";
import type {
  BillingPlan,
  BillingSubscription,
  BillingPaymentMethod,
  BillingInvoice,
  CreditBalance,
} from "../types.js";

// ── Plans ─────────────────────────────────────────────────────────────────

// (no request types needed for read-only plan endpoints)

// ── Subscription changes ──────────────────────────────────────────────────

export interface ChangePlanRequest {
  planId: string;
  immediate?: boolean;
}

export interface CancelPlanRequest {
  reason?: string;
  immediate?: boolean;
}

export interface PreviewPlanChangeRequest {
  planId: string;
}

export interface PreviewPlanChangeResponse {
  currentPlanId: string;
  newPlanId: string;
  proratedAmount: string;
  currency: string;
  effectiveDate: string;
}

// ── Payment methods ───────────────────────────────────────────────────────

export interface AddBillingPaymentMethodRequest {
  token: string;
  provider?: string;
  brand?: string;
  last4?: string;
  expMonth?: number;
  expYear?: number;
  setDefault?: boolean;
}

// ── Credits ───────────────────────────────────────────────────────────────

export interface RedeemPromoCodeRequest {
  code: string;
}

export interface RedeemPromoCodeResponse {
  creditAmount: string;
  currency: string;
  expiresAt?: string;
}

// ── Resource class ────────────────────────────────────────────────────────

export class BillingResource {
  constructor(private readonly client: MashgateClient) {}

  // Plans
  async listPlans(): Promise<BillingPlan[]> {
    const res = await this.client.request<{ plans: BillingPlan[] }>("GET", "/v1/billing/plans");
    return res.plans;
  }

  async getPlan(planId: string): Promise<BillingPlan> {
    return this.client.request<BillingPlan>("GET", `/v1/billing/plans/${planId}`);
  }

  // Subscription
  async getSubscription(): Promise<BillingSubscription> {
    return this.client.request<BillingSubscription>("GET", "/v1/billing/subscription");
  }

  async changePlan(data: ChangePlanRequest): Promise<BillingSubscription> {
    return this.client.request<BillingSubscription>("POST", "/v1/billing/subscription/change", {
      body: data,
    });
  }

  async cancelPlan(data?: CancelPlanRequest): Promise<BillingSubscription> {
    return this.client.request<BillingSubscription>("POST", "/v1/billing/subscription/cancel", {
      body: data,
    });
  }

  async previewPlanChange(data: PreviewPlanChangeRequest): Promise<PreviewPlanChangeResponse> {
    return this.client.request<PreviewPlanChangeResponse>(
      "POST",
      "/v1/billing/subscription/preview",
      { body: data },
    );
  }

  // Payment methods
  async listPaymentMethods(): Promise<BillingPaymentMethod[]> {
    const res = await this.client.request<{ paymentMethods: BillingPaymentMethod[] }>(
      "GET",
      "/v1/billing/payment-methods",
    );
    return res.paymentMethods;
  }

  async addPaymentMethod(data: AddBillingPaymentMethodRequest): Promise<BillingPaymentMethod> {
    return this.client.request<BillingPaymentMethod>("POST", "/v1/billing/payment-methods", {
      body: data,
    });
  }

  async setDefaultPaymentMethod(methodId: string): Promise<BillingPaymentMethod> {
    return this.client.request<BillingPaymentMethod>(
      "POST",
      `/v1/billing/payment-methods/${methodId}/default`,
    );
  }

  async removePaymentMethod(methodId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>(
      "DELETE",
      `/v1/billing/payment-methods/${methodId}`,
    );
    return res.success;
  }

  // Invoices
  async listInvoices(): Promise<BillingInvoice[]> {
    const res = await this.client.request<{ invoices: BillingInvoice[] }>(
      "GET",
      "/v1/billing/invoices",
    );
    return res.invoices;
  }

  async getInvoice(invoiceId: string): Promise<BillingInvoice> {
    return this.client.request<BillingInvoice>("GET", `/v1/billing/invoices/${invoiceId}`);
  }

  async payInvoice(invoiceId: string): Promise<BillingInvoice> {
    return this.client.request<BillingInvoice>("POST", `/v1/billing/invoices/${invoiceId}/pay`);
  }

  // Credits
  async getCreditBalance(): Promise<CreditBalance> {
    return this.client.request<CreditBalance>("GET", "/v1/billing/credits");
  }

  async redeemPromoCode(code: string): Promise<RedeemPromoCodeResponse> {
    return this.client.request<RedeemPromoCodeResponse>("POST", "/v1/billing/credits/redeem", {
      body: { code },
    });
  }
}
