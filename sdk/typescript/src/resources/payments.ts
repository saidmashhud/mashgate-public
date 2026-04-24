import type { MashgateClient } from "../client.js";
import type {
  CreatePaymentRequest,
  PaymentIntent,
  PaymentDetails,
  PaymentListResponse,
  CaptureRequest,
  RefundRequest,
  RefundSummary,
} from "../types.js";

export class PaymentsResource {
  constructor(private readonly client: MashgateClient) {}

  async create(data: CreatePaymentRequest): Promise<PaymentIntent> {
    const headers: Record<string, string> = {};
    if (data.idempotencyKey) {
      headers["X-Idempotency-Key"] = data.idempotencyKey;
    }
    return this.client.request<PaymentIntent>("POST", "/v1/payments", {
      body: data,
      headers,
    });
  }

  async get(paymentId: string): Promise<PaymentDetails> {
    return this.client.request<PaymentDetails>("GET", `/v1/payments/${paymentId}`);
  }

  async list(options?: { page?: number; pageSize?: number }): Promise<PaymentListResponse> {
    return this.client.request<PaymentListResponse>("GET", "/v1/payments", {
      query: {
        page: options?.page,
        page_size: options?.pageSize,
      },
    });
  }

  async authorize(
    paymentId: string,
    options?: { idempotencyKey?: string },
  ): Promise<PaymentDetails> {
    const headers: Record<string, string> = {};
    if (options?.idempotencyKey) {
      headers["X-Idempotency-Key"] = options.idempotencyKey;
    }
    return this.client.request<PaymentDetails>("POST", `/v1/payments/${paymentId}/authorize`, {
      headers,
    });
  }

  async capture(paymentId: string, data?: CaptureRequest): Promise<PaymentDetails> {
    const headers: Record<string, string> = {};
    if (data?.idempotencyKey) {
      headers["X-Idempotency-Key"] = data.idempotencyKey;
    }
    return this.client.request<PaymentDetails>("POST", `/v1/payments/${paymentId}/capture`, {
      body: data,
      headers,
    });
  }

  async void(paymentId: string): Promise<PaymentDetails> {
    return this.client.request<PaymentDetails>("POST", `/v1/payments/${paymentId}/void`);
  }

  async refund(paymentId: string, data: RefundRequest): Promise<RefundSummary> {
    const headers: Record<string, string> = {};
    if (data.idempotencyKey) {
      headers["X-Idempotency-Key"] = data.idempotencyKey;
    }
    return this.client.request<RefundSummary>("POST", `/v1/payments/${paymentId}/refund`, {
      body: data,
      headers,
    });
  }
}
