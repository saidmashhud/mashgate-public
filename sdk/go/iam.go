package mashgate

import (
	"context"
	"fmt"
)

type listAPIKeysResponse struct {
	Keys []*APIKey `json:"keys"`
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
