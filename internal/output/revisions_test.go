package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestFormatRevisionActor(t *testing.T) {
	tests := []struct {
		name  string
		actor *clients.RevisionActor
		want  string
	}{
		{name: "missing attribution", want: "—"},
		{
			name: "user",
			actor: &clients.RevisionActor{
				Type:        "user",
				DisplayName: "oliver@example.com",
			},
			want: "oliver@example.com",
		},
		{
			name: "api key",
			actor: &clients.RevisionActor{
				Type:        "api_key",
				DisplayName: "silverspin-release",
			},
			want: "silverspin-release (API key)",
		},
		{
			name: "service",
			actor: &clients.RevisionActor{
				Type:        "service",
				DisplayName: "release-controller",
			},
			want: "release-controller (service)",
		},
		{
			name:  "system",
			actor: &clients.RevisionActor{Type: "system"},
			want:  "system",
		},
		{
			name:  "unknown",
			actor: &clients.RevisionActor{Type: "unknown"},
			want:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRevisionActor(tt.actor); got != tt.want {
				t.Errorf("formatRevisionActor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintRevisions(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var out bytes.Buffer
		if err := PrintRevisions(&out, nil); err != nil {
			t.Fatalf("PrintRevisions() error = %v", err)
		}
		if got, want := out.String(), "No revisions found.\n"; got != want {
			t.Errorf("PrintRevisions() = %q, want %q", got, want)
		}
	})

	t.Run("attribution and legacy rows", func(t *testing.T) {
		revisions := []clients.RevisionMeta{
			{
				Revision: 48,
				Updated:  "2026-07-28",
				Actor: &clients.RevisionActor{
					Type:        "api_key",
					DisplayName: "silverspin-release",
				},
				Source: &clients.RevisionSource{Type: "cli", Version: "0.39.0"},
			},
			{Revision: 47, Updated: "2026-07-24"},
		}

		var out bytes.Buffer
		if err := PrintRevisions(&out, revisions); err != nil {
			t.Fatalf("PrintRevisions() error = %v", err)
		}
		want := "    REVISION   UPDATED      BY                             SOURCE\n" +
			"*   48         2026-07-28   silverspin-release (API key)   iai 0.39.0\n" +
			"    47         2026-07-24   —                              —\n"
		if got := out.String(); got != want {
			t.Errorf("PrintRevisions() mismatch\ngot:\n%q\nwant:\n%q", got, want)
		}
	})
}
