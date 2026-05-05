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

// ────────────────────────────────────────────────────────────────────────────
// Password lifecycle (v1.4.0+)
// ────────────────────────────────────────────────────────────────────────────

// ChangePasswordRequest — authenticated password change. User proves
// possession of the current password.
type ChangePasswordRequest struct {
	UserID          string `json:"user_id"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword updates the password for an already-authenticated user.
// Calls POST /v1/auth/password/change.
func (c *Client) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	return c.do(ctx, "POST", "/v1/auth/password/change", req, nil)
}

// ResetPasswordRequest — forgot-password completion. Caller already ran
// SendOtp(purpose="password_reset") and got back a 6-digit code.
type ResetPasswordRequest struct {
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// ResetPassword exchanges a valid OTP for a new password.
// Calls POST /v1/auth/password/reset.
func (c *Client) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	return c.do(ctx, "POST", "/v1/auth/password/reset", req, nil)
}

// ────────────────────────────────────────────────────────────────────────────
// Email verification (v1.4.0+)
// ────────────────────────────────────────────────────────────────────────────

// SendEmailVerificationRequest — trigger an email-verify OTP for a user.
type SendEmailVerificationRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"` // optional override
}

// SendEmailVerification sends an OTP to the user's email for verification.
// Calls POST /v1/auth/email/verify/send.
func (c *Client) SendEmailVerification(ctx context.Context, req SendEmailVerificationRequest) error {
	return c.do(ctx, "POST", "/v1/auth/email/verify/send", req, nil)
}

// ConfirmEmailVerificationRequest — completes the verification flow.
type ConfirmEmailVerificationRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// ConfirmEmailVerificationResult is the boolean outcome.
type ConfirmEmailVerificationResult struct {
	Success       bool `json:"success"`
	EmailVerified bool `json:"email_verified"`
}

// ConfirmEmailVerification finalizes email verification.
// Calls POST /v1/auth/email/verify/confirm.
func (c *Client) ConfirmEmailVerification(ctx context.Context, req ConfirmEmailVerificationRequest) (*ConfirmEmailVerificationResult, error) {
	var out ConfirmEmailVerificationResult
	if err := c.do(ctx, "POST", "/v1/auth/email/verify/confirm", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Profile mutations (v1.4.0+)
// ────────────────────────────────────────────────────────────────────────────

// UpdateUserPhoneRequest — change the user's phone, gated by an OTP that
// the caller already ran SendOtp(phone=NEW, purpose="phone_verify") for.
type UpdateUserPhoneRequest struct {
	UserID   string `json:"user_id"`
	NewPhone string `json:"new_phone"`
	Code     string `json:"code"`
}

// UpdateUserPhone applies a verified phone change.
// Calls POST /v1/auth/profile/phone.
func (c *Client) UpdateUserPhone(ctx context.Context, req UpdateUserPhoneRequest) error {
	return c.do(ctx, "POST", "/v1/auth/profile/phone", req, nil)
}

// UpdateUserEmailRequest — change the user's email, gated by an OTP that
// the caller already ran SendOtp(user_id, purpose="phone_verify") for the
// existing email; new_email is then set.
type UpdateUserEmailRequest struct {
	UserID   string `json:"user_id"`
	NewEmail string `json:"new_email"`
	Code     string `json:"code"`
}

// UpdateUserEmail applies a verified email change.
// Calls POST /v1/auth/profile/email.
func (c *Client) UpdateUserEmail(ctx context.Context, req UpdateUserEmailRequest) error {
	return c.do(ctx, "POST", "/v1/auth/profile/email", req, nil)
}

// ────────────────────────────────────────────────────────────────────────────
// Account erasure (v1.4.0+)
// ────────────────────────────────────────────────────────────────────────────

// DeleteAccountRequest — irreversible (soft-delete on auth-service side,
// downstream products consume `user.deleted` event).
type DeleteAccountRequest struct {
	UserID          string `json:"user_id"`
	CurrentPassword string `json:"current_password"`
}

// DeleteAccount soft-deletes the user from Mashgate auth.
// Calls DELETE /v1/auth/account.
func (c *Client) DeleteAccount(ctx context.Context, req DeleteAccountRequest) error {
	return c.do(ctx, "DELETE", "/v1/auth/account", req, nil)
}
