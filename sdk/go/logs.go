package mashgate

import (
	"context"
	"fmt"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// Domain types
// ────────────────────────────────────────────────────────────────────────────

// AuditLogEntry is a single entry from the audit log.
type AuditLogEntry struct {
	TenantID     string `json:"tenant_id"`
	AppID        string `json:"app_id,omitempty"`
	ActorID      string `json:"actor_id,omitempty"`
	ActorType    string `json:"actor_type,omitempty"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Changes      string `json:"changes,omitempty"`
	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	Ts           string `json:"ts"`
}

// LogsPage wraps a page of log entries with an optional cursor for pagination.
type LogsPage struct {
	Data       []*AuditLogEntry `json:"data"`
	NextCursor string           `json:"nextCursor,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// LogsQueryParams are common query parameters for log endpoints.
type LogsQueryParams struct {
	TenantID string
	From     string
	To       string
	Cursor   string
	Limit    int
}

func (p LogsQueryParams) toQuery() url.Values {
	q := url.Values{}
	q.Set("tenantId", p.TenantID)
	if p.From != "" {
		q.Set("from", p.From)
	}
	if p.To != "" {
		q.Set("to", p.To)
	}
	if p.Cursor != "" {
		q.Set("cursor", p.Cursor)
	}
	if p.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", p.Limit))
	}
	return q
}

// ────────────────────────────────────────────────────────────────────────────
// LogsClient
// ────────────────────────────────────────────────────────────────────────────

// LogsClient provides access to the logs-service REST API.
type LogsClient struct {
	c *Client
}

// Audit returns paginated audit log entries.
func (l *LogsClient) Audit(ctx context.Context, params LogsQueryParams, actor, action string) (*LogsPage, error) {
	q := params.toQuery()
	if actor != "" {
		q.Set("actor", actor)
	}
	if action != "" {
		q.Set("action", action)
	}
	var out LogsPage
	if err := l.c.do(ctx, "GET", "/v1/logs/audit?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Activity returns paginated activity log entries.
func (l *LogsClient) Activity(ctx context.Context, params LogsQueryParams, logType string) (*LogsPage, error) {
	q := params.toQuery()
	if logType != "" {
		q.Set("type", logType)
	}
	var out LogsPage
	if err := l.c.do(ctx, "GET", "/v1/logs/activity?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Payments returns paginated payment log entries.
func (l *LogsClient) Payments(ctx context.Context, params LogsQueryParams, status string) (*LogsPage, error) {
	q := params.toQuery()
	if status != "" {
		q.Set("status", status)
	}
	var out LogsPage
	if err := l.c.do(ctx, "GET", "/v1/logs/payments?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Webhooks returns paginated webhook delivery log entries.
func (l *LogsClient) Webhooks(ctx context.Context, params LogsQueryParams, endpointID string) (*LogsPage, error) {
	q := params.toQuery()
	if endpointID != "" {
		q.Set("endpoint_id", endpointID)
	}
	var out LogsPage
	if err := l.c.do(ctx, "GET", "/v1/logs/webhooks?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TrackEvent is a custom analytics event for ingestion.
type TrackEvent struct {
	TenantID string         `json:"tenantId"`
	Event    string         `json:"event"`
	Props    map[string]any `json:"props,omitempty"`
	Ts       int64          `json:"ts,omitempty"`
}

// Track ingests a custom analytics event into mgLogs.
func (l *LogsClient) Track(ctx context.Context, req TrackEvent) error {
	return l.c.do(ctx, "POST", "/v1/logs/ingest", req, nil)
}
