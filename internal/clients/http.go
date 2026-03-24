package clients

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type deploymentError struct {
	Message string `json:"message"`
}

type schemaError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type platformError struct {
	Detail struct {
		Error struct {
			Message string `json:"message"`
			Details struct {
				SchemaErrors []schemaError `json:"schema_errors"`
			} `json:"details"`
		} `json:"error"`
	} `json:"detail"`
}

type platformAPIError struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type simpleError struct {
	Detail string `json:"detail"`
}

// ExtractServerMessage extracts a human-readable error message from an API
// response body. It handles four formats:
//
//  1. Deployment API: {"message": "..."}
//  2. Platform API (nested): {"detail": {"error": {"message": "...", "details": {...}}}}
//  3. Simple API: {"detail": "..."}
//  4. Plain text fallback
func ExtractServerMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var dp deploymentError
	if err := json.Unmarshal(body, &dp); err == nil {
		if msg := strings.TrimSpace(dp.Message); msg != "" {
			return msg
		}
	}

	var pp platformError
	if err := json.Unmarshal(body, &pp); err == nil {
		msg := strings.TrimSpace(pp.Detail.Error.Message)
		errs := pp.Detail.Error.Details.SchemaErrors

		if msg != "" && len(errs) > 0 {
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

		if msg != "" {
			return msg
		}
	}

	var pa platformAPIError
	if err := json.Unmarshal(body, &pa); err == nil && !pa.Success {
		if msg := strings.TrimSpace(pa.Error.Message); msg != "" {
			return msg
		}
	}

	var sp simpleError
	if err := json.Unmarshal(body, &sp); err == nil {
		if msg := strings.TrimSpace(sp.Detail); msg != "" {
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
