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
	AcceptedCurrencies    []string          `json:"acceptedCurrencies"`
	MaxTransactionAmount  string            `json:"maxTransactionAmount"`
	DailyVolumeLimit      string            `json:"dailyVolumeLimit"`
	MonthlyVolumeLimit    string            `json:"monthlyVolumeLimit"`
	PrimaryCurrency       string            `json:"primaryCurrency"`
	CryptoEnabled         bool              `json:"cryptoEnabled"`
	FiatEnabled           bool              `json:"fiatEnabled"`
	AllowedPaymentMethods []string          `json:"allowedPaymentMethods"`
	Metadata              map[string]string `json:"metadata,omitempty"`
}

type MerchantProfile struct {
	MerchantID         string         `json:"merchantId"`
	TenantID           string         `json:"tenantId"`
	SubjectID          string         `json:"subjectId"`
	MerchantType       MerchantType   `json:"merchantType"`
	Status             MerchantStatus `json:"status"`
	DisplayName        string         `json:"displayName"`
	LegalName          string         `json:"legalName"`
	RegistrationNumber string         `json:"registrationNumber"`
	CountryCode        string         `json:"countryCode"`
	KycCheckID         string         `json:"kycCheckId"`
	Config             MerchantConfig `json:"config"`
	AcceptedBy         string         `json:"acceptedBy"`
	RejectedBy         string         `json:"rejectedBy"`
	RejectionReason    string         `json:"rejectionReason"`
	SuspendedBy        string         `json:"suspendedBy"`
	SuspendReason      string         `json:"suspendReason"`
	CreatedAt          string         `json:"createdAt"`
	UpdatedAt          string         `json:"updatedAt"`
	AcceptedAt         *string        `json:"acceptedAt,omitempty"`
	SuspendedAt        *string        `json:"suspendedAt,omitempty"`
}

type OnboardMerchantRequest struct {
	TenantID           string         `json:"tenantId"`
	SubjectID          string         `json:"subjectId"`
	MerchantType       MerchantType   `json:"merchantType"`
	DisplayName        string         `json:"displayName"`
	LegalName          string         `json:"legalName"`
	RegistrationNumber string         `json:"registrationNumber"`
	CountryCode        string         `json:"countryCode"`
	Config             MerchantConfig `json:"config"`
	IdempotencyKey     string         `json:"idempotencyKey,omitempty"`
}

type ListMerchantsResponse struct {
	Merchants  []MerchantProfile `json:"merchants"`
	NextCursor *string           `json:"nextCursor,omitempty"`
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
