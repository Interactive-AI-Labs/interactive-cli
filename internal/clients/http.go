package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ExtractServerMessage extracts a human-readable error message from an API
// response body. It handles three formats:
//
//  1. Deployment API: {"message": "..."}
//  2. Platform API (nested): {"detail": {"error": {"message": "...", "details": {...}}}}
//  3. Simple API: {"detail": "..."}
//  4. Plain text fallback
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

	var platformPayload struct {
		Detail struct {
			Error struct {
				Message string `json:"message"`
				Details struct {
					SchemaErrors []struct {
						Path    string `json:"path"`
						Message string `json:"message"`
					} `json:"schema_errors"`
				} `json:"details"`
			} `json:"error"`
		} `json:"detail"`
	}
	if err := json.Unmarshal(body, &platformPayload); err == nil {
		if msg := strings.TrimSpace(platformPayload.Detail.Error.Message); msg != "" {
			if errs := platformPayload.Detail.Error.Details.SchemaErrors; len(errs) > 0 {
				var b strings.Builder
				b.WriteString(msg)
				for _, e := range errs {
					b.WriteString("\n  - ")
					if e.Path != "" {
						b.WriteString(e.Path)
						b.WriteString(": ")
					}
					b.WriteString(e.Message)
				}
				return b.String()
			}
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
// If apiKey is provided, it sets Basic Auth with the API key, if not, it adds cookies if available.
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
