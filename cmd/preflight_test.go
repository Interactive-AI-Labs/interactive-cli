package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRunUpdatePreflight(t *testing.T) {
	tests := []struct {
		name           string
		revision       int
		updated        string
		liveErr        error
		expectSet      bool
		expectRevision int
		wantOut        string
		wantErr        string
	}{
		{
			name:     "prints banner",
			revision: 12,
			updated:  "2026-07-23T16:40:00Z",
			wantOut:  "Live: revision 12, last updated 2026-07-23 16:40 UTC — this update creates revision 13\n",
		},
		{
			name:    "fails open on describe error",
			liveErr: errors.New("boom"),
			wantOut: "⚠ could not fetch live state (boom) — proceeding without pre-flight check\n",
		},
		{
			name:           "expect-revision match proceeds",
			revision:       12,
			expectSet:      true,
			expectRevision: 12,
			wantOut:        "Live: revision 12 — this update creates revision 13\n",
		},
		{
			name:           "expect-revision mismatch errors without banner",
			revision:       14,
			expectSet:      true,
			expectRevision: 12,
			wantOut:        "",
			wantErr:        "live revision is 14, expected 12 — not applying (--expect-revision)",
		},
		{
			name:           "expect-revision with describe error errors without fail-open note",
			liveErr:        errors.New("boom"),
			expectSet:      true,
			expectRevision: 12,
			wantOut:        "",
			wantErr:        "--expect-revision 12: could not verify live revision: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runUpdatePreflight(
				&buf, tt.revision, tt.updated, tt.liveErr, tt.expectSet, tt.expectRevision,
			)
			if got := buf.String(); got != tt.wantOut {
				t.Errorf("output = %q, want %q", got, tt.wantOut)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
