package mashgate

import "context"

// Login authenticates a user and returns a JWT token pair.
//
// Calls POST /v1/auth/login.
//
//	pair, err := client.Login(ctx, "user@example.com", "secret")
//	// store pair.AccessToken in Authorization header for subsequent requests
func (c *Client) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	var out TokenPair
	if err := c.do(ctx, "POST", "/v1/auth/login", LoginRequest{
		Email:    email,
		Password: password,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RefreshToken exchanges a refresh token for a new token pair.
//
// Calls POST /v1/auth/refresh.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	var out TokenPair
	body := struct {
		RefreshToken string `json:"refreshToken"`
	}{RefreshToken: refreshToken}
	if err := c.do(ctx, "POST", "/v1/auth/refresh", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Logout invalidates the given refresh token server-side.
//
// Calls POST /v1/auth/logout.
func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	body := struct {
		RefreshToken string `json:"refreshToken"`
	}{RefreshToken: refreshToken}
	return c.do(ctx, "POST", "/v1/auth/logout", body, nil)
}
