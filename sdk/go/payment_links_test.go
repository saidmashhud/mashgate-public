package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaymentLinksUseStructuralMerchantOwnership(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch requests {
		case 1:
			if r.Method != http.MethodPost || r.URL.Path != "/v1/payment-links" {
				t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["merchantId"] != "merchant-1" {
				t.Fatalf("merchantId = %#v", body["merchantId"])
			}
			if body["description"] != "invoice 42" {
				t.Fatalf("description was mutated: %#v", body["description"])
			}
			_, _ = w.Write([]byte(`{"id":"link-1","tenantId":"tenant-1","merchantId":"merchant-1","description":"invoice 42"}`))
		case 2:
			if r.URL.Query().Get("tenantId") != "tenant-1" || r.URL.Query().Get("merchantId") != "merchant-1" {
				t.Fatalf("unexpected query %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected extra request")
		}
	}))
	defer server.Close()

	client := NewClient("server-key", WithBaseURL(server.URL))
	link, err := client.PaymentLinks.Create(context.Background(), CreatePaymentLinkRequest{
		TenantID: "tenant-1", MerchantID: "merchant-1", Amount: 42,
		Currency: "USDT", Description: "invoice 42",
	})
	if err != nil {
		t.Fatal(err)
	}
	if link.MerchantID != "merchant-1" || link.Description != "invoice 42" {
		t.Fatalf("unexpected link: %#v", link)
	}
	if _, err := client.PaymentLinks.ListForMerchant(context.Background(), "tenant-1", "merchant-1"); err != nil {
		t.Fatal(err)
	}
}
