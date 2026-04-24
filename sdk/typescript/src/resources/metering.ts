import type { MashgateClient } from "../client.js";

export type UsageResource =
  | "API_CALLS"
  | "WEBHOOK_DELIVERIES"
  | "STORAGE_BYTES"
  | "CHAT_MESSAGES"
  | "SMS_SENT"
  | "EMAIL_SENT"
  | "TEAM_MEMBERS";

export type QuotaEnforcement = "SOFT" | "HARD";

export interface ResourceUsage {
  resource: UsageResource;
  current: number;
  limit: number;
  percentage: number;
  projected: number;
  overage: number;
  displayName: string;
}

export interface UsageSummary {
  resources: ResourceUsage[];
  periodStart: string;
  periodEnd: string;
}

export interface UsageTimeSeriesPoint {
  timestamp: string;
  value: number;
}

export interface UsageTimeSeriesResponse {
  resource: UsageResource;
  points: UsageTimeSeriesPoint[];
  total: number;
  limit: number;
  projected: number;
}

export interface ResourceQuota {
  resource: UsageResource;
  limit: number;
  enforcement: QuotaEnforcement;
  alertThresholds: number[];
  threshold75Reached: boolean;
  threshold90Reached: boolean;
  threshold100Reached: boolean;
}

export interface QuotaStatus {
  quotas: ResourceQuota[];
  anyExceeded: boolean;
  anyNearLimit: boolean;
}

export class MeteringResource {
  constructor(private readonly client: MashgateClient) {}

  async getUsageSummary(tenantId: string): Promise<UsageSummary> {
    return this.client.request<UsageSummary>("GET", "/v1/metering/usage", {
      query: { tenantId },
    });
  }

  async getUsageTimeSeries(
    tenantId: string,
    resource: UsageResource,
    start: string,
    end: string,
  ): Promise<UsageTimeSeriesResponse> {
    return this.client.request<UsageTimeSeriesResponse>(
      "GET",
      `/v1/metering/usage/${resource}/timeseries`,
      { query: { tenantId, start, end } },
    );
  }

  async getQuotaStatus(tenantId: string): Promise<QuotaStatus> {
    return this.client.request<QuotaStatus>("GET", "/v1/metering/quota", {
      query: { tenantId },
    });
  }
}
