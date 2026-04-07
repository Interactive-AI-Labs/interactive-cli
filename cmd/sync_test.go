package cmd

import (
	"bytes"
	"fmt"
	"testing"
)

func TestAllowDeleteResource(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		resource string
		want     bool
	}{
		{
			name:     "nil list",
			allowed:  nil,
			resource: "vector-stores",
			want:     false,
		},
		{
			name:     "empty list",
			allowed:  []string{},
			resource: "vector-stores",
			want:     false,
		},
		{
			name:     "exact match",
			allowed:  []string{"vector-stores"},
			resource: "vector-stores",
			want:     true,
		},
		{
			name:     "case insensitive match",
			allowed:  []string{"Vector-Stores"},
			resource: "vector-stores",
			want:     true,
		},
		{
			name:     "all keyword",
			allowed:  []string{"all"},
			resource: "vector-stores",
			want:     true,
		},
		{
			name:     "ALL keyword uppercase",
			allowed:  []string{"ALL"},
			resource: "vector-stores",
			want:     true,
		},
		{
			name:     "no match",
			allowed:  []string{"services"},
			resource: "vector-stores",
			want:     false,
		},
		{
			name:     "multiple entries with match",
			allowed:  []string{"services", "vector-stores"},
			resource: "vector-stores",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := allowDeleteResource(tt.allowed, tt.resource)
			if got != tt.want {
				t.Errorf(
					"allowDeleteResource(%v, %q) = %v, want %v",
					tt.allowed, tt.resource, got, tt.want,
				)
			}
		})
	}
}

func TestPrintSyncOutcome(t *testing.T) {
	t.Run("success prints result", func(t *testing.T) {
		var buf bytes.Buffer
		result := &SyncResult{
			Created: []string{"new-vs"},
			Deleted: []string{"old-vs"},
		}
		err := printSyncOutcome(&buf, "vector stores", result, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := buf.String()
		want := "Created vector stores: new-vs\n" +
			"Deleted vector stores: old-vs\n"
		if got != want {
			t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("error with partial result", func(t *testing.T) {
		var buf bytes.Buffer
		result := &SyncResult{
			Created: []string{"svc-a"},
		}
		syncErr := fmt.Errorf("failed to create service \"svc-b\"")
		err := printSyncOutcome(&buf, "services", result, syncErr)
		if err != syncErr {
			t.Fatalf("expected original error, got: %v", err)
		}
		got := buf.String()
		want := "Created services (partial): svc-a\n"
		if got != want {
			t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("error with nil result", func(t *testing.T) {
		var buf bytes.Buffer
		syncErr := fmt.Errorf("failed to list services")
		err := printSyncOutcome(&buf, "services", nil, syncErr)
		if err != syncErr {
			t.Fatalf("expected original error, got: %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("expected no output, got: %q", buf.String())
		}
	})
}
