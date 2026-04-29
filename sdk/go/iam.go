package mashgate

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Tenant — Mashgate tenant identity (read-only mirror в downstream продуктах).
//
// Полную SoT держит Mashgate IAM (auth-service); downstream consumers получают
// `tenant.created/suspended/deleted` events на Kafka topic `tenant-events`
// (см. ADR-0020) или вызывают этот endpoint для cold-start backfill.
type Tenant struct {
	TenantID  string            `json:"tenantId"`
	Code      string            `json:"code"`
	Name      string            `json:"name"`
	Mode      string            `json:"mode"`
	Status    string            `json:"status"`
	PlanID    string            `json:"planId,omitempty"`
	PlanName  string            `json:"planName,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"createdAt,omitempty"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
	UserCount int32             `json:"userCount,omitempty"`
}

// ListTenantsOptions — optional фильтры. Соответствует ListTenantsRequest
// в `contracts/proto/v1/iam.proto`. Пустые поля → backend defaults.
type ListTenantsOptions struct {
	Status    string // "active" | "suspended" | "pending" | ""
	Search    string // free-text search by code/name
	Page      int32  // 1-based page; 0 = first page
	PageSize  int32  // 0 = backend default
	SortBy    string // "created_at" | "name" | "code"
	SortOrder string // "asc" | "desc"
}

type listAPIKeysResponse struct {
	Keys []*APIKey `json:"keys"`
}

type listTenantsResponse struct {
	Tenants    []Tenant `json:"tenants"`
	TotalCount int32    `json:"totalCount,omitempty"`
}

// ListTenants returns Mashgate tenants visible to the authenticated principal.
//
// Calls GET /v1/iam/tenants. Used for cold-start backfill в downstream verticals
// перед subscription к Kafka tenant-events (см. ADR-0020 Phase B).
func (c *Client) ListTenants(ctx context.Context, opts *ListTenantsOptions) ([]Tenant, error) {
	path := "/v1/iam/tenants"
	if opts != nil {
		params := []string{}
		if opts.Status != "" {
			params = append(params, "status="+url.QueryEscape(opts.Status))
		}
		if opts.Search != "" {
			params = append(params, "search="+url.QueryEscape(opts.Search))
		}
		if opts.Page > 0 {
			params = append(params, "page="+strconv.Itoa(int(opts.Page)))
		}
		if opts.PageSize > 0 {
			params = append(params, "pageSize="+strconv.Itoa(int(opts.PageSize)))
		}
		if opts.SortBy != "" {
			params = append(params, "sortBy="+url.QueryEscape(opts.SortBy))
		}
		if opts.SortOrder != "" {
			params = append(params, "sortOrder="+url.QueryEscape(opts.SortOrder))
		}
		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}
	}
	var out listTenantsResponse
	if err := c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Tenants, nil
}

// ListAPIKeys returns all API keys for the authenticated tenant.
//
// Calls GET /v1/iam/api-keys.
func (c *Client) ListAPIKeys(ctx context.Context) ([]*APIKey, error) {
	var out listAPIKeysResponse
	if err := c.do(ctx, "GET", "/v1/iam/api-keys", nil, &out); err != nil {
		return nil, err
	}
	return out.Keys, nil
}

// CreateAPIKey creates a new API key.
//
// Calls POST /v1/iam/api-keys. The returned APIKeyCreated.Secret is shown once — store it securely.
func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyCreated, error) {
	var out APIKeyCreated
	if err := c.do(ctx, "POST", "/v1/iam/api-keys", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAPIKey permanently removes an API key.
//
// Calls DELETE /v1/iam/api-keys/{keyID}.
func (c *Client) DeleteAPIKey(ctx context.Context, keyID string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/v1/iam/api-keys/%s", keyID), nil, nil)
}

// CheckPermission returns whether the currently authenticated principal has the given permission.
//
// Calls GET /v1/iam/check?permission={permission}.
// Useful as middleware in Zist services: if !ok { return 403 }.
func (c *Client) CheckPermission(ctx context.Context, permission string) (bool, error) {
	var out struct {
		Allowed bool `json:"allowed"`
	}
	path := fmt.Sprintf("/v1/iam/check?permission=%s", permission)
	if err := c.do(ctx, "GET", path, nil, &out); err != nil {
		return false, err
	}
	return out.Allowed, nil
}
