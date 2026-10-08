package files

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPlanResources(t *testing.T) {
	// The server's answer for this table: the stored form is the config upper-cased, and it
	// changes when the config differs from the live one.
	live := map[string]string{"api": "A", "web": "W", "old": "O"}
	plan := func(name, cfg string) (string, bool, error) {
		if cfg == "boom" {
			return "", false, errors.New("boom")
		}
		return cfg + "!", cfg != live[name], nil
	}
	tests := []struct {
		name    string
		local   map[string]string
		want    ResourcePlan[string]
		wantErr string
	}{
		{
			name:  "no resources",
			local: map[string]string{},
			want:  ResourcePlan[string]{Desired: map[string]string{}, Changed: map[string]bool{}},
		},
		{
			name:  "created resources are kept as written",
			local: map[string]string{"new": "n"},
			want: ResourcePlan[string]{
				Desired: map[string]string{"new": "n"},
				Changed: map[string]bool{},
			},
		},
		{
			name:  "existing resources take the server's form and changed flag",
			local: map[string]string{"api": "A", "web": "x"},
			want: ResourcePlan[string]{
				Desired: map[string]string{"api": "A!", "web": "x!"},
				Changed: map[string]bool{"web": true},
			},
		},
		{
			name:    "a failed plan stops",
			local:   map[string]string{"api": "boom"},
			wantErr: "boom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := planResources(tt.local, live, plan)
			if (err != nil && err.Error() != tt.wantErr) || (err == nil && tt.wantErr != "") {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("plan mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSelfHostedMcp(t *testing.T) {
	tests := []struct {
		mcpType string
		want    bool
	}{
		{mcpType: "", want: true},
		{mcpType: "self-hosted", want: true},
		{mcpType: "internal", want: true},
		{mcpType: " self-hosted ", want: true},
		{mcpType: "remote", want: false},
		{mcpType: "external", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.mcpType, func(t *testing.T) {
			if got := selfHostedMcp(tt.mcpType); got != tt.want {
				t.Fatalf("selfHostedMcp(%q) = %v, want %v", tt.mcpType, got, tt.want)
			}
		})
	}
}
