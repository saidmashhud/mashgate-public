package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey int

const claimsKey contextKey = iota

// Claims holds the parsed identity from a validated mgID access token.
type Claims struct {
	Sub      string
	UserID   string
	TenantID string
	Email    string
	Roles    []string
	Scopes   []string
	Exp      int64
}

// HasScope reports whether the claims include the given scope.
func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// AuthMiddleware returns a middleware that validates the Bearer token in the
// Authorization header by calling GET {mgID}/v1/auth/validate.
// Anonymous requests (no token) pass through; handlers call FromContext to
// check whether a principal is present.
func (c *Client) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := c.validateToken(r.Context(), token)
			if err != nil || claims == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireScope returns a middleware that responds 403 Forbidden if the
// authenticated principal does not hold the given scope. It must be used
// after AuthMiddleware (or after the gateway has set X-User-Scopes).
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := FromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			if !claims.HasScope(scope) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"insufficient_scope","required":"` + scope + `"}`)) //nolint:errcheck
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// FromContext extracts Claims from ctx, returning (nil, false) for anonymous requests.
func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok && c != nil
}

// validateToken calls GET {mgID}/v1/auth/validate and parses the response.
func (c *Client) validateToken(ctx context.Context, token string) (*Claims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/auth/validate", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var payload struct {
		Valid     bool     `json:"valid"`
		UserID    string   `json:"userId"`
		TenantID  string   `json:"tenantId"`
		Email     string   `json:"email"`
		Name      string   `json:"name"`
		Roles     []string `json:"roles"`
		Scope     string   `json:"scope"`
		ExpiresAt int64    `json:"expiresAt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.Valid {
		return nil, nil
	}

	return &Claims{
		Sub:      payload.UserID,
		UserID:   payload.UserID,
		TenantID: payload.TenantID,
		Email:    payload.Email,
		Roles:    payload.Roles,
		Scopes:   strings.Fields(payload.Scope),
		Exp:      payload.ExpiresAt,
	}, nil
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return h[7:]
	}
	return ""
}
