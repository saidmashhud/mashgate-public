package mashgate

import (
	"context"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// SettingsClient — tenant-level configuration blob.
//
// Stores arbitrary string/bool key-value pairs scoped to tenant. Used for
// per-tenant feature flags (lightweight; full feature_flags.proto for richer
// rule-based flags), notification preferences, branding, etc.
// ────────────────────────────────────────────────────────────────────────────

type SettingsClient struct {
	c *Client
}

// GetSettings returns the entire tenant settings blob.
// Filter optionally by namespace prefix (e.g. "notify.", "branding.").
func (s *SettingsClient) GetSettings(ctx context.Context, tenantID, namespace string) (map[string]string, error) {
	q := "?tenantId=" + url.QueryEscape(tenantID)
	if namespace != "" {
		q += "&namespace=" + url.QueryEscape(namespace)
	}
	var out struct {
		Settings map[string]string `json:"settings"`
	}
	if err := s.c.do(ctx, "GET", "/v1/settings"+q, nil, &out); err != nil {
		return nil, err
	}
	return out.Settings, nil
}

// GetSetting retrieves a single setting by key. Returns ("", false, nil) if missing.
func (s *SettingsClient) GetSetting(ctx context.Context, tenantID, key string) (string, bool, error) {
	q := "?tenantId=" + url.QueryEscape(tenantID)
	var out struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Found bool   `json:"found"`
	}
	if err := s.c.do(ctx, "GET", "/v1/settings/"+url.PathEscape(key)+q, nil, &out); err != nil {
		return "", false, err
	}
	return out.Value, out.Found, nil
}

// UpdateSetting upserts one key. Partial update; other keys unaffected.
func (s *SettingsClient) UpdateSetting(ctx context.Context, tenantID, key, value string) error {
	req := struct {
		TenantID string `json:"tenantId"`
		Key      string `json:"key"`
		Value    string `json:"value"`
	}{TenantID: tenantID, Key: key, Value: value}
	return s.c.do(ctx, "PUT", "/v1/settings/"+url.PathEscape(key), req, nil)
}

// UpdateSettings bulk-upserts. Atomic per tenant.
func (s *SettingsClient) UpdateSettings(ctx context.Context, tenantID string, settings map[string]string) error {
	req := struct {
		TenantID string            `json:"tenantId"`
		Settings map[string]string `json:"settings"`
	}{TenantID: tenantID, Settings: settings}
	return s.c.do(ctx, "PATCH", "/v1/settings", req, nil)
}

// DeleteSetting removes a key.
func (s *SettingsClient) DeleteSetting(ctx context.Context, tenantID, key string) error {
	q := "?tenantId=" + url.QueryEscape(tenantID)
	return s.c.do(ctx, "DELETE", "/v1/settings/"+url.PathEscape(key)+q, nil, nil)
}
