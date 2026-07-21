import type { MashgateClient } from "../client.js";

export const MerchantStatus = {
  Pending: "MERCHANT_STATUS_PENDING",
  KycRequired: "MERCHANT_STATUS_KYC_REQUIRED",
  UnderReview: "MERCHANT_STATUS_UNDER_REVIEW",
  Accepted: "MERCHANT_STATUS_ACCEPTED",
  Rejected: "MERCHANT_STATUS_REJECTED",
  Suspended: "MERCHANT_STATUS_SUSPENDED",
  Offboarded: "MERCHANT_STATUS_OFFBOARDED",
} as const;
export type MerchantStatus = (typeof MerchantStatus)[keyof typeof MerchantStatus];

export const MerchantType = {
  Individual: "MERCHANT_TYPE_INDIVIDUAL",
  Business: "MERCHANT_TYPE_BUSINESS",
  DAO: "MERCHANT_TYPE_DAO",
} as const;
export type MerchantType = (typeof MerchantType)[keyof typeof MerchantType];

export interface MerchantConfig {
  acceptedCurrencies: string[];
  maxTransactionAmount: string;
  dailyVolumeLimit: string;
  monthlyVolumeLimit: string;
  primaryCurrency: string;
  cryptoEnabled: boolean;
  fiatEnabled: boolean;
  allowedPaymentMethods: string[];
  metadata?: Record<string, string>;
}

export interface MerchantProfile {
  merchantId: string;
  tenantId: string;
  subjectId: string;
  merchantType: MerchantType;
  status: MerchantStatus;
  displayName: string;
  legalName: string;
  registrationNumber: string;
  countryCode: string;
  kycCheckId: string;
  config: MerchantConfig;
  acceptedBy: string;
  rejectedBy: string;
  rejectionReason: string;
  suspendedBy: string;
  suspendReason: string;
  createdAt: string;
  updatedAt: string;
  acceptedAt?: string;
  suspendedAt?: string;
}

export interface OnboardMerchant {
  subjectId: string;
  merchantType: MerchantType;
  displayName: string;
  legalName: string;
  registrationNumber: string;
  countryCode: string;
  config: MerchantConfig;
}

export interface ListMerchantsQuery {
  status?: MerchantStatus;
  limit?: number;
  cursor?: string;
}

export interface ListMerchantsResponse {
  merchants: MerchantProfile[];
  nextCursor?: string;
}

export class MerchantResource {
  constructor(private readonly client: MashgateClient) {}

  onboard(data: OnboardMerchant, idempotencyKey: string): Promise<MerchantProfile> {
    return this.client.request("POST", "/v1/merchants", {
      body: { ...data, tenantId: this.client.requireTenantId() },
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  get(merchantId: string): Promise<MerchantProfile> {
    return this.client.request("GET", `/v1/merchants/${merchantId}`, {
      query: { tenant_id: this.client.requireTenantId() },
    });
  }

  list(query: ListMerchantsQuery = {}): Promise<ListMerchantsResponse> {
    return this.client.request("GET", "/v1/merchants", {
      query: {
        tenant_id: this.client.requireTenantId(),
        status: query.status,
        limit: query.limit,
        cursor: query.cursor,
      },
    });
  }

  accept(merchantId: string, note: string): Promise<MerchantProfile> {
    return this.transition(merchantId, "accept", { note });
  }

  reject(merchantId: string, rejectionReason: string): Promise<MerchantProfile> {
    return this.transition(merchantId, "reject", { rejection_reason: rejectionReason });
  }

  suspend(merchantId: string, suspendReason: string): Promise<MerchantProfile> {
    return this.transition(merchantId, "suspend", { suspend_reason: suspendReason });
  }

  reinstate(merchantId: string, note: string): Promise<MerchantProfile> {
    return this.transition(merchantId, "reinstate", { note });
  }

  private transition(
    merchantId: string,
    action: string,
    fields: Record<string, string>,
  ): Promise<MerchantProfile> {
    return this.client.request("POST", `/v1/merchants/${merchantId}/${action}`, {
      body: {
        tenant_id: this.client.requireTenantId(),
        merchant_id: merchantId,
        ...fields,
      },
    });
  }
}
