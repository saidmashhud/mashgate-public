package mashgate

import (
	"context"
	"fmt"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// Domain types
// ────────────────────────────────────────────────────────────────────────────

// InvoiceLineItem is a single line on an invoice. Money fields are minor units
// (int64) encoded as JSON strings (proto3 int64 convention via the transcoder).
type InvoiceLineItem struct {
	Description string `json:"description"`
	Quantity    int32  `json:"quantity"`
	UnitPrice   int64  `json:"unitPrice,string"`
	Amount      int64  `json:"amount,string"`
}

// Invoice is a merchant invoice. Mirrors invoices.v1.Invoice (proto): number,
// computed subtotal/tax/total (int64 minor units as JSON strings), status enum
// name (e.g. "OPEN"). Timestamps are RFC3339 strings.
type Invoice struct {
	ID         string            `json:"id"`
	TenantID   string            `json:"tenantId"`
	Number     string            `json:"number"`
	CustomerID string            `json:"customerId"`
	Status     string            `json:"status"` // DRAFT | OPEN | PAID | VOID
	LineItems  []InvoiceLineItem `json:"lineItems"`
	Subtotal   int64             `json:"subtotal,string"`
	Tax        int64             `json:"tax,string"`
	Total      int64             `json:"total,string"`
	Currency   string            `json:"currency"`
	DueDate    string            `json:"dueDate,omitempty"`
	IssuedAt   string            `json:"issuedAt,omitempty"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreateInvoiceRequest creates a new invoice. Total is computed server-side from
// LineItems (proto invoices.v1.CreateInvoiceRequest has no amount field).
type CreateInvoiceRequest struct {
	TenantID       string            `json:"tenantId"`
	CustomerID     string            `json:"customerId,omitempty"`
	LineItems      []InvoiceLineItem `json:"lineItems"`
	Currency       string            `json:"currency"`
	DueDate        string            `json:"dueDate,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// InvoicesClient
// ────────────────────────────────────────────────────────────────────────────

// InvoicesClient provides access to the invoice-service REST API.
type InvoicesClient struct {
	c *Client
}

// Create creates a new invoice.
func (i *InvoicesClient) Create(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	var out Invoice
	if err := i.c.do(ctx, "POST", "/v1/invoices", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns invoices for a tenant, optionally filtered by status.
func (i *InvoicesClient) List(ctx context.Context, tenantID, status string) ([]*Invoice, error) {
	path := fmt.Sprintf("/v1/invoices?tenantId=%s", url.QueryEscape(tenantID))
	if status != "" {
		path += "&status=" + url.QueryEscape(status)
	}
	// invoice-service wraps the list: {"invoices":[...]} (proto JSON via envoy
	// transcoder), not a bare array. Decode the wrapper then return the slice.
	var out struct {
		Invoices []*Invoice `json:"invoices"`
	}
	if err := i.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Invoices, nil
}

// Get returns a single invoice by ID.
func (i *InvoicesClient) Get(ctx context.Context, id string) (*Invoice, error) {
	var out Invoice
	if err := i.c.do(ctx, "GET", fmt.Sprintf("/v1/invoices/%s", url.PathEscape(id)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Send emails the invoice to the customer via notify-service.
func (i *InvoicesClient) Send(ctx context.Context, id string) (bool, error) {
	var out struct {
		Sent bool `json:"sent"`
	}
	if err := i.c.do(ctx, "POST", fmt.Sprintf("/v1/invoices/%s/send", url.PathEscape(id)), nil, &out); err != nil {
		return false, err
	}
	return out.Sent, nil
}

// Void voids an invoice.
func (i *InvoicesClient) Void(ctx context.Context, id string) (*Invoice, error) {
	var out Invoice
	if err := i.c.do(ctx, "POST", fmt.Sprintf("/v1/invoices/%s/void", url.PathEscape(id)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
