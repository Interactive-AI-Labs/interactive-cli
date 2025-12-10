package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// It prefers a JSON payload with a "message" field, falling back to the raw
// body text if that field is missing or empty.
func ExtractServerMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var deploymentPayload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &deploymentPayload); err == nil {
		if msg := strings.TrimSpace(deploymentPayload.Message); msg != "" {
			return msg
		}
	}

	var apiPayload struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(body, &apiPayload); err == nil {
		if msg := strings.TrimSpace(apiPayload.Detail); msg != "" {
			return msg
		}
	}

	if msg := strings.TrimSpace(string(body)); msg != "" {
		return msg
	}

	return ""
}

// ApplyAuth adds authentication to an HTTP request.
// If apiKey is provided, it sets Basic Auth with the API key.
// The API key should already be in the format "publicKey:secretKey".
// Otherwise, it adds cookies if available.
// Returns an error if neither authentication method is available.
func ApplyAuth(req *http.Request, apiKey string, cookies []*http.Cookie) error {
	if apiKey != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(apiKey))
		req.Header.Set("Authorization", "Basic "+encoded)
		return nil
	}

	if len(cookies) > 0 {
		for _, c := range cookies {
			if c != nil {
				req.AddCookie(c)
			}
		}
		return nil
	}

	return fmt.Errorf("no authentication method available: provide an API key or log in")
}
