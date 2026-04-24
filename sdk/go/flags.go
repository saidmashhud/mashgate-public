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

// FeatureFlag is a tenant-scoped feature flag.
type FeatureFlag struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	FlagKey      string    `json:"flagKey"`
	Enabled      bool      `json:"enabled"`
	RolloutPct   int       `json:"rolloutPct"`
	TargetUsers  []string  `json:"targetUsers"`
	TargetGroups []string  `json:"targetGroups"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// FlagEvaluation is the result of evaluating a flag for a user.
type FlagEvaluation struct {
	FlagKey string `json:"flagKey"`
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreateFlagRequest creates a new feature flag.
type CreateFlagRequest struct {
	TenantID     string   `json:"tenantId"`
	FlagKey      string   `json:"flagKey"`
	Enabled      bool     `json:"enabled"`
	RolloutPct   int      `json:"rolloutPct,omitempty"`
	TargetUsers  []string `json:"targetUsers,omitempty"`
	TargetGroups []string `json:"targetGroups,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// UpdateFlagRequest updates an existing feature flag.
type UpdateFlagRequest struct {
	Enabled      *bool    `json:"enabled,omitempty"`
	RolloutPct   *int     `json:"rolloutPct,omitempty"`
	TargetUsers  []string `json:"targetUsers,omitempty"`
	TargetGroups []string `json:"targetGroups,omitempty"`
	Description  *string  `json:"description,omitempty"`
}

// EvaluateFlagRequest evaluates a flag for a specific user.
type EvaluateFlagRequest struct {
	TenantID string   `json:"tenantId"`
	FlagKey  string   `json:"flagKey"`
	UserID   string   `json:"userId,omitempty"`
	Groups   []string `json:"groups,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// FlagsClient
// ────────────────────────────────────────────────────────────────────────────

// FlagsClient provides access to the flags-service REST API.
type FlagsClient struct {
	c *Client
}

// Create creates a new feature flag.
func (f *FlagsClient) Create(ctx context.Context, req CreateFlagRequest) (*FeatureFlag, error) {
	var out FeatureFlag
	if err := f.c.do(ctx, "POST", "/v1/flags", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all flags for a tenant.
func (f *FlagsClient) List(ctx context.Context, tenantID string) ([]*FeatureFlag, error) {
	path := fmt.Sprintf("/v1/flags?tenantId=%s", url.QueryEscape(tenantID))
	var out []*FeatureFlag
	if err := f.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns a flag by key.
func (f *FlagsClient) Get(ctx context.Context, flagKey, tenantID string) (*FeatureFlag, error) {
	path := fmt.Sprintf("/v1/flags/%s?tenantId=%s", url.PathEscape(flagKey), url.QueryEscape(tenantID))
	var out FeatureFlag
	if err := f.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update updates a flag.
func (f *FlagsClient) Update(ctx context.Context, flagKey, tenantID string, req UpdateFlagRequest) (*FeatureFlag, error) {
	path := fmt.Sprintf("/v1/flags/%s?tenantId=%s", url.PathEscape(flagKey), url.QueryEscape(tenantID))
	var out FeatureFlag
	if err := f.c.do(ctx, "PUT", path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Evaluate evaluates a flag for a user.
func (f *FlagsClient) Evaluate(ctx context.Context, req EvaluateFlagRequest) (*FlagEvaluation, error) {
	var out FlagEvaluation
	if err := f.c.do(ctx, "POST", "/v1/flags/evaluate", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
