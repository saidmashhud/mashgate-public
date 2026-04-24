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

// ChatChannel is a messaging channel within a tenant.
type ChatChannel struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	ChannelID   string    `json:"channelId"`
	Name        string    `json:"name"`
	ChannelType string    `json:"channelType"`
	MemberIDs   []string  `json:"memberIds"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ChatMessage is a message in a channel.
type ChatMessage struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	ChannelID   string    `json:"channelId"`
	SenderID    string    `json:"senderId"`
	Content     string    `json:"content,omitempty"`
	ContentType string    `json:"contentType"`
	Payload     string    `json:"payload,omitempty"`
	DeletedAt   *string   `json:"deletedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreateChannelRequest creates a new chat channel.
type CreateChannelRequest struct {
	TenantID    string   `json:"tenantId"`
	ChannelID   string   `json:"channelId"`
	Name        string   `json:"name"`
	ChannelType string   `json:"channelType,omitempty"`
	MemberIDs   []string `json:"memberIds,omitempty"`
}

// SendMessageRequest sends a message to a channel.
type SendMessageRequest struct {
	TenantID    string `json:"tenantId"`
	SenderID    string `json:"senderId"`
	Content     string `json:"content"`
	ContentType string `json:"contentType,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// ChatClient
// ────────────────────────────────────────────────────────────────────────────

// ChatClient provides access to the chat-service REST API.
type ChatClient struct {
	c *Client
}

// CreateChannel creates a new chat channel.
func (ch *ChatClient) CreateChannel(ctx context.Context, req CreateChannelRequest) (*ChatChannel, error) {
	var out ChatChannel
	if err := ch.c.do(ctx, "POST", "/v1/chat/channels", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListChannels returns channels for a tenant.
func (ch *ChatClient) ListChannels(ctx context.Context, tenantID string) ([]*ChatChannel, error) {
	path := fmt.Sprintf("/v1/chat/channels?tenantId=%s", url.QueryEscape(tenantID))
	var out []*ChatChannel
	if err := ch.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SendMessage posts a message to a channel.
func (ch *ChatClient) SendMessage(ctx context.Context, channelID string, req SendMessageRequest) (*ChatMessage, error) {
	var out ChatMessage
	path := fmt.Sprintf("/v1/chat/channels/%s/messages", url.PathEscape(channelID))
	if err := ch.c.do(ctx, "POST", path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListMessages returns messages in a channel with optional cursor pagination.
func (ch *ChatClient) ListMessages(ctx context.Context, channelID, tenantID, before string, limit int) ([]*ChatMessage, error) {
	path := fmt.Sprintf("/v1/chat/channels/%s/messages?tenantId=%s",
		url.PathEscape(channelID), url.QueryEscape(tenantID))
	if before != "" {
		path += "&before=" + url.QueryEscape(before)
	}
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}
	var out []*ChatMessage
	if err := ch.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteMessage soft-deletes a message.
func (ch *ChatClient) DeleteMessage(ctx context.Context, channelID, messageID, tenantID string) error {
	path := fmt.Sprintf("/v1/chat/channels/%s/messages/%s?tenantId=%s",
		url.PathEscape(channelID), url.PathEscape(messageID), url.QueryEscape(tenantID))
	return ch.c.do(ctx, "DELETE", path, nil, nil)
}
