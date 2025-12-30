package internal

import (
	"net/http"
	"strings"
	"testing"
)

func TestExtractServerMessage(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want string
	}{
		{
			name: "empty body",
			body: []byte{},
			want: "",
		},
		{
			name: "deployment message field",
			body: []byte(`{"message": "something went wrong"}`),
			want: "something went wrong",
		},
		{
			name: "api detail field",
			body: []byte(`{"detail": "not found"}`),
			want: "not found",
		},
		{
			name: "message takes precedence over detail",
			body: []byte(`{"message": "error message", "detail": "detail message"}`),
			want: "error message",
		},
		{
			name: "plain text fallback",
			body: []byte(`This is a plain text error`),
			want: "This is a plain text error",
		},
		{
			name: "empty message field uses plain text",
			body: []byte(`{"message": ""}`),
			want: `{"message": ""}`,
		},
		{
			name: "whitespace-only message field uses plain text",
			body: []byte(`{"message": "   "}`),
			want: `{"message": "   "}`,
		},
		{
			name: "trims whitespace from message",
			body: []byte(`{"message": "  error message  "}`),
			want: "error message",
		},
		{
			name: "trims whitespace from detail",
			body: []byte(`{"detail": "  detail message  "}`),
			want: "detail message",
		},
		{
			name: "trims whitespace from plain text",
			body: []byte(`  plain text  `),
			want: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractServerMessage(tt.body)
			if got != tt.want {
				t.Errorf("ExtractServerMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyAuth(t *testing.T) {
	t.Run("applies API key auth", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		apiKey := "test-api-key"
		err = ApplyAuth(req, apiKey, nil)
		if err != nil {
			t.Fatalf("ApplyAuth() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			t.Fatal("Authorization header not set")
		}
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Errorf("Authorization header should start with 'Basic ', got %q", authHeader)
		}
	})

	t.Run("applies cookie auth", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		cookies := []*http.Cookie{
			{Name: "session", Value: "abc123"},
			{Name: "token", Value: "xyz789"},
		}

		err = ApplyAuth(req, "", cookies)
		if err != nil {
			t.Fatalf("ApplyAuth() error = %v", err)
		}

		reqCookies := req.Cookies()
		if len(reqCookies) != 2 {
			t.Fatalf("expected 2 cookies, got %d", len(reqCookies))
		}
	})

	t.Run("API key takes precedence over cookies", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		apiKey := "test-api-key"
		cookies := []*http.Cookie{
			{Name: "session", Value: "abc123"},
		}

		err = ApplyAuth(req, apiKey, cookies)
		if err != nil {
			t.Fatalf("ApplyAuth() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			t.Fatal("Authorization header not set")
		}

		reqCookies := req.Cookies()
		if len(reqCookies) != 0 {
			t.Errorf("expected no cookies when API key is set, got %d", len(reqCookies))
		}
	})

	t.Run("returns error when no auth available", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		err = ApplyAuth(req, "", nil)
		if err == nil {
			t.Fatal("ApplyAuth() expected error, got nil")
		}

		if !strings.Contains(err.Error(), "no authentication method available") {
			t.Errorf("error should mention 'no authentication method available', got: %v", err)
		}
	})

	t.Run("skips nil cookies", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		cookies := []*http.Cookie{
			nil,
			{Name: "session", Value: "abc123"},
			nil,
		}

		err = ApplyAuth(req, "", cookies)
		if err != nil {
			t.Fatalf("ApplyAuth() error = %v", err)
		}

		reqCookies := req.Cookies()
		if len(reqCookies) != 1 {
			t.Fatalf("expected 1 cookie, got %d", len(reqCookies))
		}
	})
}

func newTestRequest() (*http.Request, error) {
	return http.NewRequest(http.MethodGet, "http://example.com", nil)
}
