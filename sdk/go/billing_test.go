package mashgate

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// newBillingClient stands up a recording httptest.Server (via newTestServerFn)
// and returns the *BillingClient reached through the documented public surface,
// client.Billing — exercising both the wiring in initClients and the real
// *Client / do() HTTP path. (If the Billing field is ever dropped from the
// Client struct or initClients again, this test stops compiling.)
func newBillingClient(t *testing.T, fn http.HandlerFunc) *BillingClient {
	t.Helper()
	_, client := newTestServerFn(t, fn)
	return client.Billing
}

// TestBillingClient_pathsMatchProtoContract pins every BillingClient REST path +
// HTTP method to the proto contract, and for the POST bodies that carry data it
// also pins the request payload. It stands up an httptest.Server that records the
// method, path, and raw request body the SDK actually emits, points a real Client
// at it, and returns minimal valid JSON so each call decodes without error.
//
// If a path drifts (the regression that was just fixed for
// change/preview/credits/credits/redeem), or RedeemPromoCode stops wrapping its
// argument as {"code": ...}, the recorded value stops matching the expectation and
// the test FAILS.
func TestBillingClient_pathsMatchProtoContract(t *testing.T) {
	cases := []struct {
		name       string
		wantMethod string
		wantPath   string
		// wantBody, when non-nil, is the exact JSON object the SDK must send. It is
		// compared after round-tripping both sides through encoding/json so field
		// order and whitespace are irrelevant.
		wantBody any
		// body is the JSON object the fake server returns for this call. It must be
		// shaped so the corresponding method can decode it without error.
		body any
		// call invokes the method under test and, on success, asserts on the decoded
		// response so a broken decode path / wrong envelope key is caught.
		call func(t *testing.T, b *BillingClient)
	}{
		{
			name:       "ListPlans",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/billing/plans",
			body:       map[string]any{"plans": []*BillingPlan{{ID: "plan_pro"}}},
			call: func(t *testing.T, b *BillingClient) {
				plans, err := b.ListPlans(context.Background())
				if err != nil {
					t.Fatalf("ListPlans error: %v", err)
				}
				if len(plans) != 1 || plans[0].ID != "plan_pro" {
					t.Errorf("ListPlans decoded %+v, want one plan with ID plan_pro", plans)
				}
			},
		},
		{
			name:       "GetPlan",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/billing/plans/plan_pro",
			body:       BillingPlan{ID: "plan_pro"},
			call: func(t *testing.T, b *BillingClient) {
				plan, err := b.GetPlan(context.Background(), "plan_pro")
				if err != nil {
					t.Fatalf("GetPlan error: %v", err)
				}
				if plan.ID != "plan_pro" {
					t.Errorf("GetPlan decoded ID %q, want plan_pro", plan.ID)
				}
			},
		},
		{
			name:       "GetSubscription",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/billing/subscription",
			body:       BillingSubscription{ID: "sub_1"},
			call: func(t *testing.T, b *BillingClient) {
				sub, err := b.GetSubscription(context.Background())
				if err != nil {
					t.Fatalf("GetSubscription error: %v", err)
				}
				if sub.ID != "sub_1" {
					t.Errorf("GetSubscription decoded ID %q, want sub_1", sub.ID)
				}
			},
		},
		{
			// Regression-critical: must be /v1/billing/subscription/change, NOT
			// /v1/billing/subscription/change-plan or /v1/billing/plan/change.
			name:       "ChangePlan",
			wantMethod: http.MethodPost,
			wantPath:   "/v1/billing/subscription/change",
			wantBody:   map[string]any{"newPlanId": "plan_pro", "prorate": true},
			body:       BillingSubscription{ID: "sub_1"},
			call: func(t *testing.T, b *BillingClient) {
				sub, err := b.ChangePlan(context.Background(), ChangePlanRequest{NewPlanID: "plan_pro", Prorate: true})
				if err != nil {
					t.Fatalf("ChangePlan error: %v", err)
				}
				if sub.ID != "sub_1" {
					t.Errorf("ChangePlan decoded ID %q, want sub_1", sub.ID)
				}
			},
		},
		{
			name:       "CancelPlan",
			wantMethod: http.MethodPost,
			wantPath:   "/v1/billing/subscription/cancel",
			wantBody:   map[string]any{"reason": "too_expensive", "cancelImmediately": true},
			body:       BillingSubscription{ID: "sub_1"},
			call: func(t *testing.T, b *BillingClient) {
				sub, err := b.CancelPlan(context.Background(), CancelPlanRequest{Reason: "too_expensive", CancelImmediately: true})
				if err != nil {
					t.Fatalf("CancelPlan error: %v", err)
				}
				if sub.ID != "sub_1" {
					t.Errorf("CancelPlan decoded ID %q, want sub_1", sub.ID)
				}
			},
		},
		{
			// Regression-critical: must be /v1/billing/subscription/preview.
			name:       "PreviewPlanChange",
			wantMethod: http.MethodPost,
			wantPath:   "/v1/billing/subscription/preview",
			wantBody:   map[string]any{"newPlanId": "plan_pro"},
			body:       PreviewPlanChangeResponse{ProrationCents: 1234},
			call: func(t *testing.T, b *BillingClient) {
				prev, err := b.PreviewPlanChange(context.Background(), PreviewPlanChangeRequest{NewPlanID: "plan_pro"})
				if err != nil {
					t.Fatalf("PreviewPlanChange error: %v", err)
				}
				if prev.ProrationCents != 1234 {
					t.Errorf("PreviewPlanChange decoded ProrationCents %d, want 1234", prev.ProrationCents)
				}
			},
		},
		{
			name:       "ListPaymentMethods",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/billing/payment-methods",
			body:       map[string]any{"methods": []*BillingPaymentMethod{{ID: "pm_1"}}},
			call: func(t *testing.T, b *BillingClient) {
				methods, err := b.ListPaymentMethods(context.Background())
				if err != nil {
					t.Fatalf("ListPaymentMethods error: %v", err)
				}
				if len(methods) != 1 || methods[0].ID != "pm_1" {
					t.Errorf("ListPaymentMethods decoded %+v, want one method with ID pm_1", methods)
				}
			},
		},
		{
			name:       "AddPaymentMethod",
			wantMethod: http.MethodPost,
			wantPath:   "/v1/billing/payment-methods",
			wantBody:   map[string]any{"type": "card"},
			body:       BillingPaymentMethod{ID: "pm_1"},
			call: func(t *testing.T, b *BillingClient) {
				pm, err := b.AddPaymentMethod(context.Background(), AddBillingPaymentMethodRequest{Type: "card"})
				if err != nil {
					t.Fatalf("AddPaymentMethod error: %v", err)
				}
				if pm.ID != "pm_1" {
					t.Errorf("AddPaymentMethod decoded ID %q, want pm_1", pm.ID)
				}
			},
		},
		{
			name:       "SetDefaultPaymentMethod",
			wantMethod: http.MethodPost,
			wantPath:   "/v1/billing/payment-methods/pm_1/default",
			body:       BillingPaymentMethod{ID: "pm_1"},
			call: func(t *testing.T, b *BillingClient) {
				pm, err := b.SetDefaultPaymentMethod(context.Background(), "pm_1")
				if err != nil {
					t.Fatalf("SetDefaultPaymentMethod error: %v", err)
				}
				if pm.ID != "pm_1" {
					t.Errorf("SetDefaultPaymentMethod decoded ID %q, want pm_1", pm.ID)
				}
			},
		},
		{
			name:       "RemovePaymentMethod",
			wantMethod: http.MethodDelete,
			wantPath:   "/v1/billing/payment-methods/pm_1",
			body:       nil,
			call: func(t *testing.T, b *BillingClient) {
				if err := b.RemovePaymentMethod(context.Background(), "pm_1"); err != nil {
					t.Fatalf("RemovePaymentMethod error: %v", err)
				}
			},
		},
		{
			name:       "ListInvoices",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/billing/invoices",
			body:       map[string]any{"invoices": []*BillingInvoice{{ID: "inv_1"}}},
			call: func(t *testing.T, b *BillingClient) {
				invoices, err := b.ListInvoices(context.Background())
				if err != nil {
					t.Fatalf("ListInvoices error: %v", err)
				}
				if len(invoices) != 1 || invoices[0].ID != "inv_1" {
					t.Errorf("ListInvoices decoded %+v, want one invoice with ID inv_1", invoices)
				}
			},
		},
		{
			name:       "GetInvoice",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/billing/invoices/inv_1",
			body:       BillingInvoice{ID: "inv_1"},
			call: func(t *testing.T, b *BillingClient) {
				inv, err := b.GetInvoice(context.Background(), "inv_1")
				if err != nil {
					t.Fatalf("GetInvoice error: %v", err)
				}
				if inv.ID != "inv_1" {
					t.Errorf("GetInvoice decoded ID %q, want inv_1", inv.ID)
				}
			},
		},
		{
			name:       "PayInvoice",
			wantMethod: http.MethodPost,
			wantPath:   "/v1/billing/invoices/inv_1/pay",
			body:       BillingInvoice{ID: "inv_1"},
			call: func(t *testing.T, b *BillingClient) {
				inv, err := b.PayInvoice(context.Background(), "inv_1")
				if err != nil {
					t.Fatalf("PayInvoice error: %v", err)
				}
				if inv.ID != "inv_1" {
					t.Errorf("PayInvoice decoded ID %q, want inv_1", inv.ID)
				}
			},
		},
		{
			// Regression-critical: must be /v1/billing/credits.
			name:       "GetCreditBalance",
			wantMethod: http.MethodGet,
			wantPath:   "/v1/billing/credits",
			body:       CreditBalance{AmountCents: 500, Currency: "UZS"},
			call: func(t *testing.T, b *BillingClient) {
				bal, err := b.GetCreditBalance(context.Background())
				if err != nil {
					t.Fatalf("GetCreditBalance error: %v", err)
				}
				if bal.AmountCents != 500 || bal.Currency != "UZS" {
					t.Errorf("GetCreditBalance decoded %+v, want {500 UZS}", bal)
				}
			},
		},
		{
			// Regression-critical: path must be /v1/billing/credits/redeem AND the
			// raw promo string must be wrapped as {"code": ...} (billing.go:157-160).
			name:       "RedeemPromoCode",
			wantMethod: http.MethodPost,
			wantPath:   "/v1/billing/credits/redeem",
			wantBody:   map[string]any{"code": "PROMO50"},
			body:       RedeemPromoCodeResponse{Applied: true},
			call: func(t *testing.T, b *BillingClient) {
				resp, err := b.RedeemPromoCode(context.Background(), "PROMO50")
				if err != nil {
					t.Fatalf("RedeemPromoCode error: %v", err)
				}
				if !resp.Applied {
					t.Errorf("RedeemPromoCode decoded Applied=false, want true")
				}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody []byte

			b := newBillingClient(t, func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if tc.body != nil {
					_ = json.NewEncoder(w).Encode(tc.body)
				}
			})

			tc.call(t, b)

			if gotMethod != tc.wantMethod {
				t.Errorf("%s: method = %q, want %q", tc.name, gotMethod, tc.wantMethod)
			}
			if gotPath != tc.wantPath {
				t.Errorf("%s: path = %q, want %q", tc.name, gotPath, tc.wantPath)
			}
			if tc.wantBody != nil {
				assertJSONBodyEquals(t, tc.name, gotBody, tc.wantBody)
			}
		})
	}
}

// assertJSONBodyEquals fails the test unless the raw request body decodes to the
// same JSON value as want. Both sides are normalised through encoding/json so the
// comparison is insensitive to key order and whitespace, but sensitive to the set
// of keys and their values — so dropping/renaming a field (e.g. RedeemPromoCode's
// "code") is caught.
func assertJSONBodyEquals(t *testing.T, name string, got []byte, want any) {
	t.Helper()

	wantBytes, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("%s: marshal want body: %v", name, err)
	}

	var gotVal, wantVal any
	if err := json.Unmarshal(bytes.TrimSpace(got), &gotVal); err != nil {
		t.Fatalf("%s: request body is not valid JSON (%q): %v", name, string(got), err)
	}
	if err := json.Unmarshal(wantBytes, &wantVal); err != nil {
		t.Fatalf("%s: unmarshal want body: %v", name, err)
	}

	gotNorm, _ := json.Marshal(gotVal)
	wantNorm, _ := json.Marshal(wantVal)
	if string(gotNorm) != string(wantNorm) {
		t.Errorf("%s: request body = %s, want %s", name, gotNorm, wantNorm)
	}
}
