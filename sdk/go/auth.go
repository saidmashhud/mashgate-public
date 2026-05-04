package mashgate

import "context"

// Login authenticates a user and returns a JWT token pair (with optional
// User identity claims when present in the response).
//
// Calls POST /v1/auth/login.
//
//	pair, err := client.Login(ctx, "user@example.com", "secret")
//	// pair.User may be nil if upstream omits it; pair.AccessToken always set.
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

// Register creates a new user via Mashgate auth-service. Returns the user
// record only — no tokens. Caller chains Login to obtain a session.
//
// Calls POST /v1/auth/register.
//
//	user, err := client.Register(ctx, mashgate.RegisterRequest{
//	    Email: "merchant@example.com", Password: "...",
//	    FullName: "Merchant", TenantID: tenantID, Role: "merchant",
//	})
//	if err == nil {
//	    pair, _ := client.Login(ctx, user.Email, password)
//	    // ...
//	}
//
// See SECURITY NOTE in models.go on RegisterRequest.Role.
func (c *Client) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	var out RegisterResponse
	if err := c.do(ctx, "POST", "/v1/auth/register", req, &out); err != nil {
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

// SendOtp triggers OTP delivery via Mashgate notify-service. Provide either
// user_id (existing user) or phone (registration / passwordless flow).
//
// Calls POST /v1/auth/otp/send.
//
// Purpose ∈ {"login", "password_reset", "phone_verify"}.
//
//	if err := client.SendOtp(ctx, mashgate.SendOtpRequest{
//	    Phone: "+992900000000", Purpose: "login",
//	}); err != nil { ... }
func (c *Client) SendOtp(ctx context.Context, req SendOtpRequest) error {
	return c.do(ctx, "POST", "/v1/auth/otp/send", req, nil)
}

// VerifyOtp validates a previously-sent OTP. Returns true if the code is
// valid + within validity window.
//
// Calls POST /v1/auth/otp/verify.
//
//	ok, err := client.VerifyOtp(ctx, mashgate.VerifyOtpRequest{
//	    UserID: userID, Code: "123456", Purpose: "login",
//	})
//	if err != nil || !ok { /* reject */ }
func (c *Client) VerifyOtp(ctx context.Context, req VerifyOtpRequest) (bool, error) {
	var out struct {
		Valid bool `json:"valid"`
	}
	if err := c.do(ctx, "POST", "/v1/auth/otp/verify", req, &out); err != nil {
		return false, err
	}
	return out.Valid, nil
}
