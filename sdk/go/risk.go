package mashgate

import (
	"context"
	"fmt"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// RiskClient — fraud-service wrapper. Mirrors RiskService in risk.proto (9 RPCs).
//
// Advisory only — RiskService returns scores + recommended actions,
// payments-orchestrator decides whether to block. Tenant always has the
// final say.
// ────────────────────────────────────────────────────────────────────────────

type RiskClient struct {
	c *Client
}

// ── Assessments ──────────────────────────────────────────────────────────

// AssessTransaction submits a transaction for risk scoring.
// Returns score (0-100), risk level, recommended_action ("approve"|"review"|"decline"),
// and list of triggered rules. Score is advisory.
func (r *RiskClient) AssessTransaction(ctx context.Context, req AssessTransactionRequest) (*RiskAssessment, error) {
	var out RiskAssessment
	if err := r.c.do(ctx, "POST", "/v1/risk/assess", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAssessment retrieves a single assessment by id.
func (r *RiskClient) GetAssessment(ctx context.Context, assessmentID string) (*RiskAssessment, error) {
	var out RiskAssessment
	if err := r.c.do(ctx, "GET", "/v1/risk/assessments/"+assessmentID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAssessments returns recent assessments for a tenant.
func (r *RiskClient) ListAssessments(ctx context.Context, tenantID string, page, pageSize int) ([]*RiskAssessment, error) {
	q := fmt.Sprintf("?tenantId=%s", url.QueryEscape(tenantID))
	if page > 0 {
		q += fmt.Sprintf("&page=%d", page)
	}
	if pageSize > 0 {
		q += fmt.Sprintf("&pageSize=%d", pageSize)
	}
	var out struct {
		Assessments []*RiskAssessment `json:"assessments"`
	}
	if err := r.c.do(ctx, "GET", "/v1/risk/assessments"+q, nil, &out); err != nil {
		return nil, err
	}
	return out.Assessments, nil
}

// ── Blocklist management ─────────────────────────────────────────────────

// AddBlocklistEntry adds an entry to tenant blocklist (email, phone, card BIN,
// IP, country). Returns 409 if already exists.
func (r *RiskClient) AddBlocklistEntry(ctx context.Context, req AddBlocklistEntryRequest) (*BlocklistEntry, error) {
	var out BlocklistEntry
	if err := r.c.do(ctx, "POST", "/v1/risk/blocklist", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListBlocklistEntries returns all blocklist entries for tenant, optionally
// filtered by entry_type.
func (r *RiskClient) ListBlocklistEntries(ctx context.Context, tenantID, entryType string) ([]*BlocklistEntry, error) {
	q := fmt.Sprintf("?tenantId=%s", url.QueryEscape(tenantID))
	if entryType != "" {
		q += "&entryType=" + url.QueryEscape(entryType)
	}
	var out struct {
		Entries []*BlocklistEntry `json:"entries"`
	}
	if err := r.c.do(ctx, "GET", "/v1/risk/blocklist"+q, nil, &out); err != nil {
		return nil, err
	}
	return out.Entries, nil
}

// RemoveBlocklistEntry deletes a blocklist entry by id.
func (r *RiskClient) RemoveBlocklistEntry(ctx context.Context, entryID string) error {
	return r.c.do(ctx, "DELETE", "/v1/risk/blocklist/"+entryID, nil, nil)
}

// ── Rules ─────────────────────────────────────────────────────────────────

// ListRiskRules returns active rules for tenant. Read-only — rule authoring is
// via admin console.
func (r *RiskClient) ListRiskRules(ctx context.Context, tenantID string) ([]*RiskRule, error) {
	var out struct {
		Rules []*RiskRule `json:"rules"`
	}
	if err := r.c.do(ctx, "GET", "/v1/risk/rules?tenantId="+url.QueryEscape(tenantID), nil, &out); err != nil {
		return nil, err
	}
	return out.Rules, nil
}

// GetRiskProfile returns aggregate risk profile of an email/phone/card (across
// all of tenant's history). Used in pre-checkout screening.
func (r *RiskClient) GetRiskProfile(ctx context.Context, tenantID, identifierType, identifier string) (*RiskProfile, error) {
	q := fmt.Sprintf("?tenantId=%s&identifierType=%s&identifier=%s",
		url.QueryEscape(tenantID), url.QueryEscape(identifierType), url.QueryEscape(identifier))
	var out RiskProfile
	if err := r.c.do(ctx, "GET", "/v1/risk/profile"+q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
