package sync

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
			resource: "databases",
			want:     false,
		},
		{
			name:     "empty list",
			allowed:  []string{},
			resource: "databases",
			want:     false,
		},
		{
			name:     "exact match",
			allowed:  []string{"databases"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "case insensitive match",
			allowed:  []string{"Databases"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "all keyword",
			allowed:  []string{"all"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "ALL keyword uppercase",
			allowed:  []string{"ALL"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "no match",
			allowed:  []string{"services"},
			resource: "databases",
			want:     false,
		},
		{
			name:     "multiple entries with match",
			allowed:  []string{"services", "databases"},
			resource: "databases",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllowDeleteResource(tt.allowed, tt.resource)
			if got != tt.want {
				t.Errorf(
					"AllowDeleteResource(%v, %q) = %v, want %v",
					tt.allowed, tt.resource, got, tt.want,
				)
			}
		})
	}
}

func TestPrintResult(t *testing.T) {
	tests := []struct {
		name    string
		label   string
		result  *Result
		syncErr error
		want    string
		wantErr bool
	}{
		{
			name:  "created and deleted",
			label: "databases",
			result: &Result{
				Created: []string{"new-db"},
				Deleted: []string{"old-db"},
			},
			want: "Created databases: new-db\n" +
				"Deleted databases: old-db\n",
		},
		{
			name:  "updated items",
			label: "services",
			result: &Result{
				Updated: []string{"svc-a"},
			},
			want: "Updated services: svc-a\n",
		},
		{
			name:   "no changes",
			label:  "services",
			result: &Result{},
			want:   "No changes required; services already match config.\n",
		},
		{
			name:  "multiple items joined with comma",
			label: "services",
			result: &Result{
				Created: []string{"svc-a", "svc-b"},
				Deleted: []string{"svc-c", "svc-d"},
			},
			want: "Created services: svc-a, svc-b\n" +
				"Deleted services: svc-c, svc-d\n",
		},
		{
			name:  "protected items print warning",
			label: "databases",
			result: &Result{
				Created:   []string{"new-db"},
				Protected: []string{"old-db"},
			},
			want: "Created databases: new-db\n" +
				"\nProtected databases (not deleted): old-db\n" +
				"Use --allow-delete=databases to delete them.\n",
		},
		{
			name:    "error with partial result",
			label:   "services",
			result:  &Result{Created: []string{"svc-a"}},
			syncErr: fmt.Errorf("failed to create service \"svc-b\""),
			wantErr: true,
			want:    "Created services (partial): svc-a\n",
		},
		{
			name:    "error with nil result",
			label:   "services",
			syncErr: fmt.Errorf("failed to list services"),
			wantErr: true,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintResult(&buf, tt.label, tt.result, tt.syncErr)
			if tt.wantErr && err != tt.syncErr {
				t.Fatalf("expected original error, got: %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
