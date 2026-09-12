package mashgate

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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
//
// Ключ признака идёт в ПУТИ, а не в теле: контракт объявляет
// `post: "/v1/flags/{flag_key}/evaluate"`. До 12.09.2026 здесь стоял
// `/v1/flags/evaluate` — такого маршрута у платформы нет, и шлюз отвечал
// `403 REST route not in authorization map` (проверка прав fail-closed).
// Вызывающие при этом трактовали отказ как «признак выключен», поэтому
// механизм признаков не работал НИ РАЗУ и выглядел как настройка.
//
// UserID и Groups уходят в `context`: в контракте у запроса три поля —
// tenant_id, flag_key и map<string,string> context. Отдельных userId/groups
// там нет, и раньше они просто не доезжали.
func (f *FlagsClient) Evaluate(ctx context.Context, req EvaluateFlagRequest) (*FlagEvaluation, error) {
	if req.FlagKey == "" {
		return nil, fmt.Errorf("mashgate: flags.Evaluate: пустой ключ признака")
	}
	body := struct {
		TenantID string            `json:"tenantId"`
		FlagKey  string            `json:"flagKey"`
		Context  map[string]string `json:"context,omitempty"`
	}{TenantID: req.TenantID, FlagKey: req.FlagKey}
	if req.UserID != "" || len(req.Groups) > 0 {
		body.Context = map[string]string{}
		if req.UserID != "" {
			body.Context["userId"] = req.UserID
		}
		if len(req.Groups) > 0 {
			body.Context["groups"] = strings.Join(req.Groups, ",")
		}
	}
	path := fmt.Sprintf("/v1/flags/%s/evaluate", url.PathEscape(req.FlagKey))
	var out FlagEvaluation
	if err := f.c.do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
