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

// NotifyTemplate is a reusable message template (SMS or email).
type NotifyTemplate struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	TemplateKey  string    `json:"templateKey"`
	Channels     []string  `json:"channels"`
	EmailSubject string    `json:"emailSubject,omitempty"`
	EmailBody    string    `json:"emailBodyHtml,omitempty"`
	SmsText      string    `json:"smsText,omitempty"`
	Vars         []string  `json:"vars"`
	CreatedAt    time.Time `json:"createdAt"`
}

// NotificationLog is a record of a sent notification.
type NotificationLog struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenantId"`
	Channel       string    `json:"channel"`
	Recipient     string    `json:"recipient"`
	TemplateKey   string    `json:"templateKey,omitempty"`
	Status        string    `json:"status"`
	Provider      string    `json:"provider,omitempty"`
	ProviderMsgID string    `json:"providerMsgId,omitempty"`
	Error         string    `json:"error,omitempty"`
	SentAt        time.Time `json:"sentAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// SendSmsRequest sends an SMS to a phone number.
type SendSmsRequest struct {
	TenantID    string            `json:"tenantId"`
	To          string            `json:"to"`
	TemplateKey string            `json:"templateKey,omitempty"`
	Vars        map[string]string `json:"vars,omitempty"`
	Text        string            `json:"text,omitempty"`
}

// SendEmailRequest sends an email using a template.
type SendEmailRequest struct {
	TenantID    string            `json:"tenantId"`
	To          string            `json:"to"`
	TemplateKey string            `json:"templateKey"`
	Vars        map[string]string `json:"vars,omitempty"`
}

// CreateTemplateRequest defines a new notification template.
type CreateTemplateRequest struct {
	TenantID     string   `json:"tenantId"`
	TemplateKey  string   `json:"templateKey"`
	Channels     []string `json:"channels"`
	EmailSubject string   `json:"emailSubject,omitempty"`
	EmailBody    string   `json:"emailBodyHtml,omitempty"`
	SmsText      string   `json:"smsText,omitempty"`
	Vars         []string `json:"vars,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// NotifyClient
// ────────────────────────────────────────────────────────────────────────────

// NotifyClient provides access to the notify-service REST API.
type NotifyClient struct {
	c *Client
}

// SendSms sends an SMS via notify-service.
func (n *NotifyClient) SendSms(ctx context.Context, req SendSmsRequest) (*NotificationLog, error) {
	var out NotificationLog
	if err := n.c.do(ctx, "POST", "/v1/notify/sms", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendEmail sends an email via notify-service.
func (n *NotifyClient) SendEmail(ctx context.Context, req SendEmailRequest) (*NotificationLog, error) {
	var out NotificationLog
	if err := n.c.do(ctx, "POST", "/v1/notify/email", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateTemplate creates a notification template.
func (n *NotifyClient) CreateTemplate(ctx context.Context, req CreateTemplateRequest) (*NotifyTemplate, error) {
	var out NotifyTemplate
	if err := n.c.do(ctx, "POST", "/v1/notify/templates", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTemplates returns templates for a tenant.
func (n *NotifyClient) ListTemplates(ctx context.Context, tenantID string) ([]*NotifyTemplate, error) {
	path := fmt.Sprintf("/v1/notify/templates?tenantId=%s", url.QueryEscape(tenantID))
	var out []*NotifyTemplate
	if err := n.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListLogs returns notification logs for a tenant.
func (n *NotifyClient) ListLogs(ctx context.Context, tenantID string, page int) ([]*NotificationLog, error) {
	path := fmt.Sprintf("/v1/notify/logs?tenantId=%s&page=%d", url.QueryEscape(tenantID), page)
	var out []*NotificationLog
	if err := n.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Telegram channel (added v1.6.0)
// ────────────────────────────────────────────────────────────────────────────

// SendTelegramRequest sends a Telegram message via notify-service.
// Either TemplateKey OR Text is required.
type SendTelegramRequest struct {
	TenantID    string            `json:"tenantId"`
	ChatID      string            `json:"chatId"`
	TemplateKey string            `json:"templateKey,omitempty"`
	Vars        map[string]string `json:"vars,omitempty"`
	Text        string            `json:"text,omitempty"`
}

// SendTelegramResponse contains notification log id + provider message id.
type SendTelegramResponse struct {
	NotificationID string `json:"notificationId"`
	ProviderMsgID  string `json:"providerMsgId,omitempty"`
	Status         string `json:"status"` // "sent" | "failed"
}

// SendTelegram sends a Telegram message via mashgate notify-service.
//
// Replaces vertical-local Telegram clients (e.g. qrapp/internal/telegram/).
// Pairing flow (chat_id lookup from /start <token>) stays in the calling vertical;
// notify-service is delivery-only.
func (n *NotifyClient) SendTelegram(ctx context.Context, req SendTelegramRequest) (*SendTelegramResponse, error) {
	var out SendTelegramResponse
	if err := n.c.do(ctx, "POST", "/v1/notify/telegram", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Tenant SMS provider (added v1.21.0)
// ────────────────────────────────────────────────────────────────────────────
//
// The tenant owner connects their OWN operator account (e.g. OsonSMS); SMS
// from that tenant is then sent with those credentials. The secret is stored
// encrypted by notify-service and is never returned — only SecretPreview.
// Requires the notify:providers:manage scope.

// SetSmsProviderRequest connects or updates the tenant's SMS provider.
// An empty Secret on update keeps the previously stored secret.
type SetSmsProviderRequest struct {
	TenantID       string `json:"tenantId"`
	Provider       string `json:"provider"` // "osonsms"
	Login          string `json:"login"`
	Secret         string `json:"secret,omitempty"`
	Sender         string `json:"sender"`
	BaseURL        string `json:"baseUrl,omitempty"`
	IsConfidential bool   `json:"isConfidential,omitempty"`
}

// SmsProviderInfo is the masked view of the tenant's SMS provider.
type SmsProviderInfo struct {
	TenantID       string    `json:"tenantId"`
	Provider       string    `json:"provider"`
	Login          string    `json:"login"`
	Sender         string    `json:"sender"`
	BaseURL        string    `json:"baseUrl,omitempty"`
	IsConfidential bool      `json:"isConfidential"`
	HasSecret      bool      `json:"hasSecret"`
	SecretPreview  string    `json:"secretPreview,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	LastCheckAt    *time.Time `json:"lastCheckAt,omitempty"`
	LastCheckOk    bool      `json:"lastCheckOk"`
	LastCheckError string    `json:"lastCheckError,omitempty"`
	LastBalance    string    `json:"lastBalance,omitempty"`
}

// TestSmsProviderResponse is the result of a live balance check at the operator
// (and of the optional test SMS).
type TestSmsProviderResponse struct {
	Ok            bool   `json:"ok"`
	Balance       string `json:"balance,omitempty"`
	Error         string `json:"error,omitempty"`
	TestMessageID string `json:"testMessageId,omitempty"`
}

// GetSmsProvider returns the tenant's SMS provider (masked). 404 when none.
func (n *NotifyClient) GetSmsProvider(ctx context.Context, tenantID string) (*SmsProviderInfo, error) {
	path := fmt.Sprintf("/v1/notify/sms-provider?tenantId=%s", url.QueryEscape(tenantID))
	var out SmsProviderInfo
	if err := n.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetSmsProvider connects or updates the tenant's SMS provider credentials.
func (n *NotifyClient) SetSmsProvider(ctx context.Context, req SetSmsProviderRequest) (*SmsProviderInfo, error) {
	var out SmsProviderInfo
	if err := n.c.do(ctx, "PUT", "/v1/notify/sms-provider", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSmsProvider disconnects the tenant's SMS provider and erases the secret.
func (n *NotifyClient) DeleteSmsProvider(ctx context.Context, tenantID string) error {
	path := fmt.Sprintf("/v1/notify/sms-provider?tenantId=%s", url.QueryEscape(tenantID))
	return n.c.do(ctx, "DELETE", path, nil, nil)
}

// TestSmsProvider checks the stored credentials live (balance query). With a
// non-empty testPhone it also sends one test SMS to that number (billed by the
// operator, recorded in the delivery log).
func (n *NotifyClient) TestSmsProvider(ctx context.Context, tenantID, testPhone string) (*TestSmsProviderResponse, error) {
	body := map[string]string{"tenantId": tenantID}
	if testPhone != "" {
		body["testPhone"] = testPhone
	}
	var out TestSmsProviderResponse
	if err := n.c.do(ctx, "POST", "/v1/notify/sms-provider/test", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
