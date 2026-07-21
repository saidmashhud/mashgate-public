import type { MashgateClient } from "../client.js";

export const KycStatus = {
  Unspecified: "KYC_STATUS_UNSPECIFIED",
  Pending: "KYC_STATUS_PENDING",
  InReview: "KYC_STATUS_IN_REVIEW",
  Passed: "KYC_STATUS_PASSED",
  Failed: "KYC_STATUS_FAILED",
  Expired: "KYC_STATUS_EXPIRED",
  Overridden: "KYC_STATUS_OVERRIDDEN",
} as const;
export type KycStatus = (typeof KycStatus)[keyof typeof KycStatus];

export const KycSubjectType = {
  Individual: "KYC_SUBJECT_INDIVIDUAL",
  Business: "KYC_SUBJECT_BUSINESS",
} as const;
export type KycSubjectType = (typeof KycSubjectType)[keyof typeof KycSubjectType];

export const KycCheckType = {
  Identity: "KYC_CHECK_IDENTITY",
  AML: "KYC_CHECK_AML",
  Sanctions: "KYC_CHECK_SANCTIONS",
  PEP: "KYC_CHECK_PEP",
  Full: "KYC_CHECK_FULL",
} as const;
export type KycCheckType = (typeof KycCheckType)[keyof typeof KycCheckType];

export interface KycRiskSignal {
  code: string;
  description: string;
  severity: string;
}

export interface KycCheck {
  checkId: string;
  tenantId: string;
  subjectId: string;
  subjectType: KycSubjectType;
  checkType: KycCheckType;
  status: KycStatus;
  provider: string;
  providerRef: string;
  riskSignals: KycRiskSignal[];
  failureCode: string;
  failureReason: string;
  overrideBy: string;
  overrideNote: string;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
}

export interface RequestKycCheck {
  subjectId: string;
  subjectType: KycSubjectType;
  checkType: KycCheckType;
  provider?: string;
  metadata?: Record<string, string>;
}

export interface RequestKycCheckResponse {
  check: KycCheck;
  redirectUrl?: string;
}

export interface ListKycChecksQuery {
  subjectId?: string;
  status?: KycStatus;
  limit?: number;
  cursor?: string;
}

export interface ListKycChecksResponse {
  checks: KycCheck[];
  nextCursor?: string;
}

export class KycResource {
  constructor(private readonly client: MashgateClient) {}

  request(data: RequestKycCheck, idempotencyKey: string): Promise<RequestKycCheckResponse> {
    return this.client.request("POST", "/v1/kyc/checks", {
      body: { ...data, tenantId: this.client.requireTenantId() },
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  get(checkId: string): Promise<KycCheck> {
    return this.client.request("GET", `/v1/kyc/checks/${checkId}`, {
      query: { tenant_id: this.client.requireTenantId() },
    });
  }

  list(query: ListKycChecksQuery = {}): Promise<ListKycChecksResponse> {
    return this.client.request("GET", "/v1/kyc/checks", {
      query: {
        tenant_id: this.client.requireTenantId(),
        subject_id: query.subjectId,
        status: query.status,
        limit: query.limit,
        cursor: query.cursor,
      },
    });
  }

  override(checkId: string, status: KycStatus, overrideNote: string): Promise<KycCheck> {
    return this.client.request("POST", `/v1/kyc/checks/${checkId}/override`, {
      body: {
        tenantId: this.client.requireTenantId(),
        checkId,
        status,
        overrideNote,
      },
    });
  }
}
