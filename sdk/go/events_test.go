package mashgate_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
)

// ────────────────────────────────────────────────────────────────────────────
// helpers
// ────────────────────────────────────────────────────────────────────────────

func endpointJSON(id, url string) string {
	return `{"id":"` + id + `","url":"` + url + `","eventTypes":["payment.captured"],"status":"active","createdAt":0,"updatedAt":0}`
}

func newHLServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *mashgate.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := mashgate.New(srv.URL, "test-key").
		WithEvents(mashgate.InternalHooklineConfig(srv.URL, "hl-key"))
	return srv, client
}

func newMGServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *mashgate.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := mashgate.New(srv.URL, "mg-key").
		WithEvents(mashgate.EventsConfig{
			MashgateEventsURL: srv.URL,
		})
	return srv, client
}

// ────────────────────────────────────────────────────────────────────────────
// CreateEndpoint — hookline mode
// ────────────────────────────────────────────────────────────────────────────

func TestCreateEndpoint_HooklineMode(t *testing.T) {
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/endpoints" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer hl-key" {
			t.Errorf("wrong auth header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(endpointJSON("ep_001", "https://example.com/hook")))
	})

	ep, err := client.Events.CreateEndpoint(context.Background(), mashgate.CreateEndpointRequest{
		URL:        "https://example.com/hook",
		EventTypes: []string{"payment.captured"},
	})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}
	if ep.ID != "ep_001" {
		t.Errorf("got ID %q, want ep_001", ep.ID)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// CreateEndpoint — mashgate mode
// ────────────────────────────────────────────────────────────────────────────

func TestCreateEndpoint_MashgateMode(t *testing.T) {
	_, client := newMGServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/events/endpoints" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "mg-key" {
			t.Errorf("wrong X-API-Key header: %s", r.Header.Get("X-API-Key"))
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"endpoint": {
				"endpoint_id":"ep_mg1",
				"url":"https://mg.example.com/hook",
				"description":"",
				"event_types":["checkout.completed"],
				"status":"active",
				"signing_secret":"",
				"created_at":0,
				"updated_at":0
			}
		}`))
	})

	ep, err := client.Events.CreateEndpoint(context.Background(), mashgate.CreateEndpointRequest{
		URL:        "https://mg.example.com/hook",
		EventTypes: []string{"checkout.completed"},
	})
	if err != nil {
		t.Fatalf("CreateEndpoint (mashgate): %v", err)
	}
	if ep.ID != "ep_mg1" {
		t.Errorf("got ID %q, want ep_mg1", ep.ID)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ListEndpoints — hookline mode
// ────────────────────────────────────────────────────────────────────────────

func TestListEndpoints_HooklineMode(t *testing.T) {
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/endpoints" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"endpoints":[` + endpointJSON("ep_A", "https://a.io") + `,` + endpointJSON("ep_B", "https://b.io") + `]}`))
	})

	eps, err := client.Events.ListEndpoints(context.Background())
	if err != nil {
		t.Fatalf("ListEndpoints: %v", err)
	}
	if len(eps) != 2 {
		t.Errorf("got %d endpoints, want 2", len(eps))
	}
}

// ────────────────────────────────────────────────────────────────────────────
// 404 → EndpointNotFoundError
// ────────────────────────────────────────────────────────────────────────────

func TestDeleteEndpoint_NotFound(t *testing.T) {
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"id":"ep_missing","message":"endpoint not found"}`))
	})

	err := client.Events.DeleteEndpoint(context.Background(), "ep_missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	nfe, ok := err.(*mashgate.EndpointNotFoundError)
	if !ok {
		t.Fatalf("expected *EndpointNotFoundError, got %T: %v", err, err)
	}
	if nfe.ID != "ep_missing" {
		t.Errorf("got ID %q, want ep_missing", nfe.ID)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// 409 → SubscriptionConflictError
// ────────────────────────────────────────────────────────────────────────────

func TestCreateSubscription_Conflict(t *testing.T) {
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"message":"subscription already exists"}`))
	})

	_, err := client.Events.CreateSubscription(context.Background(), "ep_001", []string{"payment.captured"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*mashgate.SubscriptionConflictError); !ok {
		t.Fatalf("expected *SubscriptionConflictError, got %T: %v", err, err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// 429 → QuotaExceededError (and retry)
// ────────────────────────────────────────────────────────────────────────────

func TestRetry_On429_ThenSuccess(t *testing.T) {
	attempts := 0
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message":"rate limited","retry_after":1}`))
			return
		}
		// Third attempt succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(endpointJSON("ep_retry", "https://retry.io")))
	})

	ep, err := client.Events.CreateEndpoint(context.Background(), mashgate.CreateEndpointRequest{
		URL:        "https://retry.io",
		EventTypes: []string{"payment.captured"},
	})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if ep.ID != "ep_retry" {
		t.Errorf("got ID %q", ep.ID)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_On429_ExhaustsRetries(t *testing.T) {
	attempts := 0
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"always rate limited","retry_after":1}`))
	})

	_, err := client.Events.CreateEndpoint(context.Background(), mashgate.CreateEndpointRequest{
		URL:        "https://never.io",
		EventTypes: []string{"payment.captured"},
	})
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	if _, ok := err.(*mashgate.QuotaExceededError); !ok {
		t.Fatalf("expected *QuotaExceededError, got %T", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Trace propagation
// ────────────────────────────────────────────────────────────────────────────

func TestTraceparent_Propagated(t *testing.T) {
	const wantTp = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Traceparent")
		if got != wantTp {
			t.Errorf("Traceparent header: got %q, want %q", got, wantTp)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"endpoints":[]}`))
	})

	ctx := mashgate.WithTraceparent(context.Background(), wantTp)
	_, err := client.Events.ListEndpoints(ctx)
	if err != nil {
		t.Fatalf("ListEndpoints: %v", err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Idempotency — creating the same endpoint twice returns the existing one
// ────────────────────────────────────────────────────────────────────────────

func TestCreateEndpoint_Idempotent(t *testing.T) {
	calls := 0
	stored := endpointJSON("ep_idem", "https://idem.io")

	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "ep_idem") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(stored))
			return
		}
		// First POST → 201, second POST → 409 with id
		if calls == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(stored))
			return
		}
		// Simulate server returning 409; client should surface SubscriptionConflictError or similar
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		b, _ := json.Marshal(map[string]string{"message": "endpoint already exists", "id": "ep_idem"})
		w.Write(b)
	})

	req := mashgate.CreateEndpointRequest{URL: "https://idem.io", EventTypes: []string{"payment.captured"}}

	ep1, err := client.Events.CreateEndpoint(context.Background(), req)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	if ep1.ID != "ep_idem" {
		t.Errorf("first create: got ID %q", ep1.ID)
	}

	// Second create → conflict
	_, err = client.Events.CreateEndpoint(context.Background(), req)
	if err == nil {
		t.Fatal("second create: expected conflict error, got nil")
	}
	if _, ok := err.(*mashgate.SubscriptionConflictError); !ok {
		t.Fatalf("second create: expected *SubscriptionConflictError, got %T: %v", err, err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// RotateSecret
// ────────────────────────────────────────────────────────────────────────────

func TestRotateSecret(t *testing.T) {
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/endpoints/ep_001/rotate-secret" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ep_001","url":"https://example.com","signingSecret":"new-secret-xyz","eventTypes":[],"status":"active","createdAt":0,"updatedAt":0}`))
	})

	ep, err := client.Events.RotateSecret(context.Background(), "ep_001")
	if err != nil {
		t.Fatalf("RotateSecret: %v", err)
	}
	if ep.SigningSecret != "new-secret-xyz" {
		t.Errorf("signing secret: got %q, want new-secret-xyz", ep.SigningSecret)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// 503 → EventsServiceUnavailableError (retried)
// ────────────────────────────────────────────────────────────────────────────

func TestRetry_On503_ThenSuccess(t *testing.T) {
	attempts := 0
	_, client := newHLServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"message":"temporarily unavailable"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"endpoints":[]}`))
	})

	_, err := client.Events.ListEndpoints(context.Background())
	if err != nil {
		t.Fatalf("ListEndpoints after 503 retry: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
