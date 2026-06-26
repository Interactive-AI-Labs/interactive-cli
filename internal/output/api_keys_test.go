package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintRouterAPIKeyListFormatsMoney(t *testing.T) {
	oldExpiry := "2000-01-01T00:00:00Z"
	keys := []clients.RouterAPIKey{
		{
			ID:             "1",
			Name:           "a",
			KeyPreview:     "sk",
			Limit:          "100",
			LimitRemaining: "99",
			CreatedAt:      "2026-01-01T00:00:00Z",
		},
		{
			ID:         "2",
			Name:       "b",
			KeyPreview: "sk",
			Limit:      "100.00",
			ExpiresAt:  &oldExpiry,
			CreatedAt:  "2026-01-01T00:00:00Z",
		},
		{
			ID:         "3",
			Name:       "c",
			KeyPreview: "sk",
			Limit:      float64(1.5),
			CreatedAt:  "2026-01-01T00:00:00Z",
		},
	}

	var buf bytes.Buffer
	if err := PrintRouterAPIKeyList(
		&buf,
		keys,
		[]string{"id", "name", "status", "key", "limit", "remaining", "expires_at", "created_at"},
	); err != nil {
		t.Fatalf("PrintRouterAPIKeyList() error = %v", err)
	}
	got := buf.String()
	for _, want := range []string{"STATUS", "REMAINING", "EXPIRES", "$100.00", "$99.00", "$1.50", "expired", "never"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\t100\t") || strings.Contains(got, "\t100.00\t") {
		t.Fatalf("output used raw limit formatting:\n%s", got)
	}
}
