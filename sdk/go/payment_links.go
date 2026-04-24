package mashgate

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Domain types
// ────────────────────────────────────────────────────────────────────────────

// PaymentLink is a shareable payment URL.
type PaymentLink struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenantId"`
	LinkID      string     `json:"linkId"`
	URL         string     `json:"url"`
	Amount      int64      `json:"amount"`
	Currency    string     `json:"currency"`
	Description string     `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Status      string     `json:"status"` // "active" | "paid" | "expired"
	PaymentID   string     `json:"paymentId,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreatePaymentLinkRequest creates a new payment link.
type CreatePaymentLinkRequest struct {
	TenantID    string     `json:"tenantId"`
	Amount      int64      `json:"amount"`
	Currency    string     `json:"currency"`
	Description string     `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// PaymentLinksClient
// ────────────────────────────────────────────────────────────────────────────

// PaymentLinksClient provides access to payment link endpoints on the checkout-service.
type PaymentLinksClient struct {
	c *Client
}

// Create creates a new payment link.
func (p *PaymentLinksClient) Create(ctx context.Context, req CreatePaymentLinkRequest) (*PaymentLink, error) {
	var out PaymentLink
	if err := p.c.do(ctx, "POST", "/v1/payment-links", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all payment links for a tenant.
func (p *PaymentLinksClient) List(ctx context.Context, tenantID string) ([]*PaymentLink, error) {
	path := fmt.Sprintf("/v1/payment-links?tenantId=%s", url.QueryEscape(tenantID))
	var out []*PaymentLink
	if err := p.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns a payment link by ID.
func (p *PaymentLinksClient) Get(ctx context.Context, id string) (*PaymentLink, error) {
	var out PaymentLink
	if err := p.c.do(ctx, "GET", fmt.Sprintf("/v1/payment-links/%s", url.PathEscape(id)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
