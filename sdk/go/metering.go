package mashgate

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// ────────────────────────────────────────────────────────────────────────────
// MeteringClient — usage tracking input для billing.
// Mirrors MeteringService in metering.proto (3 RPCs).
//
// Verticals shouldn't call RecordUsage directly for metered actions — the
// platform records billable events automatically on Payment/Storage/Chain RPCs.
// This client exists for custom-metering use cases (e.g. AI tokens consumed).
// ────────────────────────────────────────────────────────────────────────────

type MeteringClient struct {
	c *Client
}

// RecordUsage emits a usage event. Idempotent via IdempotencyKey.
//
// MeterCode is one of platform-known meters (e.g. "api_call", "storage_gb_hour",
// "chain_tx_bytes") OR a tenant-custom meter registered via control-plane.
func (m *MeteringClient) RecordUsage(ctx context.Context, req RecordUsageRequest) (*UsageRecord, error) {
	key := req.IdempotencyKey
	if key == "" {
		key = uuid.NewString()
	}
	headers := map[string]string{"Idempotency-Key": key}
	var out UsageRecord
	if err := m.c.doWithHeader(ctx, "POST", "/v1/metering/usage", headers, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListUsage returns raw usage records over a time window for a tenant.
// Used for invoice line-item drilldown.
func (m *MeteringClient) ListUsage(ctx context.Context, params ListUsageParams) ([]*UsageRecord, error) {
	q := url.Values{}
	q.Set("tenantId", params.TenantID)
	if params.MeterCode != "" {
		q.Set("meterCode", params.MeterCode)
	}
	if !params.From.IsZero() {
		q.Set("from", params.From.Format(time.RFC3339))
	}
	if !params.To.IsZero() {
		q.Set("to", params.To.Format(time.RFC3339))
	}
	if params.PageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", params.PageSize))
	}
	var out struct {
		Records []*UsageRecord `json:"records"`
	}
	if err := m.c.do(ctx, "GET", "/v1/metering/usage?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return out.Records, nil
}

// GetUsageSummary returns aggregated usage by meter for billing-period.
// Drives invoice line items.
func (m *MeteringClient) GetUsageSummary(ctx context.Context, tenantID string, from, to time.Time) (*UsageSummary, error) {
	q := url.Values{
		"tenantId": {tenantID},
		"from":     {from.Format(time.RFC3339)},
		"to":       {to.Format(time.RFC3339)},
	}
	var out UsageSummary
	if err := m.c.do(ctx, "GET", "/v1/metering/summary?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
