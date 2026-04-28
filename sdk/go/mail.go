package mashgate

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Mail capability — Mashgate core primitive (ADR-0019). Mirrors mail.v1.MailService
// from mashgate/contracts/proto/v1/mail.proto, exposed over the gateway as REST
// via google.api.http transcoding.
//
// Auth: pass a user JWT for self-service mailbox operations (mail:read /
// mail:write scope), or admin/service-account credentials for tenant-wide
// operations (mail:admin scope).
//
// Events (subscribe via webhooks): mail.received / mail.sent / mail.delivered /
// mail.bounced — see contracts/events/mail.*.json.

// ────────────────────────────────────────────────────────────────────────────
// Enum-like string types from mail.proto
// ────────────────────────────────────────────────────────────────────────────

type MessageFolder string

const (
	MessageFolderInbox  MessageFolder = "MESSAGE_FOLDER_INBOX"
	MessageFolderSent   MessageFolder = "MESSAGE_FOLDER_SENT"
	MessageFolderDrafts MessageFolder = "MESSAGE_FOLDER_DRAFTS"
	MessageFolderSpam   MessageFolder = "MESSAGE_FOLDER_SPAM"
	MessageFolderTrash  MessageFolder = "MESSAGE_FOLDER_TRASH"
)

type MailboxStatus string

const (
	MailboxStatusActive MailboxStatus = "MAILBOX_STATUS_ACTIVE"
	MailboxStatusFrozen MailboxStatus = "MAILBOX_STATUS_FROZEN"
	MailboxStatusClosed MailboxStatus = "MAILBOX_STATUS_CLOSED"
)

type DomainStatus string

const (
	DomainStatusPending   DomainStatus = "DOMAIN_STATUS_PENDING"
	DomainStatusActive    DomainStatus = "DOMAIN_STATUS_ACTIVE"
	DomainStatusSuspended DomainStatus = "DOMAIN_STATUS_SUSPENDED"
)

type SendStatus string

const (
	SendStatusQueued    SendStatus = "SEND_STATUS_QUEUED"
	SendStatusSent      SendStatus = "SEND_STATUS_SENT"
	SendStatusDelivered SendStatus = "SEND_STATUS_DELIVERED"
	SendStatusFailed    SendStatus = "SEND_STATUS_FAILED"
)

// ────────────────────────────────────────────────────────────────────────────
// Domain types
// ────────────────────────────────────────────────────────────────────────────

type Mailbox struct {
	MailboxID   string        `json:"mailbox_id"`
	TenantID    string        `json:"tenant_id"`
	SubjectID   string        `json:"subject_id"`
	Email       string        `json:"email"`
	DisplayName string        `json:"display_name,omitempty"`
	Status      MailboxStatus `json:"status"`
	QuotaBytes  int64         `json:"quota_bytes"`
	UsedBytes   int64         `json:"used_bytes"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type MailMessagePreview struct {
	MessageID      string        `json:"message_id"`
	TenantID       string        `json:"tenant_id"`
	MailboxID      string        `json:"mailbox_id"`
	From           string        `json:"from"`
	To             []string      `json:"to"`
	Subject        string        `json:"subject"`
	Preview        string        `json:"preview,omitempty"`
	ReceivedAt     time.Time     `json:"received_at"`
	Read           bool          `json:"read"`
	Folder         MessageFolder `json:"folder"`
	Labels         []string      `json:"labels,omitempty"`
	HasAttachments bool          `json:"has_attachments,omitempty"`
}

type MailAttachment struct {
	AttachmentID string `json:"attachment_id"`
	Filename     string `json:"filename"`
	ContentType  string `json:"content_type"`
	SizeBytes    int64  `json:"size_bytes"`
	URL          string `json:"url"` // signed URL via Mashgate storage-service
}

type MailMessage struct {
	Preview     MailMessagePreview `json:"preview"`
	BodyText    string             `json:"body_text,omitempty"`
	BodyHTML    string             `json:"body_html,omitempty"`
	Headers     map[string]string  `json:"headers,omitempty"`
	Attachments []MailAttachment   `json:"attachments,omitempty"`
	CC          []string           `json:"cc,omitempty"`
	BCC         []string           `json:"bcc,omitempty"`
	SentAt      *time.Time         `json:"sent_at,omitempty"`
	InReplyTo   string             `json:"in_reply_to,omitempty"`
}

type MailDomain struct {
	DomainID            string       `json:"domain_id"`
	TenantID            string       `json:"tenant_id"`
	Name                string       `json:"name"`
	Status              DomainStatus `json:"status"`
	DKIMSelector        string       `json:"dkim_selector,omitempty"`
	DKIMPublicKey       string       `json:"dkim_public_key,omitempty"`
	MXRecords           []string     `json:"mx_records"`
	SPFRecord           string       `json:"spf_record,omitempty"`
	DMARCRecord         string       `json:"dmarc_record,omitempty"`
	VerificationErrors  []string     `json:"verification_errors,omitempty"`
	CreatedAt           time.Time    `json:"created_at"`
	LastVerifiedAt      *time.Time   `json:"last_verified_at,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request / Response types
// ────────────────────────────────────────────────────────────────────────────

type ListMailMessagesQuery struct {
	Folder            MessageFolder
	Limit             int
	Cursor            string
	MailboxIDOverride string // mail:admin only
}

type ListMailMessagesResponse struct {
	Items      []MailMessagePreview `json:"items"`
	NextCursor string               `json:"next_cursor,omitempty"`
	Total      int                  `json:"total,omitempty"`
}

type SendMailRequest struct {
	To             []string          `json:"to"`
	CC             []string          `json:"cc,omitempty"`
	BCC            []string          `json:"bcc,omitempty"`
	Subject        string            `json:"subject"`
	BodyText       string            `json:"body_text,omitempty"`
	BodyHTML       string            `json:"body_html,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	AttachmentIDs  []string          `json:"attachment_ids,omitempty"`
	InReplyTo      string            `json:"in_reply_to,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
}

type SendMailResponse struct {
	MessageID string     `json:"message_id"`
	Status    SendStatus `json:"status"`
	Note      string     `json:"note,omitempty"`
	QueuedAt  time.Time  `json:"queued_at"`
}

type UpdateMailMessageRequest struct {
	Read   *bool         `json:"read,omitempty"`
	Folder MessageFolder `json:"folder,omitempty"`
	Labels []string      `json:"labels,omitempty"`
}

type ListMailboxesQuery struct {
	Status MailboxStatus
	Limit  int
	Cursor string
}

type ListMailboxesResponse struct {
	Items      []Mailbox `json:"items"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

type CreateMailboxRequest struct {
	SubjectID      string `json:"subject_id"`
	Email          string `json:"email"`
	DisplayName    string `json:"display_name,omitempty"`
	QuotaBytes     int64  `json:"quota_bytes,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type ListMailDomainsResponse struct {
	Items []MailDomain `json:"items"`
}

type CreateMailDomainRequest struct {
	Name string `json:"name"`
}

type RotateDKIMRequest struct {
	KeyBits int `json:"key_bits,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// MailClient
// ────────────────────────────────────────────────────────────────────────────

type MailClient struct {
	c *Client
}

// ── User-facing (mail:read / mail:write) ────────────────────────────────────

// GetMyMailbox returns the mailbox for the authenticated subject.
func (m *MailClient) GetMyMailbox(ctx context.Context) (*Mailbox, error) {
	var out Mailbox
	if err := m.c.do(ctx, "GET", "/v1/mail/mailboxes/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListMessages returns a page of message previews from the authenticated
// subject's mailbox.
func (m *MailClient) ListMessages(ctx context.Context, q ListMailMessagesQuery) (*ListMailMessagesResponse, error) {
	v := url.Values{}
	if q.Folder != "" {
		v.Set("folder", string(q.Folder))
	}
	if q.Limit > 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Cursor != "" {
		v.Set("cursor", q.Cursor)
	}
	if q.MailboxIDOverride != "" {
		v.Set("mailbox_id_override", q.MailboxIDOverride)
	}
	path := "/v1/mail/messages"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out ListMailMessagesResponse
	if err := m.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMessage returns a single message with body and headers.
func (m *MailClient) GetMessage(ctx context.Context, messageID string) (*MailMessage, error) {
	var out MailMessage
	path := fmt.Sprintf("/v1/mail/messages/%s", url.PathEscape(messageID))
	if err := m.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendMessage queues a message for delivery. Idempotent via IdempotencyKey:
// the same key on the same tenant returns the original message_id.
func (m *MailClient) SendMessage(ctx context.Context, req SendMailRequest) (*SendMailResponse, error) {
	var out SendMailResponse
	if err := m.c.do(ctx, "POST", "/v1/mail/messages", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMessage patches mutable flags (read/folder/labels) on a message.
// Pass only the fields you want to change.
func (m *MailClient) UpdateMessage(ctx context.Context, messageID string, req UpdateMailMessageRequest) (*MailMessage, error) {
	var out MailMessage
	path := fmt.Sprintf("/v1/mail/messages/%s", url.PathEscape(messageID))
	if err := m.c.do(ctx, "PATCH", path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMessage moves a message to TRASH. If hardDelete=true and the message
// is already in TRASH, it is permanently removed.
func (m *MailClient) DeleteMessage(ctx context.Context, messageID string, hardDelete bool) error {
	path := fmt.Sprintf("/v1/mail/messages/%s?hard_delete=%t", url.PathEscape(messageID), hardDelete)
	return m.c.do(ctx, "DELETE", path, nil, nil)
}

// ── Admin (mail:admin) ──────────────────────────────────────────────────────

func (m *MailClient) ListMailboxes(ctx context.Context, q ListMailboxesQuery) (*ListMailboxesResponse, error) {
	v := url.Values{}
	if q.Status != "" {
		v.Set("status", string(q.Status))
	}
	if q.Limit > 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Cursor != "" {
		v.Set("cursor", q.Cursor)
	}
	path := "/v1/mail/mailboxes"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out ListMailboxesResponse
	if err := m.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (m *MailClient) CreateMailbox(ctx context.Context, req CreateMailboxRequest) (*Mailbox, error) {
	var out Mailbox
	if err := m.c.do(ctx, "POST", "/v1/mail/mailboxes", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (m *MailClient) ListDomains(ctx context.Context, status DomainStatus) (*ListMailDomainsResponse, error) {
	path := "/v1/mail/domains"
	if status != "" {
		path += "?status=" + string(status)
	}
	var out ListMailDomainsResponse
	if err := m.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (m *MailClient) CreateDomain(ctx context.Context, req CreateMailDomainRequest) (*MailDomain, error) {
	var out MailDomain
	if err := m.c.do(ctx, "POST", "/v1/mail/domains", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyDomain re-checks DNS records for the domain. On success status flips
// to active. On failure, returns Domain with verification_errors populated.
func (m *MailClient) VerifyDomain(ctx context.Context, domainID string) (*MailDomain, error) {
	var out MailDomain
	path := fmt.Sprintf("/v1/mail/domains/%s/verify", url.PathEscape(domainID))
	if err := m.c.do(ctx, "POST", path, struct{}{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RotateDKIM generates a new DKIM key for the domain. The previous selector
// remains valid for ~30 days for graceful DNS propagation. Default key_bits=2048.
func (m *MailClient) RotateDKIM(ctx context.Context, domainID string, keyBits int) (*MailDomain, error) {
	if keyBits == 0 {
		keyBits = 2048
	}
	req := RotateDKIMRequest{KeyBits: keyBits}
	var out MailDomain
	path := fmt.Sprintf("/v1/mail/domains/%s/dkim/rotate", url.PathEscape(domainID))
	if err := m.c.do(ctx, "POST", path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
