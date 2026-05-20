package versioncheck

import (
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"0.28.1", "0.29.0", true},
		{"0.28.1", "0.28.2", true},
		{"0.28.1", "1.0.0", true},
		{"0.28.1", "0.28.1", false},
		{"0.28.1", "0.28.0", false},
		{"1.0.0", "0.99.99", false},
		{"0.28.0", "0.28.1-beta.1", false},
		{"0.28.1-beta.1", "0.28.1", true},
		{"0.28.1", "0.28.1-beta.1", false},
		{"0.28.1-beta.1", "0.28.2-beta.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.current+"_vs_"+tt.latest, func(t *testing.T) {
			if got := IsNewer(tt.current, tt.latest); got != tt.want {
				t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input string
		want  []int
	}{
		{"1.2.3", []int{1, 2, 3}},
		{"v1.2.3", []int{1, 2, 3}},
		{"0.28.1-beta.1", []int{0, 28, 1}},
		{"1.21", []int{1, 21, 0}},
		{"invalid", nil},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSemver(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseSemver(%q) = %v, want nil", tt.input, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("parseSemver(%q) = nil, want %v", tt.input, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("parseSemver(%q)[%d] = %d, want %d", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}
