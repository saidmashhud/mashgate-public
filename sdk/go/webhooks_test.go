package mashgate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// computeSignature builds a valid "v1=<hex>" signature for test use.
func computeSignature(secret, timestamp, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + body))
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}

func nowTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func TestVerifySignature_valid(t *testing.T) {
	secret := "whsec_test_secret"
	ts := nowTimestamp()
	body := `{"event_type":"payment.captured","data":{}}`
	sig := computeSignature(secret, ts, body)

	if err := VerifySignature(secret, ts, body, sig); err != nil {
		t.Fatalf("expected valid signature to pass, got: %v", err)
	}
}

func TestVerifySignature_wrongSignature(t *testing.T) {
	secret := "whsec_test_secret"
	ts := nowTimestamp()
	body := `{"event_type":"payment.captured"}`
	sig := computeSignature("wrong_secret", ts, body)

	err := VerifySignature(secret, ts, body, sig)
	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestVerifySignature_tooOld(t *testing.T) {
	secret := "whsec_test_secret"
	oldTime := time.Now().Add(-6 * time.Minute).Unix()
	ts := fmt.Sprintf("%d", oldTime)
	body := `{"event_type":"payment.captured"}`
	sig := computeSignature(secret, ts, body)

	err := VerifySignature(secret, ts, body, sig)
	if err != ErrTimestampTooOld {
		t.Fatalf("expected ErrTimestampTooOld, got: %v", err)
	}
}

func TestVerifySignature_missingPrefix(t *testing.T) {
	secret := "whsec_test_secret"
	ts := nowTimestamp()
	body := `{"event_type":"payment.captured"}`

	// Valid HMAC but without "v1=" prefix
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + body))
	sigNaked := hex.EncodeToString(mac.Sum(nil)) // no "v1=" prefix

	err := VerifySignature(secret, ts, body, sigNaked)
	if err == nil {
		t.Fatal("expected error for missing v1= prefix, got nil")
	}
}

func TestVerifySignature_emptySecret(t *testing.T) {
	ts := nowTimestamp()
	body := `{"event_type":"payment.captured"}`
	sig := computeSignature("", ts, body)

	err := VerifySignature("", ts, body, sig)
	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature for empty secret, got: %v", err)
	}
}

func TestVerifySignature_emptyTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	body := `{"event_type":"payment.captured"}`
	sig := computeSignature(secret, "", body)

	err := VerifySignature(secret, "", body, sig)
	if err != ErrMalformedHeader {
		t.Fatalf("expected ErrMalformedHeader for empty timestamp, got: %v", err)
	}
}

func TestVerifySignature_malformedTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	body := `{"event_type":"payment.captured"}`

	err := VerifySignature(secret, "not-a-number", body, "v1=abc")
	if err == nil {
		t.Fatal("expected error for non-integer timestamp, got nil")
	}
}

func TestParseEvent_valid(t *testing.T) {
	raw := []byte(`{
		"event_id": "evt_123",
		"event_type": "payment.captured",
		"event_version": 1,
		"tenant_id": "tenant_abc",
		"occurred_at": 1700000000,
		"aggregate_id": "pay_xyz",
		"data": {"paymentId": "pay_xyz", "status": "captured"}
	}`)

	event, err := ParseEvent(raw)
	if err != nil {
		t.Fatalf("ParseEvent failed: %v", err)
	}
	if event.EventID != "evt_123" {
		t.Errorf("expected event_id evt_123, got %q", event.EventID)
	}
	if event.EventType != EventPaymentCaptured {
		t.Errorf("expected event_type %q, got %q", EventPaymentCaptured, event.EventType)
	}
	if event.Data == nil {
		t.Error("expected non-nil Data")
	}
}

func TestParseEvent_invalid(t *testing.T) {
	_, err := ParseEvent([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
