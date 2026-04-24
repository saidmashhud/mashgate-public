package mashgate

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Domain types
// ────────────────────────────────────────────────────────────────────────────

// RateLimitConfig is a per-tenant rate limit rule.
type RateLimitConfig struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
	RPM       int       `json:"rpm"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// IPBlocklistEntry is a blocked IP for a tenant.
type IPBlocklistEntry struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenantId"`
	IPAddress string     `json:"ipAddress"`
	Reason    string     `json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// GuardCheckResult is the response from a rate-limit + blocklist check.
type GuardCheckResult struct {
	Allowed   bool   `json:"allowed"`
	Remaining int    `json:"remaining"`
	ResetAt   int64  `json:"resetAt"`
	Reason    string `json:"reason,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// GuardCheckRequest is the input for a guard check.
type GuardCheckRequest struct {
	TenantID string `json:"tenantId"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	IP       string `json:"ip"`
}

// UpsertRateLimitRequest creates or updates a rate limit config.
type UpsertRateLimitRequest struct {
	TenantID string `json:"tenantId"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	RPM      int    `json:"rpm"`
}

// BlockIPRequest adds an IP to the blocklist.
type BlockIPRequest struct {
	TenantID  string     `json:"tenantId"`
	IP        string     `json:"ip"`
	Reason    string     `json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// GuardClient
// ────────────────────────────────────────────────────────────────────────────

// GuardClient provides access to the guard-service REST API.
type GuardClient struct {
	c *Client
}

// Check performs a rate-limit and IP-blocklist check.
func (g *GuardClient) Check(ctx context.Context, req GuardCheckRequest) (*GuardCheckResult, error) {
	var out GuardCheckResult
	if err := g.c.do(ctx, "POST", "/v1/guard/check", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpsertRateLimit creates or updates a rate limit config.
func (g *GuardClient) UpsertRateLimit(ctx context.Context, req UpsertRateLimitRequest) (*RateLimitConfig, error) {
	var out RateLimitConfig
	if err := g.c.do(ctx, "POST", "/v1/guard/rate-limits", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRateLimits returns all rate limit configs for a tenant.
func (g *GuardClient) ListRateLimits(ctx context.Context, tenantID string) ([]*RateLimitConfig, error) {
	path := fmt.Sprintf("/v1/guard/rate-limits?tenantId=%s", url.QueryEscape(tenantID))
	var out []*RateLimitConfig
	if err := g.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BlockIP adds an IP to the tenant blocklist.
func (g *GuardClient) BlockIP(ctx context.Context, req BlockIPRequest) (*IPBlocklistEntry, error) {
	var out IPBlocklistEntry
	if err := g.c.do(ctx, "POST", "/v1/guard/blocklist/ips", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UnblockIP removes an IP from the tenant blocklist.
func (g *GuardClient) UnblockIP(ctx context.Context, tenantID, ip string) (bool, error) {
	path := fmt.Sprintf("/v1/guard/blocklist/ips/%s?tenantId=%s",
		url.PathEscape(ip), url.QueryEscape(tenantID))
	var out struct {
		Success bool `json:"success"`
	}
	if err := g.c.do(ctx, "DELETE", path, nil, &out); err != nil {
		return false, err
	}
	return out.Success, nil
}

// ListBlockedIPs returns all blocked IPs for a tenant.
func (g *GuardClient) ListBlockedIPs(ctx context.Context, tenantID string) ([]*IPBlocklistEntry, error) {
	path := fmt.Sprintf("/v1/guard/blocklist/ips?tenantId=%s", url.QueryEscape(tenantID))
	var out []*IPBlocklistEntry
	if err := g.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
