package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDecodeTRPCData(t *testing.T) {
	body := []byte(`{"result":{"data":{"json":{"publicKey":"pk","secretKey":"sk"}}}}`)
	var got ProjectAPIKey
	if err := decodeTRPCData(body, &got); err != nil {
		t.Fatalf("decodeTRPCData() error = %v", err)
	}
	if got.PublicKey != "pk" || got.SecretKey != "sk" {
		t.Fatalf("decoded key = %#v", got)
	}
}

func TestCreateRouterAPIKeyUsesCookieAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/project-1/openrouter-keys" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		if _, err := r.Cookie("next-auth.session-token"); err != nil {
			t.Fatalf("missing session cookie: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(
			[]byte(
				`{"id":"id","key":"sk-or","name":"dev","key_preview":"sk-or-***","project_id":"project-1","user_id":"u","disabled":false,"created_at":"2026-01-01T00:00:00Z"}`,
			),
		)
	}))
	defer server.Close()

	client, err := NewAPIClient(
		server.URL,
		5*time.Second,
		"",
		"",
		[]*http.Cookie{{Name: "next-auth.session-token", Value: "cookie"}},
	)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}
	got, err := client.CreateRouterAPIKey(
		context.Background(),
		"project-1",
		CreateRouterAPIKeyBody{Name: "dev"},
	)
	if err != nil {
		t.Fatalf("CreateRouterAPIKey() error = %v", err)
	}
	if got.Key != "sk-or" {
		t.Fatalf("key = %q", got.Key)
	}
}

func TestKeysRejectAPIKeyAuth(t *testing.T) {
	client := &APIClient{apiKey: "pk:sk"}
	if err := client.requireCookieMode(); err == nil {
		t.Fatal("requireCookieMode() expected error")
	}
}
