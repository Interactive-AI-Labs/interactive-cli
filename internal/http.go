package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// It prefers a JSON payload with a "message" field, falling back to the raw
// body text if that field is missing or empty.
func ExtractServerMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if msg := strings.TrimSpace(payload.Message); msg != "" {
			return msg
		}
	}

	if msg := strings.TrimSpace(string(body)); msg != "" {
		return msg
	}

	return ""
}

// NewJSONRequestWithCookies constructs an *http.Request with a JSON body
// and attaches the provided cookies.
func NewJSONRequestWithCookies(ctx context.Context, method, url string, body []byte, cookies []*http.Cookie) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for _, c := range cookies {
		if c == nil {
			continue
		}
		req.AddCookie(c)
	}

	return req, nil
}
