package mashgate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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

type guardConditionWire struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type guardRuleWire struct {
	ID         string               `json:"id"`
	TenantID   string               `json:"tenantId"`
	Name       string               `json:"name"`
	Resource   string               `json:"resource"`
	Action     string               `json:"action"`
	Conditions []guardConditionWire `json:"conditions"`
	Priority   int                  `json:"priority"`
	CreatedAt  time.Time            `json:"createdAt"`
	UpdatedAt  time.Time            `json:"updatedAt"`
}

func rateLimitFromGuardRule(rule guardRuleWire) *RateLimitConfig {
	method := "ANY"
	for _, condition := range rule.Conditions {
		if strings.EqualFold(condition.Field, "method") && strings.TrimSpace(condition.Value) != "" {
			method = strings.ToUpper(strings.TrimSpace(condition.Value))
			break
		}
	}
	return &RateLimitConfig{
		ID: rule.ID, TenantID: rule.TenantID, Path: rule.Resource, Method: method,
		RPM: rule.Priority, CreatedAt: rule.CreatedAt, UpdatedAt: rule.UpdatedAt,
	}
}

func guardDecisionAllows(raw json.RawMessage) bool {
	value := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	return value == "1" || value == "ALLOW" || value == "GUARD_ACTION_ALLOW"
}

// Check performs a rate-limit and IP-blocklist check.
func (g *GuardClient) Check(ctx context.Context, req GuardCheckRequest) (*GuardCheckResult, error) {
	var out struct {
		Decision json.RawMessage `json:"decision"`
		Reason   string          `json:"reason"`
	}
	body := struct {
		TenantID string            `json:"tenantId"`
		Resource string            `json:"resource"`
		Action   string            `json:"action"`
		Context  map[string]string `json:"context"`
	}{
		TenantID: req.TenantID,
		Resource: req.Path,
		Action:   strings.ToUpper(req.Method),
		Context:  map[string]string{"ip": req.IP},
	}
	if err := g.c.do(ctx, "POST", "/v1/guard/evaluate", body, &out); err != nil {
		return nil, err
	}
	return &GuardCheckResult{Allowed: guardDecisionAllows(out.Decision), Reason: out.Reason}, nil
}

// UpsertRateLimit creates or updates a rate limit config.
func (g *GuardClient) UpsertRateLimit(ctx context.Context, req UpsertRateLimitRequest) (*RateLimitConfig, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "ANY"
	}
	body := struct {
		TenantID   string               `json:"tenantId"`
		Name       string               `json:"name"`
		Resource   string               `json:"resource"`
		Action     string               `json:"action"`
		Conditions []guardConditionWire `json:"conditions"`
		Priority   int                  `json:"priority"`
	}{
		TenantID:   req.TenantID,
		Name:       req.Method + " " + req.Path,
		Resource:   req.Path,
		Action:     "ALLOW",
		Conditions: []guardConditionWire{{Field: "method", Operator: "eq", Value: method}},
		Priority:   req.RPM,
	}
	var out guardRuleWire
	if err := g.c.do(ctx, "POST", "/v1/guard/rules", body, &out); err != nil {
		return nil, err
	}
	return rateLimitFromGuardRule(out), nil
}

// ListRateLimits returns all rate limit configs for a tenant.
func (g *GuardClient) ListRateLimits(ctx context.Context, tenantID string) ([]*RateLimitConfig, error) {
	path := fmt.Sprintf("/v1/guard/rules?tenantId=%s", url.QueryEscape(tenantID))
	var out struct {
		Rules []guardRuleWire `json:"rules"`
	}
	if err := g.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	result := make([]*RateLimitConfig, 0, len(out.Rules))
	for _, rule := range out.Rules {
		result = append(result, rateLimitFromGuardRule(rule))
	}
	return result, nil
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
