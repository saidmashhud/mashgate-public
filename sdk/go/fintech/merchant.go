package fintech

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type MerchantService struct{ c *Client }

type MerchantStatus string

const (
	MerchantStatusPending     MerchantStatus = "MERCHANT_STATUS_PENDING"
	MerchantStatusKycRequired MerchantStatus = "MERCHANT_STATUS_KYC_REQUIRED"
	MerchantStatusUnderReview MerchantStatus = "MERCHANT_STATUS_UNDER_REVIEW"
	MerchantStatusAccepted    MerchantStatus = "MERCHANT_STATUS_ACCEPTED"
	MerchantStatusRejected    MerchantStatus = "MERCHANT_STATUS_REJECTED"
	MerchantStatusSuspended   MerchantStatus = "MERCHANT_STATUS_SUSPENDED"
	MerchantStatusOffboarded  MerchantStatus = "MERCHANT_STATUS_OFFBOARDED"
)

type MerchantType string

const (
	MerchantTypeIndividual MerchantType = "MERCHANT_TYPE_INDIVIDUAL"
	MerchantTypeBusiness   MerchantType = "MERCHANT_TYPE_BUSINESS"
	MerchantTypeDAO        MerchantType = "MERCHANT_TYPE_DAO"
)

type MerchantConfig struct {
	AcceptedCurrencies    []string          `json:"accepted_currencies"`
	MaxTransactionAmount  string            `json:"max_transaction_amount"`
	DailyVolumeLimit      string            `json:"daily_volume_limit"`
	MonthlyVolumeLimit    string            `json:"monthly_volume_limit"`
	PrimaryCurrency       string            `json:"primary_currency"`
	CryptoEnabled         bool              `json:"crypto_enabled"`
	FiatEnabled           bool              `json:"fiat_enabled"`
	AllowedPaymentMethods []string          `json:"allowed_payment_methods"`
	Metadata              map[string]string `json:"metadata,omitempty"`
}

type MerchantProfile struct {
	MerchantID         string         `json:"merchant_id"`
	TenantID           string         `json:"tenant_id"`
	SubjectID          string         `json:"subject_id"`
	MerchantType       MerchantType   `json:"merchant_type"`
	Status             MerchantStatus `json:"status"`
	DisplayName        string         `json:"display_name"`
	LegalName          string         `json:"legal_name"`
	RegistrationNumber string         `json:"registration_number"`
	CountryCode        string         `json:"country_code"`
	KycCheckID         string         `json:"kyc_check_id"`
	Config             MerchantConfig `json:"config"`
	AcceptedBy         string         `json:"accepted_by"`
	RejectedBy         string         `json:"rejected_by"`
	RejectionReason    string         `json:"rejection_reason"`
	SuspendedBy        string         `json:"suspended_by"`
	SuspendReason      string         `json:"suspend_reason"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
	AcceptedAt         *string        `json:"accepted_at,omitempty"`
	SuspendedAt        *string        `json:"suspended_at,omitempty"`
}

type OnboardMerchantRequest struct {
	TenantID           string         `json:"tenant_id"`
	SubjectID          string         `json:"subject_id"`
	MerchantType       MerchantType   `json:"merchant_type"`
	DisplayName        string         `json:"display_name"`
	LegalName          string         `json:"legal_name"`
	RegistrationNumber string         `json:"registration_number"`
	CountryCode        string         `json:"country_code"`
	Config             MerchantConfig `json:"config"`
	IdempotencyKey     string         `json:"idempotency_key,omitempty"`
}

type ListMerchantsResponse struct {
	Merchants  []MerchantProfile `json:"merchants"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

func (s *MerchantService) Onboard(ctx context.Context, req OnboardMerchantRequest, idempotencyKey string) (*MerchantProfile, error) {
	req.TenantID = s.c.tenantID
	var out MerchantProfile
	if err := s.c.do(ctx, http.MethodPost, "/v1/merchants", req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MerchantService) Get(ctx context.Context, merchantID string) (*MerchantProfile, error) {
	path := fmt.Sprintf("/v1/merchants/%s?tenant_id=%s", merchantID, s.c.tenantID)
	var out MerchantProfile
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MerchantService) List(ctx context.Context, status MerchantStatus, limit int, cursor string) (*ListMerchantsResponse, error) {
	qs := url.Values{}
	qs.Set("tenant_id", s.c.tenantID)
	if status != "" {
		qs.Set("status", string(status))
	}
	if limit > 0 {
		qs.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		qs.Set("cursor", cursor)
	}
	var out ListMerchantsResponse
	if err := s.c.do(ctx, http.MethodGet, "/v1/merchants?"+qs.Encode(), nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MerchantService) Accept(ctx context.Context, merchantID, note string) (*MerchantProfile, error) {
	body := map[string]string{"tenant_id": s.c.tenantID, "merchant_id": merchantID, "note": note}
	var out MerchantProfile
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/merchants/%s/accept", merchantID), body, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MerchantService) Reject(ctx context.Context, merchantID, reason string) (*MerchantProfile, error) {
	body := map[string]string{"tenant_id": s.c.tenantID, "merchant_id": merchantID, "rejection_reason": reason}
	var out MerchantProfile
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/merchants/%s/reject", merchantID), body, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MerchantService) Suspend(ctx context.Context, merchantID, reason string) (*MerchantProfile, error) {
	body := map[string]string{"tenant_id": s.c.tenantID, "merchant_id": merchantID, "suspend_reason": reason}
	var out MerchantProfile
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/merchants/%s/suspend", merchantID), body, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *MerchantService) Reinstate(ctx context.Context, merchantID, note string) (*MerchantProfile, error) {
	body := map[string]string{"tenant_id": s.c.tenantID, "merchant_id": merchantID, "note": note}
	var out MerchantProfile
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/merchants/%s/reinstate", merchantID), body, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
