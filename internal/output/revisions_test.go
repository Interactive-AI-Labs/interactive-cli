package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestFormatRevisionActor(t *testing.T) {
	tests := []struct {
		name  string
		actor *deployment.RevisionActor
		want  string
	}{
		{name: "missing attribution", want: "—"},
		{
			name: "user",
			actor: &deployment.RevisionActor{
				Type:        "user",
				DisplayName: "oliver@example.com",
			},
			want: "oliver@example.com",
		},
		{
			name: "api key",
			actor: &deployment.RevisionActor{
				Type:        "api_key",
				DisplayName: "silverspin-release",
			},
			want: "silverspin-release (API key)",
		},
		{
			name: "service",
			actor: &deployment.RevisionActor{
				Type:        "service",
				DisplayName: "release-controller",
			},
			want: "release-controller (service)",
		},
		{
			name:  "system",
			actor: &deployment.RevisionActor{Type: "system"},
			want:  "system",
		},
		{
			name: "named system",
			actor: &deployment.RevisionActor{
				Type:        "system",
				DisplayName: "platform-auto-rollback",
			},
			want: "platform-auto-rollback (system)",
		},
		{
			name:  "unknown",
			actor: &deployment.RevisionActor{Type: "unknown"},
			want:  "unknown",
		},
		{
			name: "future actor type",
			actor: &deployment.RevisionActor{
				Type:        "workload",
				DisplayName: "release-job",
			},
			want: "release-job",
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
	tests := []struct {
		name      string
		revisions []deployment.RevisionMeta
		want      string
	}{
		{
			name: "empty",
			want: "No revisions found.\n",
		},
		{
			name: "attribution and legacy rows",
			revisions: []deployment.RevisionMeta{
				{
					Revision: 48,
					Updated:  "2026-07-28",
					Actor: &deployment.RevisionActor{
						Type:        "api_key",
						DisplayName: "silverspin-release",
					},
					Source: &deployment.RevisionSource{Type: "cli", Version: "0.39.0"},
				},
				{Revision: 47},
			},
			want: "    REVISION   UPDATED      BY                             SOURCE\n" +
				"*   48         2026-07-28   silverspin-release (API key)   iai 0.39.0\n" +
				"    47         —            —                              —\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintRevisions(&out, tt.revisions); err != nil {
				t.Fatalf("PrintRevisions() error = %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("PrintRevisions() mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
