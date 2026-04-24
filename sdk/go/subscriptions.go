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

// Plan is a recurring billing plan.
type Plan struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	Name      string    `json:"name"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Interval  string    `json:"interval"` // "monthly" | "yearly" | "weekly"
	TrialDays int       `json:"trialDays"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

// Subscription is an active recurring subscription.
type Subscription struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenantId"`
	CustomerID         string     `json:"customerId"`
	PlanID             string     `json:"planId"`
	Status             string     `json:"status"`
	PaymentMethodToken string     `json:"paymentMethodToken"`
	CurrentPeriodStart *time.Time `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"currentPeriodEnd,omitempty"`
	TrialEndsAt        *time.Time `json:"trialEndsAt,omitempty"`
	CancelledAt        *time.Time `json:"cancelledAt,omitempty"`
	RetryCount         int        `json:"retryCount"`
	NextRetryAt        *time.Time `json:"nextRetryAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreatePlanRequest creates a subscription plan.
type CreatePlanRequest struct {
	TenantID  string `json:"tenantId"`
	Name      string `json:"name"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Interval  string `json:"interval"`
	TrialDays int    `json:"trialDays,omitempty"`
}

// CreateSubscriptionRequest subscribes a customer to a plan.
type CreateSubscriptionRequest struct {
	TenantID           string `json:"tenantId"`
	CustomerID         string `json:"customerId"`
	PlanID             string `json:"planId"`
	PaymentMethodToken string `json:"paymentMethodToken"`
}

// ────────────────────────────────────────────────────────────────────────────
// SubscriptionsClient
// ────────────────────────────────────────────────────────────────────────────

// SubscriptionsClient provides access to the subscription-service REST API.
type SubscriptionsClient struct {
	c *Client
}

// CreatePlan creates a new subscription plan.
func (s *SubscriptionsClient) CreatePlan(ctx context.Context, req CreatePlanRequest) (*Plan, error) {
	var out Plan
	if err := s.c.do(ctx, "POST", "/v1/subscriptions/plans", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPlans returns all plans for a tenant.
func (s *SubscriptionsClient) ListPlans(ctx context.Context, tenantID string) ([]*Plan, error) {
	path := fmt.Sprintf("/v1/subscriptions/plans?tenantId=%s", url.QueryEscape(tenantID))
	var out []*Plan
	if err := s.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a new subscription.
func (s *SubscriptionsClient) Create(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	var out Subscription
	if err := s.c.do(ctx, "POST", "/v1/subscriptions", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all subscriptions for a tenant.
func (s *SubscriptionsClient) List(ctx context.Context, tenantID string) ([]*Subscription, error) {
	path := fmt.Sprintf("/v1/subscriptions?tenantId=%s", url.QueryEscape(tenantID))
	var out []*Subscription
	if err := s.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Cancel cancels a subscription.
func (s *SubscriptionsClient) Cancel(ctx context.Context, id string) (*Subscription, error) {
	var out Subscription
	if err := s.c.do(ctx, "POST", fmt.Sprintf("/v1/subscriptions/%s/cancel", url.PathEscape(id)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Pause pauses an active subscription.
func (s *SubscriptionsClient) Pause(ctx context.Context, id string) (*Subscription, error) {
	var out Subscription
	if err := s.c.do(ctx, "POST", fmt.Sprintf("/v1/subscriptions/%s/pause", url.PathEscape(id)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Resume resumes a paused subscription.
func (s *SubscriptionsClient) Resume(ctx context.Context, id string) (*Subscription, error) {
	var out Subscription
	if err := s.c.do(ctx, "POST", fmt.Sprintf("/v1/subscriptions/%s/resume", url.PathEscape(id)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
