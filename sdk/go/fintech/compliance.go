package fintech

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ComplianceService struct{ c *Client }

type AlertStatus string

const (
	AlertStatusOpen         AlertStatus = "ALERT_STATUS_OPEN"
	AlertStatusUnderReview  AlertStatus = "ALERT_STATUS_UNDER_REVIEW"
	AlertStatusResolved     AlertStatus = "ALERT_STATUS_RESOLVED"
	AlertStatusEscalated    AlertStatus = "ALERT_STATUS_ESCALATED"
	AlertStatusSARFiled     AlertStatus = "ALERT_STATUS_SAR_FILED"
	AlertStatusClosed       AlertStatus = "ALERT_STATUS_CLOSED"
)

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "ALERT_SEVERITY_LOW"
	AlertSeverityMedium   AlertSeverity = "ALERT_SEVERITY_MEDIUM"
	AlertSeverityHigh     AlertSeverity = "ALERT_SEVERITY_HIGH"
	AlertSeverityCritical AlertSeverity = "ALERT_SEVERITY_CRITICAL"
)

type AlertCategory string

const (
	AlertCategoryAML         AlertCategory = "ALERT_CATEGORY_AML"
	AlertCategorySanctions   AlertCategory = "ALERT_CATEGORY_SANCTIONS"
	AlertCategoryPEP         AlertCategory = "ALERT_CATEGORY_PEP"
	AlertCategoryFraud       AlertCategory = "ALERT_CATEGORY_FRAUD"
	AlertCategoryTransaction AlertCategory = "ALERT_CATEGORY_TRANSACTION"
	AlertCategoryKYC         AlertCategory = "ALERT_CATEGORY_KYC"
	AlertCategoryWatchlist   AlertCategory = "ALERT_CATEGORY_WATCHLIST"
)

type AlertEvidence struct {
	EvidenceType string `json:"evidenceType"`
	ReferenceID  string `json:"referenceId"`
	Description  string `json:"description"`
	URL          string `json:"url"`
}

type ComplianceAlert struct {
	AlertID      string          `json:"alertId"`
	TenantID     string          `json:"tenantId"`
	SubjectID    string          `json:"subjectId"`
	SubjectType  string          `json:"subjectType"`
	Category     AlertCategory   `json:"category"`
	Severity     AlertSeverity   `json:"severity"`
	Status       AlertStatus     `json:"status"`
	Source       string          `json:"source"`
	SourceRef    string          `json:"sourceRef"`
	Description  string          `json:"description"`
	Evidence     []AlertEvidence `json:"evidence"`
	AssignedTo   string          `json:"assignedTo"`
	ResolvedBy   string          `json:"resolvedBy"`
	ResolveNote  string          `json:"resolveNote"`
	EscalatedTo  string          `json:"escalatedTo"`
	EscalateNote string          `json:"escalateNote"`
	SarID        string          `json:"sarId"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
	DueAt        *string         `json:"dueAt,omitempty"`
}

type RaiseAlertRequest struct {
	TenantID       string          `json:"tenantId"`
	SubjectID      string          `json:"subjectId"`
	SubjectType    string          `json:"subjectType"`
	Category       AlertCategory   `json:"category"`
	Severity       AlertSeverity   `json:"severity"`
	Source         string          `json:"source"`
	SourceRef      string          `json:"sourceRef"`
	Description    string          `json:"description"`
	Evidence       []AlertEvidence `json:"evidence,omitempty"`
	IdempotencyKey string          `json:"idempotencyKey,omitempty"`
}

type ListAlertsResponse struct {
	Alerts     []ComplianceAlert `json:"alerts"`
	NextCursor *string           `json:"nextCursor,omitempty"`
}

func (s *ComplianceService) Raise(ctx context.Context, req RaiseAlertRequest, idempotencyKey string) (*ComplianceAlert, error) {
	req.TenantID = s.c.tenantID
	var out ComplianceAlert
	if err := s.c.do(ctx, http.MethodPost, "/v1/compliance/alerts", req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ComplianceService) Get(ctx context.Context, alertID string) (*ComplianceAlert, error) {
	path := fmt.Sprintf("/v1/compliance/alerts/%s?tenant_id=%s", alertID, s.c.tenantID)
	var out ComplianceAlert
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ComplianceService) List(ctx context.Context, subjectID string, status AlertStatus, severity AlertSeverity, limit int, cursor string) (*ListAlertsResponse, error) {
	qs := url.Values{}
	qs.Set("tenant_id", s.c.tenantID)
	if subjectID != "" {
		qs.Set("subject_id", subjectID)
	}
	if status != "" {
		qs.Set("status", string(status))
	}
	if severity != "" {
		qs.Set("severity", string(severity))
	}
	if limit > 0 {
		qs.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		qs.Set("cursor", cursor)
	}
	var out ListAlertsResponse
	if err := s.c.do(ctx, http.MethodGet, "/v1/compliance/alerts?"+qs.Encode(), nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ComplianceService) Resolve(ctx context.Context, alertID, resolveNote string) (*ComplianceAlert, error) {
	body := map[string]string{"tenant_id": s.c.tenantID, "alert_id": alertID, "resolve_note": resolveNote}
	var out ComplianceAlert
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/compliance/alerts/%s/resolve", alertID), body, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ComplianceService) Escalate(ctx context.Context, alertID, escalatedTo, note string) (*ComplianceAlert, error) {
	body := map[string]string{"tenant_id": s.c.tenantID, "alert_id": alertID, "escalated_to": escalatedTo, "escalate_note": note}
	var out ComplianceAlert
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/compliance/alerts/%s/escalate", alertID), body, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// HasOpenAlerts is a convenience check used before merchant withdrawals.
// See integration-matrix §5 (compliance gates).
func (s *ComplianceService) HasOpenAlerts(ctx context.Context, subjectID string) (bool, error) {
	resp, err := s.List(ctx, subjectID, AlertStatusOpen, "", 1, "")
	if err != nil {
		return false, err
	}
	return len(resp.Alerts) > 0, nil
}
