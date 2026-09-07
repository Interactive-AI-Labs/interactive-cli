package clients

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/buildinfo"
)

// FileRefAmbiguousCode is the platform's error code for a name matching more than one file.
const FileRefAmbiguousCode = "FILE_REF_AMBIGUOUS"

// FileRefCandidate is one file a name matched, as reported alongside a FileRefAmbiguousCode refusal.
type FileRefCandidate struct {
	FileId    string    `json:"fileId"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

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
			Code    string `json:"code"`
			Message string `json:"message"`
			Details struct {
				SchemaErrors []schemaError      `json:"schema_errors"`
				Candidates   []FileRefCandidate `json:"candidates"`
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

type trpcError struct {
	Error struct {
		JSON struct {
			Message string `json:"message"`
		} `json:"json"`
		Message string `json:"message"`
	} `json:"error"`
}

// ExtractServerMessage extracts a human-readable error message from an API
// response body. It handles five formats:
//
//  1. Deployment API: {"message": "..."}
//  2. Platform API (nested): {"detail": {"error": {"message": "...", "details": {...}}}}
//  3. Simple API: {"detail": "..."}
//  4. Platform success envelope: {"data": {"message": "..."}}
//  5. Plain text fallback
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

	var te trpcError
	if err := json.Unmarshal(body, &te); err == nil {
		if msg := strings.TrimSpace(te.Error.JSON.Message); msg != "" {
			return msg
		}
		if msg := strings.TrimSpace(te.Error.Message); msg != "" {
			return msg
		}
	}

	var pm struct {
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &pm); err == nil {
		if msg := strings.TrimSpace(pm.Data.Message); msg != "" {
			return msg
		}
	}

	if msg := strings.TrimSpace(string(body)); msg != "" {
		return msg
	}

	return ""
}

// ExtractFileRefCandidates returns the candidate files from a FileRefAmbiguousCode
// refusal, or nil for any other response.
func ExtractFileRefCandidates(body []byte) []FileRefCandidate {
	var pp platformError
	if err := json.Unmarshal(body, &pp); err != nil {
		return nil
	}
	if pp.Detail.Error.Code != FileRefAmbiguousCode {
		return nil
	}
	return pp.Detail.Error.Details.Candidates
}

// ApplyRequestHeaders adds authentication to an HTTP request.
// Priority: Bearer token > API key (Basic) > session cookies.
// Returns an error if no authentication method is available.
func ApplyRequestHeaders(
	req *http.Request,
	token string,
	apiKey string,
	cookies []*http.Cookie,
) error {
	req.Header.Set("User-Agent", buildinfo.UserAgent)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

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

	return fmt.Errorf("no authentication method available: provide a token, API key, or log in")
}
