import type { MashgateClient } from "../client.js";

// ── Types ────────────────────────────────────────────────────────────────

export interface LocalPayment {
  paymentId: string;
  tenantId: string;
  provider: string;
  status: string;
  amount: string;
  currency: string;
  orderId?: string;
  cardLast4?: string;
  phone?: string;
  transactionId?: string;
  authCode?: string;
  errorCode?: string;
  errorMessage?: string;
  metadata?: Record<string, string>;
  createdAt: string;
  settledAt?: string;
}

export interface OtpConfirmResult {
  paymentId: string;
  status: string;
  transactionId?: string;
  authCode?: string;
}

export interface RedirectPaymentResult {
  paymentId: string;
  status: string;
  redirectUrl: string;
}

export interface LocalRefundResult {
  refundId: string;
  status: string;
  amount: string;
  provider: string;
}

export interface ProviderConfig {
  id: string;
  tenantId: string;
  provider: string;
  enabled: boolean;
  merchantId: string;
  serviceId?: string;
  callbackUrl?: string;
}

// ── Resource ─────────────────────────────────────────────────────────────

export class LocalPaymentsResource {
  constructor(private readonly client: MashgateClient) {}

  // Card payments (Uzcard / Humo) — OTP flow
  async payByCard(data: {
    provider: string;
    cardNumber: string;
    expiryDate: string;
    amount: string;
    currency?: string;
    orderId?: string;
    metadata?: Record<string, string>;
  }): Promise<LocalPayment> {
    return this.client.request("POST", "/v1/local/card-payment", { body: data });
  }

  async confirmOtp(data: {
    paymentId: string;
    otpCode: string;
  }): Promise<OtpConfirmResult> {
    return this.client.request("POST", `/v1/local/card-payment/${data.paymentId}/confirm-otp`, {
      body: { otpCode: data.otpCode },
    });
  }

  // Mobile wallet payments (Click / Payme / Oson) — redirect flow
  async payByWallet(data: {
    provider: string;
    phone?: string;
    amount: string;
    currency?: string;
    orderId?: string;
    returnUrl?: string;
    metadata?: Record<string, string>;
  }): Promise<RedirectPaymentResult> {
    return this.client.request("POST", "/v1/local/wallet-payment", { body: data });
  }

  // Get payment status
  async getPayment(paymentId: string): Promise<LocalPayment> {
    return this.client.request("GET", `/v1/local/payments/${paymentId}`);
  }

  // Refund
  async refund(data: {
    paymentId: string;
    amount?: string;
    reason?: string;
  }): Promise<LocalRefundResult> {
    return this.client.request("POST", `/v1/local/payments/${data.paymentId}/refund`, {
      body: data,
    });
  }

  // Provider configuration
  async listProviders(): Promise<ProviderConfig[]> {
    return this.client.request("GET", "/v1/local/providers");
  }

  async upsertProvider(data: {
    tenantId: string;
    provider: string;
    merchantId: string;
    serviceId?: string;
    callbackUrl?: string;
    enabled?: boolean;
    credentials?: Record<string, string>;
  }): Promise<ProviderConfig> {
    return this.client.request("PUT", "/v1/local/providers", { body: data });
  }

  // Webhook callback (provider → us)
  async handleCallback(data: {
    provider: string;
    payload: string;
    signature?: string;
  }): Promise<{ status: string }> {
    return this.client.request("POST", "/v1/local/callback", { body: data });
  }
}
