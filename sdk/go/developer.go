package mashgate

import (
	"context"
	"fmt"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// DeveloperClient — self-service API keys + webhook endpoint inspection.
// Mirrors DeveloperService in developer.proto (8 RPCs).
//
// Note: API key management ALSO exists at top-level Client (CreateAPIKey via
// auth-service). DeveloperClient is the *tenant-facing self-service portal*
// surface — same backend, scoped to caller's tenant only.
// ────────────────────────────────────────────────────────────────────────────

type DeveloperClient struct {
	c *Client
}

// ── API keys ─────────────────────────────────────────────────────────────

// ListAPIKeys returns API keys belonging to the caller's tenant.
// Secret material is NEVER returned — only key id, prefix, scopes, created.
func (d *DeveloperClient) ListAPIKeys(ctx context.Context, tenantID string) ([]*APIKey, error) {
	path := fmt.Sprintf("/v1/developer/api-keys?tenantId=%s", url.QueryEscape(tenantID))
	var out struct {
		Keys []*APIKey `json:"keys"`
	}
	if err := d.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Keys, nil
}

// CreateAPIKey provisions a new API key. The secret is returned ONCE on creation
// — store it server-side; later listings expose only the prefix.
func (d *DeveloperClient) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyCreated, error) {
	var out APIKeyCreated
	if err := d.c.do(ctx, "POST", "/v1/developer/api-keys", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokeAPIKey permanently disables an API key. Active requests using it fail 401.
func (d *DeveloperClient) RevokeAPIKey(ctx context.Context, keyID string) error {
	return d.c.do(ctx, "DELETE", "/v1/developer/api-keys/"+keyID, nil, nil)
}

// ── Webhook endpoints (read-only inspection) ─────────────────────────────

// ListWebhookEndpoints — convenience wrapper around mg-events ListEndpoints.
// For full CRUD use top-level WebhookSubscription methods (events.go).
func (d *DeveloperClient) ListWebhookEndpoints(ctx context.Context, tenantID string) ([]*Endpoint, error) {
	path := fmt.Sprintf("/v1/developer/webhook-endpoints?tenantId=%s", url.QueryEscape(tenantID))
	var out struct {
		Endpoints []*Endpoint `json:"endpoints"`
	}
	if err := d.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Endpoints, nil
}

// ── Activity & health ────────────────────────────────────────────────────

// GetRecentActivity returns recent API call / webhook delivery summary
// shown in the developer dashboard.
func (d *DeveloperClient) GetRecentActivity(ctx context.Context, tenantID string, limit int) (*DeveloperActivity, error) {
	path := fmt.Sprintf("/v1/developer/activity?tenantId=%s", url.QueryEscape(tenantID))
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}
	var out DeveloperActivity
	if err := d.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIntegrationHealth returns integration status: API key validity,
// webhook delivery success rate, last error.
func (d *DeveloperClient) GetIntegrationHealth(ctx context.Context, tenantID string) (*IntegrationHealth, error) {
	path := fmt.Sprintf("/v1/developer/health?tenantId=%s", url.QueryEscape(tenantID))
	var out IntegrationHealth
	if err := d.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
