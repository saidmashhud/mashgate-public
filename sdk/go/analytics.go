package mashgate

import (
	"context"
	"fmt"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// AnalyticsClient — payment + customer analytics (read-only).
// Mirrors AnalyticsService in contracts/proto/v1/analytics.proto (12 RPCs).
// ────────────────────────────────────────────────────────────────────────────

type AnalyticsClient struct {
	c *Client
}

// AnalyticsPeriod — one of: "1d" | "7d" | "30d" | "90d" | "365d" | "custom".
type AnalyticsPeriod string

// Granularity — one of: "hour" | "day" | "week" | "month".
type Granularity string

// ── Payment analytics ─────────────────────────────────────────────────────

// GetPaymentMetrics returns aggregate payment volume/count/avg ticket for the period.
func (a *AnalyticsClient) GetPaymentMetrics(ctx context.Context, tenantID string, period AnalyticsPeriod) (*PaymentMetrics, error) {
	var out PaymentMetrics
	path := fmt.Sprintf("/v1/analytics/payments/metrics?tenantId=%s&period=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetVolumeTimeSeries returns payment volume bucketed by granularity over period.
func (a *AnalyticsClient) GetVolumeTimeSeries(ctx context.Context, tenantID string, period AnalyticsPeriod, gran Granularity) (*VolumeTimeSeries, error) {
	var out VolumeTimeSeries
	path := fmt.Sprintf("/v1/analytics/payments/volume?tenantId=%s&period=%s&granularity=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)), url.QueryEscape(string(gran)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTransactionCount returns transaction count bucketed by granularity over period.
func (a *AnalyticsClient) GetTransactionCount(ctx context.Context, tenantID string, period AnalyticsPeriod, gran Granularity) (*TransactionCountSeries, error) {
	var out TransactionCountSeries
	path := fmt.Sprintf("/v1/analytics/payments/transactions?tenantId=%s&period=%s&granularity=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)), url.QueryEscape(string(gran)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPaymentMethodBreakdown returns share of card/wallet/bank/local over period.
func (a *AnalyticsClient) GetPaymentMethodBreakdown(ctx context.Context, tenantID string, period AnalyticsPeriod) (*PaymentMethodBreakdown, error) {
	var out PaymentMethodBreakdown
	path := fmt.Sprintf("/v1/analytics/payments/methods?tenantId=%s&period=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetGeoDistribution returns volume/count grouped by customer country.
func (a *AnalyticsClient) GetGeoDistribution(ctx context.Context, tenantID string, period AnalyticsPeriod) (*GeoDistribution, error) {
	var out GeoDistribution
	path := fmt.Sprintf("/v1/analytics/payments/geo?tenantId=%s&period=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetFailureAnalysis returns failed-payment breakdown by reason code over period.
func (a *AnalyticsClient) GetFailureAnalysis(ctx context.Context, tenantID string, period AnalyticsPeriod) (*FailureAnalysis, error) {
	var out FailureAnalysis
	path := fmt.Sprintf("/v1/analytics/payments/failures?tenantId=%s&period=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ── Customer analytics ────────────────────────────────────────────────────

// GetCustomerMetrics returns customer cohort metrics over period (new/repeat/churned).
func (a *AnalyticsClient) GetCustomerMetrics(ctx context.Context, tenantID string, period AnalyticsPeriod) (*CustomerMetrics, error) {
	var out CustomerMetrics
	path := fmt.Sprintf("/v1/analytics/customers/metrics?tenantId=%s&period=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCohortAnalysis returns retention/revenue cohort matrix.
func (a *AnalyticsClient) GetCohortAnalysis(ctx context.Context, tenantID string, period AnalyticsPeriod) (*CohortAnalysis, error) {
	var out CohortAnalysis
	path := fmt.Sprintf("/v1/analytics/customers/cohorts?tenantId=%s&period=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCustomerSegments returns customer segments (VIP / casual / churned / new).
func (a *AnalyticsClient) GetCustomerSegments(ctx context.Context, tenantID string) ([]*CustomerSegment, error) {
	var out struct {
		Segments []*CustomerSegment `json:"segments"`
	}
	path := fmt.Sprintf("/v1/analytics/customers/segments?tenantId=%s", url.QueryEscape(tenantID))
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Segments, nil
}

// GetTopCustomers returns top customers by lifetime value (default limit 10).
func (a *AnalyticsClient) GetTopCustomers(ctx context.Context, tenantID string, period AnalyticsPeriod, limit int) ([]*TopCustomer, error) {
	var out struct {
		Customers []*TopCustomer `json:"customers"`
	}
	path := fmt.Sprintf("/v1/analytics/customers/top?tenantId=%s&period=%s",
		url.QueryEscape(tenantID), url.QueryEscape(string(period)))
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}
	if err := a.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Customers, nil
}
