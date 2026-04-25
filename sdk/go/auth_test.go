package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRefreshTokenUsesCamelCaseRequest(t *testing.T) {
	var captured map[string]any

	_, client := newTestServerFn(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"accessToken":  "access_new",
			"refreshToken": "refresh_new",
		})
	})

	pair, err := client.RefreshToken(context.Background(), "refresh_old")
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}
	if captured["refreshToken"] != "refresh_old" {
		t.Fatalf("expected refreshToken body field, got %#v", captured)
	}
	if _, ok := captured["refresh_token"]; ok {
		t.Fatalf("did not expect refresh_token body field: %#v", captured)
	}
	if pair.AccessToken != "access_new" || pair.RefreshToken != "refresh_new" {
		t.Fatalf("unexpected token pair: %#v", pair)
	}
}
