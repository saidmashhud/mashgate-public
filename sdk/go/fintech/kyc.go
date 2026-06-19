package fintech

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type KYCService struct{ c *Client }

// ── Enums (string values matching proto enum names) ───────────────────

type KycStatus string

const (
	KycStatusUnspecified KycStatus = "KYC_STATUS_UNSPECIFIED"
	KycStatusPending     KycStatus = "KYC_STATUS_PENDING"
	KycStatusInReview    KycStatus = "KYC_STATUS_IN_REVIEW"
	KycStatusPassed      KycStatus = "KYC_STATUS_PASSED"
	KycStatusFailed      KycStatus = "KYC_STATUS_FAILED"
	KycStatusExpired     KycStatus = "KYC_STATUS_EXPIRED"
	KycStatusOverridden  KycStatus = "KYC_STATUS_OVERRIDDEN"
)

type KycSubjectType string

const (
	KycSubjectIndividual KycSubjectType = "KYC_SUBJECT_INDIVIDUAL"
	KycSubjectBusiness   KycSubjectType = "KYC_SUBJECT_BUSINESS"
)

type KycCheckType string

const (
	KycCheckIdentity  KycCheckType = "KYC_CHECK_IDENTITY"
	KycCheckAML       KycCheckType = "KYC_CHECK_AML"
	KycCheckSanctions KycCheckType = "KYC_CHECK_SANCTIONS"
	KycCheckPEP       KycCheckType = "KYC_CHECK_PEP"
	KycCheckFull      KycCheckType = "KYC_CHECK_FULL"
)

// ── Messages ──────────────────────────────────────────────────────────

type KycRiskSignal struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

type KycCheck struct {
	CheckID       string          `json:"checkId"`
	TenantID      string          `json:"tenantId"`
	SubjectID     string          `json:"subjectId"`
	SubjectType   KycSubjectType  `json:"subjectType"`
	CheckType     KycCheckType    `json:"checkType"`
	Status        KycStatus       `json:"status"`
	Provider      string          `json:"provider"`
	ProviderRef   string          `json:"providerRef"`
	RiskSignals   []KycRiskSignal `json:"riskSignals"`
	FailureCode   string          `json:"failureCode"`
	FailureReason string          `json:"failureReason"`
	OverrideBy    string          `json:"overrideBy"`
	OverrideNote  string          `json:"overrideNote"`
	CreatedAt     string          `json:"createdAt"`
	UpdatedAt     string          `json:"updatedAt"`
	ExpiresAt     *string         `json:"expiresAt,omitempty"`
}

type RequestCheckRequest struct {
	TenantID       string            `json:"tenantId"`
	SubjectID      string            `json:"subjectId"`
	SubjectType    KycSubjectType    `json:"subjectType"`
	CheckType      KycCheckType      `json:"checkType"`
	Provider       string            `json:"provider,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
}

type RequestCheckResponse struct {
	Check       KycCheck `json:"check"`
	RedirectURL *string  `json:"redirectUrl,omitempty"`
}

type ListChecksResponse struct {
	Checks     []KycCheck `json:"checks"`
	NextCursor *string    `json:"nextCursor,omitempty"`
}

type OverrideCheckRequest struct {
	TenantID     string    `json:"tenantId"`
	CheckID      string    `json:"checkId"`
	Status       KycStatus `json:"status"`
	OverrideNote string    `json:"overrideNote"`
}

// ── RPCs ──────────────────────────────────────────────────────────────

// Request submits a new KYC check. The response's RedirectURL, if set,
// must be surfaced to the end-user for provider-hosted verification.
func (s *KYCService) Request(ctx context.Context, req RequestCheckRequest, idempotencyKey string) (*RequestCheckResponse, error) {
	req.TenantID = s.c.tenantID
	var out RequestCheckResponse
	if err := s.c.do(ctx, http.MethodPost, "/v1/kyc/checks", req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a previously-submitted check by ID.
func (s *KYCService) Get(ctx context.Context, checkID string) (*KycCheck, error) {
	path := fmt.Sprintf("/v1/kyc/checks/%s?tenant_id=%s", checkID, s.c.tenantID)
	var out KycCheck
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns KYC checks for this tenant with optional filters.
func (s *KYCService) List(ctx context.Context, subjectID string, status KycStatus, limit int, cursor string) (*ListChecksResponse, error) {
	qs := url.Values{}
	qs.Set("tenant_id", s.c.tenantID)
	if subjectID != "" {
		qs.Set("subject_id", subjectID)
	}
	if status != "" {
		qs.Set("status", string(status))
	}
	if limit > 0 {
		qs.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		qs.Set("cursor", cursor)
	}
	var out ListChecksResponse
	if err := s.c.do(ctx, http.MethodGet, "/v1/kyc/checks?"+qs.Encode(), nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Override manually sets a check's status (platform operator only).
func (s *KYCService) Override(ctx context.Context, req OverrideCheckRequest) (*KycCheck, error) {
	req.TenantID = s.c.tenantID
	path := fmt.Sprintf("/v1/kyc/checks/%s/override", req.CheckID)
	var out KycCheck
	if err := s.c.do(ctx, http.MethodPost, path, req, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
