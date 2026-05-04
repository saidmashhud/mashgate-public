package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRefreshTokenUsesCamelCaseRequest(t *testing.T) {
	var captured map[string]any

	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"accessToken":  "access_new",
			"refreshToken": "refresh_new",
		})
	})

	pair, err := client.RefreshToken(context.Background(), "refresh_old")
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}
	if captured["refreshToken"] != "refresh_old" {
		t.Fatalf("expected refreshToken body field, got %#v", captured)
	}
	if _, ok := captured["refresh_token"]; ok {
		t.Fatalf("did not expect refresh_token body field: %#v", captured)
	}
	if pair.AccessToken != "access_new" || pair.RefreshToken != "refresh_new" {
		t.Fatalf("unexpected token pair: %#v", pair)
	}
}

func TestLoginParsesUserAndExpires(t *testing.T) {
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/login" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accessToken":  "a",
			"refreshToken": "r",
			"expiresAt":    1777839438,
			"user": map[string]any{
				"userId":   "u-1",
				"email":    "x@y.z",
				"fullName": "X",
				"tenantId": "t-1",
				"roles":    []string{"admin"},
			},
		})
	})
	pair, err := client.Login(context.Background(), "x@y.z", "p")
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if pair.User == nil || pair.User.UserID != "u-1" {
		t.Fatalf("user not parsed: %#v", pair.User)
	}
	if pair.ExpiresAt != 1777839438 {
		t.Fatalf("expiresAt: %d", pair.ExpiresAt)
	}
}

func TestRegisterPostsSnakeCaseFields(t *testing.T) {
	var captured map[string]any
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"userId":    "new-user",
			"email":     "x@y.z",
			"tenantId":  "t-1",
			"createdAt": 123,
		})
	})
	out, err := client.Register(context.Background(), RegisterRequest{
		Email: "x@y.z", Password: "12345678901", FullName: "X", TenantID: "t-1", Role: "merchant",
	})
	if err != nil {
		t.Fatalf("Register err: %v", err)
	}
	if captured["full_name"] != "X" || captured["tenant_id"] != "t-1" || captured["role"] != "merchant" {
		t.Fatalf("expected snake_case body, got %#v", captured)
	}
	if out.UserID != "new-user" {
		t.Fatalf("unexpected userId: %#v", out)
	}
}

func TestSendOtpPostsExpectedBody(t *testing.T) {
	var captured map[string]any
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/otp/send" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	if err := client.SendOtp(context.Background(), SendOtpRequest{Phone: "+992900000000", Purpose: "login"}); err != nil {
		t.Fatalf("SendOtp err: %v", err)
	}
	if captured["phone"] != "+992900000000" || captured["purpose"] != "login" {
		t.Fatalf("body: %#v", captured)
	}
}

func TestVerifyOtpReturnsValid(t *testing.T) {
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/otp/verify" {
			t.Fatalf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"valid": true})
	})
	ok, err := client.VerifyOtp(context.Background(), VerifyOtpRequest{
		UserID: "u-1", Code: "123456", Purpose: "login",
	})
	if err != nil {
		t.Fatalf("VerifyOtp err: %v", err)
	}
	if !ok {
		t.Fatalf("expected valid=true")
	}
}
