package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMail_GetMyMailbox(t *testing.T) {
	want := Mailbox{
		MailboxID:   "mb-1",
		TenantID:    "tnt-1",
		SubjectID:   "u-1",
		Email:       "demo@mail.entry-i.com",
		Status:      MailboxStatusActive,
		QuotaBytes:  5_368_709_120,
		UsedBytes:   42_342_342,
		DisplayName: "Demo User",
		CreatedAt:   time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	}

	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/mail/mailboxes/me" {
			t.Errorf("expected path /v1/mail/mailboxes/me, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	})

	got, err := client.Mail.GetMyMailbox(context.Background())
	if err != nil {
		t.Fatalf("GetMyMailbox: %v", err)
	}
	if got.Email != want.Email || got.Status != want.Status {
		t.Errorf("mismatch: got %+v, want %+v", got, want)
	}
}

func TestMail_ListMessages_query(t *testing.T) {
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("folder") != string(MessageFolderSent) {
			t.Errorf("expected folder=SENT, got %q", q.Get("folder"))
		}
		if q.Get("limit") != "25" {
			t.Errorf("expected limit=25, got %q", q.Get("limit"))
		}
		if q.Get("cursor") != "abc" {
			t.Errorf("expected cursor=abc, got %q", q.Get("cursor"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ListMailMessagesResponse{Items: []MailMessagePreview{}})
	})

	_, err := client.Mail.ListMessages(context.Background(), ListMailMessagesQuery{
		Folder: MessageFolderSent,
		Limit:  25,
		Cursor: "abc",
	})
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
}

func TestMail_SendMessage_idempotency(t *testing.T) {
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/mail/messages" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body SendMailRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.IdempotencyKey != "send-1" {
			t.Errorf("expected idempotency_key send-1, got %q", body.IdempotencyKey)
		}
		if len(body.To) != 1 || body.To[0] != "alice@example.com" {
			t.Errorf("unexpected to=%v", body.To)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SendMailResponse{
			MessageID: "msg-1",
			Status:    SendStatusQueued,
			QueuedAt:  time.Now(),
		})
	})

	out, err := client.Mail.SendMessage(context.Background(), SendMailRequest{
		To:             []string{"alice@example.com"},
		Subject:        "hi",
		BodyText:       "test",
		IdempotencyKey: "send-1",
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if out.Status != SendStatusQueued || out.MessageID != "msg-1" {
		t.Errorf("unexpected response: %+v", out)
	}
}

func TestMail_DeleteMessage_softByDefault(t *testing.T) {
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.RawQuery, "hard_delete=false") {
			t.Errorf("expected hard_delete=false, got %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := client.Mail.DeleteMessage(context.Background(), "msg-1", false); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
}

func TestMail_RotateDKIM_defaultsTo2048(t *testing.T) {
	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		var body RotateDKIMRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.KeyBits != 2048 {
			t.Errorf("expected default key_bits=2048, got %d", body.KeyBits)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MailDomain{
			DomainID: "d-1",
			Status:   DomainStatusActive,
		})
	})
	_, err := client.Mail.RotateDKIM(context.Background(), "d-1", 0)
	if err != nil {
		t.Fatalf("RotateDKIM: %v", err)
	}
}
