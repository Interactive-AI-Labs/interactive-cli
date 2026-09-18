package platform

import (
	"errors"
	"net/http"
	"testing"
)

// Typing a 403 or a 500 as not-found would report a refusal or an outage as a
// missing record.
func TestPromptError(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		status       string
		body         string
		wantNotFound bool
		wantMessage  string
	}{
		{
			name:         "404 with a server message",
			statusCode:   http.StatusNotFound,
			status:       "404 Not Found",
			body:         `{"success":false,"error":{"message":"No prompt named 'nope'"}}`,
			wantNotFound: true,
			wantMessage:  "No prompt named 'nope'",
		},
		{
			// ExtractServerMessage falls through to the raw body, so an unparsable
			// error still reaches the user rather than being swallowed.
			name:         "404 with an unparsable body",
			statusCode:   http.StatusNotFound,
			status:       "404 Not Found",
			body:         "not json",
			wantNotFound: true,
			wantMessage:  "not json",
		},
		{
			name:         "404 with an empty body still says what happened",
			statusCode:   http.StatusNotFound,
			status:       "404 Not Found",
			wantNotFound: true,
			wantMessage:  "failed to get prompt: server returned 404 Not Found",
		},
		{
			name:        "403 is not a not-found",
			statusCode:  http.StatusForbidden,
			status:      "403 Forbidden",
			body:        `{"success":false,"error":{"message":"forbidden"}}`,
			wantMessage: "forbidden",
		},
		{
			name:        "500 is not a not-found",
			statusCode:  http.StatusInternalServerError,
			status:      "500 Internal Server Error",
			body:        `{"success":false,"error":{"message":"boom"}}`,
			wantMessage: "boom",
		},
		{
			name:        "410 is not a not-found",
			statusCode:  http.StatusGone,
			status:      "410 Gone",
			wantMessage: "failed to get prompt: server returned 410 Gone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := promptError(tt.statusCode, tt.status, []byte(tt.body))
			var notFound *NotFoundError
			if errors.As(err, &notFound) != tt.wantNotFound {
				t.Errorf(
					"errors.As(NotFoundError) = %v, want %v",
					!tt.wantNotFound,
					tt.wantNotFound,
				)
			}
			if err.Error() != tt.wantMessage {
				t.Errorf("error = %q, want %q", err, tt.wantMessage)
			}
		})
	}
}
