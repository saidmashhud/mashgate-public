package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListTenantsNoOptions(t *testing.T) {
	var capturedPath string

	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tenants": []map[string]any{
				{
					"tenantId": "11111111-1111-1111-1111-111111111111",
					"code":     "acme-corp",
					"name":     "ACME Corporation",
					"mode":     "live",
					"status":   "active",
				},
				{
					"tenantId": "22222222-2222-2222-2222-222222222222",
					"code":     "demo",
					"name":     "Demo Tenant",
					"mode":     "sandbox",
					"status":   "active",
				},
			},
			"totalCount": 2,
		})
	})

	tenants, err := client.ListTenants(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTenants returned error: %v", err)
	}
	if capturedPath != "/v1/iam/tenants" {
		t.Fatalf("expected path /v1/iam/tenants, got %q", capturedPath)
	}
	if len(tenants) != 2 {
		t.Fatalf("expected 2 tenants, got %d", len(tenants))
	}
	if tenants[0].Code != "acme-corp" || tenants[0].Mode != "live" {
		t.Fatalf("unexpected tenant[0]: %+v", tenants[0])
	}
	if tenants[1].TenantID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("unexpected tenant[1] id: %q", tenants[1].TenantID)
	}
}

func TestListTenantsWithOptions(t *testing.T) {
	var capturedPath string

	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tenants":    []map[string]any{},
			"totalCount": 0,
		})
	})

	_, err := client.ListTenants(context.Background(), &ListTenantsOptions{
		Status:    "active",
		Search:    "acme",
		Page:      2,
		PageSize:  50,
		SortBy:    "name",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("ListTenants returned error: %v", err)
	}
	expected := "/v1/iam/tenants?status=active&search=acme&page=2&pageSize=50&sortBy=name&sortOrder=asc"
	if capturedPath != expected {
		t.Fatalf("expected path %q, got %q", expected, capturedPath)
	}
}

func TestListTenantsEmpty(t *testing.T) {
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
	})

	tenants, err := client.ListTenants(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTenants returned error: %v", err)
	}
	if len(tenants) != 0 {
		t.Fatalf("expected empty slice, got %d tenants", len(tenants))
	}
}
