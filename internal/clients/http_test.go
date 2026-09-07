package clients

import (
	"encoding/base64"
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
		{
			name: "platform nested error with schema errors",
			body: []byte(
				`{"detail":{"success":false,"error":{"code":"SCHEMA_INVALID","message":"Schema validation failed for type 'routine'","details":{"schema_errors":[{"path":"steps","message":"Input should be a valid list"}]}}}}`,
			),
			want: "Schema validation failed for type 'routine'\n  - steps: Input should be a valid list",
		},
		{
			name: "platform nested error with multiple schema errors",
			body: []byte(
				`{"detail":{"success":false,"error":{"code":"SCHEMA_INVALID","message":"Schema validation failed","details":{"schema_errors":[{"path":"steps[0].step","message":"Field required"},{"path":"steps[0].type","message":"Input should be 'node', 'branch', 'finish' or 'branchnode'"}]}}}}`,
			),
			want: "Schema validation failed\n  - steps[0].step: Field required\n  - steps[0].type: Input should be 'node', 'branch', 'finish' or 'branchnode'",
		},
		{
			name: "platform nested error without schema details",
			body: []byte(
				`{"detail":{"success":false,"error":{"code":"PLATFORM_ERROR","message":"Invalid request data"}}}`,
			),
			want: "Invalid request data",
		},
		{
			name: "trpc json error",
			body: []byte(
				`{"error":{"json":{"message":"Unsupported POST-request to query procedure","code":-32005}}}`,
			),
			want: "Unsupported POST-request to query procedure",
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

func TestExtractFileRefCandidates(t *testing.T) {
	ambiguous := []byte(`{"detail":{"success":false,"error":{"code":"FILE_REF_AMBIGUOUS",` +
		`"message":"ambiguous","details":{"candidates":[` +
		`{"fileId":"f-1","name":"report.pdf","size":10,"createdAt":"2026-01-01T00:00:00Z"},` +
		`{"fileId":"f-2","name":"report.pdf","size":20,"createdAt":"2026-01-02T00:00:00Z"}` +
		`]}}}}`)
	candidates := ExtractFileRefCandidates(ambiguous)
	if len(candidates) != 2 {
		t.Fatalf("len(candidates) = %d, want 2", len(candidates))
	}
	if candidates[0].FileID != "f-1" || candidates[1].FileID != "f-2" {
		t.Errorf("candidates = %+v, want f-1 then f-2", candidates)
	}

	otherCode := []byte(`{"detail":{"success":false,"error":{"code":"FILE_VERSION_CONFLICT",` +
		`"message":"lost the race","details":{}}}}`)
	if got := ExtractFileRefCandidates(otherCode); got != nil {
		t.Errorf("candidates for a different error code = %+v, want nil", got)
	}

	plainMessage := []byte(`{"detail":{"error":{"message":"not found"}}}`)
	if got := ExtractFileRefCandidates(plainMessage); got != nil {
		t.Errorf("candidates for a body with no code = %+v, want nil", got)
	}
}

func TestApplyRequestHeaders(t *testing.T) {
	t.Run("applies Bearer token auth", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		err = ApplyRequestHeaders(req, "test-jwt-token", "", nil)
		if err != nil {
			t.Fatalf("ApplyRequestHeaders() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader != "Bearer test-jwt-token" {
			t.Errorf("Authorization header = %q, want %q", authHeader, "Bearer test-jwt-token")
		}
	})

	t.Run("applies API key auth", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		err = ApplyRequestHeaders(req, "", "test-api-key", nil)
		if err != nil {
			t.Fatalf("ApplyRequestHeaders() error = %v", err)
		}

		wantEncoded := base64.StdEncoding.EncodeToString([]byte("test-api-key"))
		authHeader := req.Header.Get("Authorization")
		if authHeader != "Basic "+wantEncoded {
			t.Errorf("Authorization header = %q, want %q", authHeader, "Basic "+wantEncoded)
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

		err = ApplyRequestHeaders(req, "", "", cookies)
		if err != nil {
			t.Fatalf("ApplyRequestHeaders() error = %v", err)
		}

		reqCookies := req.Cookies()
		if len(reqCookies) != 2 {
			t.Fatalf("expected 2 cookies, got %d", len(reqCookies))
		}
		if reqCookies[0].Name != "session" || reqCookies[0].Value != "abc123" {
			t.Errorf(
				"cookies[0] = %s=%s, want session=abc123",
				reqCookies[0].Name,
				reqCookies[0].Value,
			)
		}
		if reqCookies[1].Name != "token" || reqCookies[1].Value != "xyz789" {
			t.Errorf(
				"cookies[1] = %s=%s, want token=xyz789",
				reqCookies[1].Name,
				reqCookies[1].Value,
			)
		}
	})

	t.Run("token takes precedence over API key and cookies", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		cookies := []*http.Cookie{
			{Name: "session", Value: "abc123"},
		}

		err = ApplyRequestHeaders(req, "test-jwt-token", "test-api-key", cookies)
		if err != nil {
			t.Fatalf("ApplyRequestHeaders() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader != "Bearer test-jwt-token" {
			t.Errorf("Authorization header = %q, want %q", authHeader, "Bearer test-jwt-token")
		}

		reqCookies := req.Cookies()
		if len(reqCookies) != 0 {
			t.Errorf("expected no cookies when token is set, got %d", len(reqCookies))
		}
	})

	t.Run("API key takes precedence over cookies", func(t *testing.T) {
		req, err := newTestRequest()
		if err != nil {
			t.Fatalf("failed to create test request: %v", err)
		}

		cookies := []*http.Cookie{
			{Name: "session", Value: "abc123"},
		}

		err = ApplyRequestHeaders(req, "", "test-api-key", cookies)
		if err != nil {
			t.Fatalf("ApplyRequestHeaders() error = %v", err)
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

		err = ApplyRequestHeaders(req, "", "", nil)
		if err == nil {
			t.Fatal("ApplyRequestHeaders() expected error, got nil")
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

		err = ApplyRequestHeaders(req, "", "", cookies)
		if err != nil {
			t.Fatalf("ApplyRequestHeaders() error = %v", err)
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
