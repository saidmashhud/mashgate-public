package mashgate

import (
	"context"

	"github.com/google/uuid"
)

// ────────────────────────────────────────────────────────────────────────────
// LocalPaymentsClient — country-specific payment providers (TJ: Tcell, Korti
// Milli, Alif, Eskhata; UZ: Click, Payme, Apelsin).
//
// Mirrors LocalPaymentsService in local_payments.proto (9 RPCs).
// ADR-0009 governs which providers are wired per Mashgate deployment.
//
// Status: providers not provisioned in srv2 prod as of 2026-05-12 — all calls
// return mock-only responses. Once owner adds provider credentials to
// payments-orchestrator secrets, real flow starts.
// ────────────────────────────────────────────────────────────────────────────

type LocalPaymentsClient struct {
	c *Client
}

// ListSupportedMethods returns providers available for tenant's country.
// E.g. for TJ tenant: [{id: "tcell-mobile", name: "Tcell Mobile Money"},
//
//	{id: "korti-milli", name: "Korti Milli"}, ...].
func (l *LocalPaymentsClient) ListSupportedMethods(ctx context.Context, tenantID, country string) ([]*LocalPaymentMethod, error) {
	q := "?tenantId=" + tenantID
	if country != "" {
		q += "&country=" + country
	}
	var out struct {
		Methods []*LocalPaymentMethod `json:"methods"`
	}
	if err := l.c.do(ctx, "GET", "/v1/local-payments/methods"+q, nil, &out); err != nil {
		return nil, err
	}
	return out.Methods, nil
}

// InitiatePayment starts a local-rail payment flow. Returns provider-specific
// next_step (URL to redirect, USSD code to dial, QR to scan).
// Idempotent via IdempotencyKey header.
func (l *LocalPaymentsClient) InitiatePayment(ctx context.Context, req InitiateLocalPaymentRequest) (*LocalPaymentInitiated, error) {
	key := req.IdempotencyKey
	if key == "" {
		key = uuid.NewString()
	}
	headers := map[string]string{"Idempotency-Key": key}
	var out LocalPaymentInitiated
	if err := l.c.doWithHeader(ctx, "POST", "/v1/local-payments/initiate", headers, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConfirmPayment submits provider callback/confirmation (e.g. SMS OTP, USSD code).
func (l *LocalPaymentsClient) ConfirmPayment(ctx context.Context, paymentID string, req ConfirmLocalPaymentRequest) (*LocalPayment, error) {
	var out LocalPayment
	if err := l.c.do(ctx, "POST", "/v1/local-payments/"+paymentID+"/confirm", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPaymentStatus returns current state of a local payment.
func (l *LocalPaymentsClient) GetPaymentStatus(ctx context.Context, paymentID string) (*LocalPayment, error) {
	var out LocalPayment
	if err := l.c.do(ctx, "GET", "/v1/local-payments/"+paymentID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelPayment cancels a pending local payment (best-effort; depends on provider).
func (l *LocalPaymentsClient) CancelPayment(ctx context.Context, paymentID, reason string) error {
	req := struct {
		Reason string `json:"reason,omitempty"`
	}{Reason: reason}
	return l.c.do(ctx, "POST", "/v1/local-payments/"+paymentID+"/cancel", req, nil)
}

// ListPayments returns local payment history for tenant.
func (l *LocalPaymentsClient) ListPayments(ctx context.Context, tenantID string, page, pageSize int) ([]*LocalPayment, error) {
	q := "?tenantId=" + tenantID
	if page > 0 {
		q += "&page="
	}
	var out struct {
		Payments []*LocalPayment `json:"payments"`
	}
	if err := l.c.do(ctx, "GET", "/v1/local-payments"+q, nil, &out); err != nil {
		return nil, err
	}
	return out.Payments, nil
}
