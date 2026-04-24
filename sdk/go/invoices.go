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

// InvoiceLineItem is a single line on an invoice.
type InvoiceLineItem struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitAmount  int64  `json:"unitAmount"`
}

// Invoice is a merchant invoice.
type Invoice struct {
	ID             string            `json:"id"`
	TenantID       string            `json:"tenantId"`
	InvoiceNumber  string            `json:"invoiceNumber"`
	CustomerID     string            `json:"customerId,omitempty"`
	PaymentID      string            `json:"paymentId,omitempty"`
	SubscriptionID string            `json:"subscriptionId,omitempty"`
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Status         string            `json:"status"` // "draft" | "open" | "paid" | "void"
	LineItems      []InvoiceLineItem `json:"lineItems"`
	DueDate        string            `json:"dueDate,omitempty"`
	PaidAt         *time.Time        `json:"paidAt,omitempty"`
	VoidedAt       *time.Time        `json:"voidedAt,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreateInvoiceRequest creates a new invoice.
type CreateInvoiceRequest struct {
	TenantID       string            `json:"tenantId"`
	CustomerID     string            `json:"customerId,omitempty"`
	PaymentID      string            `json:"paymentId,omitempty"`
	SubscriptionID string            `json:"subscriptionId,omitempty"`
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	LineItems      []InvoiceLineItem `json:"lineItems,omitempty"`
	DueDate        string            `json:"dueDate,omitempty"`
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
	var out []*Invoice
	if err := i.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
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
