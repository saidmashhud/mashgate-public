import type { MashgateClient } from "../client.js";
import type {
  PaymentMetrics,
  VolumeTimeSeries,
  TransactionCountSeries,
  PaymentMethodBreakdown,
  GeoDistribution,
  FailureAnalysis,
  CustomerMetrics,
  CohortAnalysis,
  CustomerSegment,
  TopCustomer,
} from "../types.js";

// ── Shared query helpers ──────────────────────────────────────────────────

export type AnalyticsPeriod = "1d" | "7d" | "30d" | "90d" | "365d" | "custom";
export type Granularity = "hour" | "day" | "week" | "month";

export interface AnalyticsQuery {
  tenantId: string;
  period: AnalyticsPeriod;
}

export interface TimeSeriesQuery extends AnalyticsQuery {
  granularity: Granularity;
}

export interface TopCustomersQuery {
  tenantId: string;
  period: AnalyticsPeriod;
  limit?: number;
}

// ── Resource class ────────────────────────────────────────────────────────

export class AnalyticsResource {
  constructor(private readonly client: MashgateClient) {}

  // Payment analytics
  async getPaymentMetrics(tenantId: string, period: AnalyticsPeriod): Promise<PaymentMetrics> {
    return this.client.request<PaymentMetrics>("GET", "/v1/analytics/payments/metrics", {
      query: { tenantId, period },
    });
  }

  async getVolumeTimeSeries(
    tenantId: string,
    period: AnalyticsPeriod,
    granularity: Granularity,
  ): Promise<VolumeTimeSeries> {
    return this.client.request<VolumeTimeSeries>("GET", "/v1/analytics/payments/volume", {
      query: { tenantId, period, granularity },
    });
  }

  async getTransactionCount(
    tenantId: string,
    period: AnalyticsPeriod,
    granularity: Granularity,
  ): Promise<TransactionCountSeries> {
    return this.client.request<TransactionCountSeries>(
      "GET",
      "/v1/analytics/payments/transactions",
      { query: { tenantId, period, granularity } },
    );
  }

  async getPaymentMethodBreakdown(
    tenantId: string,
    period: AnalyticsPeriod,
  ): Promise<PaymentMethodBreakdown> {
    return this.client.request<PaymentMethodBreakdown>("GET", "/v1/analytics/payments/methods", {
      query: { tenantId, period },
    });
  }

  async getGeoDistribution(
    tenantId: string,
    period: AnalyticsPeriod,
  ): Promise<GeoDistribution> {
    return this.client.request<GeoDistribution>("GET", "/v1/analytics/payments/geo", {
      query: { tenantId, period },
    });
  }

  async getFailureAnalysis(
    tenantId: string,
    period: AnalyticsPeriod,
  ): Promise<FailureAnalysis> {
    return this.client.request<FailureAnalysis>("GET", "/v1/analytics/payments/failures", {
      query: { tenantId, period },
    });
  }

  // Customer analytics
  async getCustomerMetrics(
    tenantId: string,
    period: AnalyticsPeriod,
  ): Promise<CustomerMetrics> {
    return this.client.request<CustomerMetrics>("GET", "/v1/analytics/customers/metrics", {
      query: { tenantId, period },
    });
  }

  async getCohortAnalysis(
    tenantId: string,
    period: AnalyticsPeriod,
  ): Promise<CohortAnalysis> {
    return this.client.request<CohortAnalysis>("GET", "/v1/analytics/customers/cohorts", {
      query: { tenantId, period },
    });
  }

  async getCustomerSegments(tenantId: string): Promise<CustomerSegment[]> {
    const res = await this.client.request<{ segments: CustomerSegment[] }>(
      "GET",
      "/v1/analytics/customers/segments",
      { query: { tenantId } },
    );
    return res.segments;
  }

  async getTopCustomers(
    tenantId: string,
    period: AnalyticsPeriod,
    limit?: number,
  ): Promise<TopCustomer[]> {
    const res = await this.client.request<{ customers: TopCustomer[] }>(
      "GET",
      "/v1/analytics/customers/top",
      { query: { tenantId, period, limit } },
    );
    return res.customers;
  }
}
