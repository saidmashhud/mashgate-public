package mashgate

import (
	"context"
	"fmt"
)

// ────────────────────────────────────────────────────────────────────────────
// BillingClient — platform subscription, plan, invoice, payment method.
// Mirrors BillingService in contracts/proto/v1/billing.proto (29 RPCs).
//
// Usage:
//
//	sub, err := client.Billing.GetSubscription(ctx)
//	plans, err := client.Billing.ListPlans(ctx)
// ────────────────────────────────────────────────────────────────────────────

// BillingClient wraps billing REST endpoints behind the main mashgate.Client.
type BillingClient struct {
	c *Client
}

// ListPlans returns all platform plans available to the current tenant.
func (b *BillingClient) ListPlans(ctx context.Context) ([]*BillingPlan, error) {
	var out struct {
		Plans []*BillingPlan `json:"plans"`
	}
	if err := b.c.do(ctx, "GET", "/v1/billing/plans", nil, &out); err != nil {
		return nil, err
	}
	return out.Plans, nil
}

// GetPlan retrieves a platform plan by ID.
func (b *BillingClient) GetPlan(ctx context.Context, planID string) (*BillingPlan, error) {
	var out BillingPlan
	if err := b.c.do(ctx, "GET", "/v1/billing/plans/"+planID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSubscription returns the current tenant's active subscription.
func (b *BillingClient) GetSubscription(ctx context.Context) (*BillingSubscription, error) {
	var out BillingSubscription
	if err := b.c.do(ctx, "GET", "/v1/billing/subscription", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ChangePlan switches the tenant's subscription to a new plan.
func (b *BillingClient) ChangePlan(ctx context.Context, req ChangePlanRequest) (*BillingSubscription, error) {
	var out BillingSubscription
	if err := b.c.do(ctx, "POST", "/v1/billing/subscription/change-plan", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelPlan cancels the current subscription (effective at period end unless
// CancelImmediately is set).
func (b *BillingClient) CancelPlan(ctx context.Context, req CancelPlanRequest) (*BillingSubscription, error) {
	var out BillingSubscription
	if err := b.c.do(ctx, "POST", "/v1/billing/subscription/cancel", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PreviewPlanChange computes prorated charges + effective dates for a hypothetical
// plan switch without applying it. Useful for confirmation UIs.
func (b *BillingClient) PreviewPlanChange(ctx context.Context, req PreviewPlanChangeRequest) (*PreviewPlanChangeResponse, error) {
	var out PreviewPlanChangeResponse
	if err := b.c.do(ctx, "POST", "/v1/billing/subscription/preview-change", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPaymentMethods returns all billing payment methods registered for the tenant.
func (b *BillingClient) ListPaymentMethods(ctx context.Context) ([]*BillingPaymentMethod, error) {
	var out struct {
		Methods []*BillingPaymentMethod `json:"methods"`
	}
	if err := b.c.do(ctx, "GET", "/v1/billing/payment-methods", nil, &out); err != nil {
		return nil, err
	}
	return out.Methods, nil
}

// AddPaymentMethod registers a new payment method for the tenant.
func (b *BillingClient) AddPaymentMethod(ctx context.Context, req AddBillingPaymentMethodRequest) (*BillingPaymentMethod, error) {
	var out BillingPaymentMethod
	if err := b.c.do(ctx, "POST", "/v1/billing/payment-methods", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetDefaultPaymentMethod marks a method as the auto-billing default.
func (b *BillingClient) SetDefaultPaymentMethod(ctx context.Context, methodID string) (*BillingPaymentMethod, error) {
	var out BillingPaymentMethod
	path := fmt.Sprintf("/v1/billing/payment-methods/%s/default", methodID)
	if err := b.c.do(ctx, "POST", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RemovePaymentMethod deletes a billing payment method.
func (b *BillingClient) RemovePaymentMethod(ctx context.Context, methodID string) error {
	return b.c.do(ctx, "DELETE", "/v1/billing/payment-methods/"+methodID, nil, nil)
}

// ListInvoices returns all billing invoices for the tenant.
func (b *BillingClient) ListInvoices(ctx context.Context) ([]*BillingInvoice, error) {
	var out struct {
		Invoices []*BillingInvoice `json:"invoices"`
	}
	if err := b.c.do(ctx, "GET", "/v1/billing/invoices", nil, &out); err != nil {
		return nil, err
	}
	return out.Invoices, nil
}

// GetInvoice retrieves a single invoice by ID.
func (b *BillingClient) GetInvoice(ctx context.Context, invoiceID string) (*BillingInvoice, error) {
	var out BillingInvoice
	if err := b.c.do(ctx, "GET", "/v1/billing/invoices/"+invoiceID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PayInvoice triggers an immediate payment attempt on an open invoice.
func (b *BillingClient) PayInvoice(ctx context.Context, invoiceID string) (*BillingInvoice, error) {
	var out BillingInvoice
	if err := b.c.do(ctx, "POST", "/v1/billing/invoices/"+invoiceID+"/pay", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCreditBalance returns the tenant's current credit balance.
func (b *BillingClient) GetCreditBalance(ctx context.Context) (*CreditBalance, error) {
	var out CreditBalance
	if err := b.c.do(ctx, "GET", "/v1/billing/credit-balance", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RedeemPromoCode applies a promo code to the tenant.
func (b *BillingClient) RedeemPromoCode(ctx context.Context, code string) (*RedeemPromoCodeResponse, error) {
	var out RedeemPromoCodeResponse
	req := struct {
		Code string `json:"code"`
	}{Code: code}
	if err := b.c.do(ctx, "POST", "/v1/billing/promo/redeem", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
