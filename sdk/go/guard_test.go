package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGuardUsesCanonicalRuleAndEvaluateRoutes(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("X-API-Key") != "psk_test" {
			t.Fatalf("missing API key")
		}
		switch r.URL.Path {
		case "/v1/guard/rules":
			var body struct {
				TenantID   string               `json:"tenantId"`
				Resource   string               `json:"resource"`
				Conditions []guardConditionWire `json:"conditions"`
				Priority   int                  `json:"priority"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.TenantID != "tenant-1" || body.Resource != "/api/auth/otp/verify" || body.Priority != 5 {
				t.Fatalf("unexpected create rule body: %+v", body)
			}
			if len(body.Conditions) != 1 || body.Conditions[0].Value != "POST" {
				t.Fatalf("method condition missing: %+v", body.Conditions)
			}
			writeTestResponse(t, w, map[string]any{
				"id": "rule-1", "tenantId": "tenant-1", "resource": body.Resource,
				"conditions": body.Conditions, "priority": body.Priority,
			})
		case "/v1/guard/evaluate":
			var body struct {
				TenantID string            `json:"tenantId"`
				Resource string            `json:"resource"`
				Action   string            `json:"action"`
				Context  map[string]string `json:"context"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Action != "POST" || body.Context["ip"] != "203.0.113.7" {
				t.Fatalf("method/ip not forwarded: %+v", body)
			}
			writeTestResponse(t, w, map[string]any{"decision": "ALLOW", "reason": "allowed"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := New(srv.URL, "psk_test")
	rule, err := client.Guard.UpsertRateLimit(context.Background(), UpsertRateLimitRequest{
		TenantID: "tenant-1", Path: "/api/auth/otp/verify", Method: "POST", RPM: 5,
	})
	if err != nil || rule.Method != "POST" || rule.RPM != 5 {
		t.Fatalf("upsert result=%+v err=%v", rule, err)
	}
	decision, err := client.Guard.Check(context.Background(), GuardCheckRequest{
		TenantID: "tenant-1", Path: "/api/auth/otp/verify", Method: "POST", IP: "203.0.113.7",
	})
	if err != nil || !decision.Allowed {
		t.Fatalf("check result=%+v err=%v", decision, err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}

func writeTestResponse(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatal(err)
	}
}
