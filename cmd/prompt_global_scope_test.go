package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
)

// What the listing shows has to match what the Copilot loads: a project skill
// hides the global one of the same name, a folder does not.
func TestMergeGlobalRows(t *testing.T) {
	tests := []struct {
		name    string
		project []platform.PromptInfo
		global  []platform.PromptInfo
		want    []platform.PromptInfo
	}{
		{
			name:    "global skills are appended and tagged",
			project: []platform.PromptInfo{{Name: "summarize"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "summarize"},
				{Name: "routines", Scope: "global"},
			},
		},
		{
			name:    "a project skill hides the global one",
			project: []platform.PromptInfo{{Name: "routines"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want:    []platform.PromptInfo{{Name: "routines"}},
		},
		{
			// A folder already reads as "routines/", so hiding the skill would lose it.
			name:    "a folder does not hide the global skill of that name",
			project: []platform.PromptInfo{{Name: "routines", RowType: "folder"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "routines", RowType: "folder"},
				{Name: "routines", Scope: "global"},
			},
		},
		{
			// What a server without scope=global returns: the project's page again.
			name:    "a server echoing the project page adds nothing",
			project: []platform.PromptInfo{{Name: "summarize"}, {Name: "triage"}},
			global:  []platform.PromptInfo{{Name: "summarize"}, {Name: "triage"}},
			want:    []platform.PromptInfo{{Name: "summarize"}, {Name: "triage"}},
		},
		{
			name:    "no global skills leaves the listing untouched",
			project: []platform.PromptInfo{{Name: "summarize"}},
			want:    []platform.PromptInfo{{Name: "summarize"}},
		},
		{
			name:   "an empty project still lists the global skills",
			global: []platform.PromptInfo{{Name: "routines"}},
			want:   []platform.PromptInfo{{Name: "routines", Scope: "global"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, mergeGlobalRows(tt.project, tt.global)); diff != "" {
				t.Errorf("rows mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// Only a 404 may fall back, and only for a version the global scope can serve.
func TestCanFallBackToGlobal(t *testing.T) {
	notFound := &platform.NotFoundError{Message: "no such prompt"}

	tests := []struct {
		name        string
		globalScope bool
		version     int
		label       string
		err         error
		want        bool
	}{
		{name: "404 on a global-scoped type", globalScope: true, err: notFound, want: true},
		{
			name:        "the active label resolves",
			globalScope: true,
			label:       "active",
			err:         notFound,
			want:        true,
		},
		{name: "another label cannot", globalScope: true, label: "staging", err: notFound},
		{
			name:        "a pinned version belongs to the project",
			globalScope: true,
			version:     3,
			err:         notFound,
		},
		{name: "types without a global scope never fall back", err: notFound},
		{name: "a transport failure is the answer", globalScope: true, err: errors.New("refused")},
		{
			// The command layer wraps, so errors.As has to see through it.
			name:        "a wrapped 404 qualifies",
			globalScope: true,
			err:         fmt.Errorf("failed to get skill: %w", notFound),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptCfg := PromptTypeConfig{GlobalScope: tt.globalScope}
			if got := canFallBackToGlobal(ptCfg, tt.version, tt.label, tt.err); got != tt.want {
				t.Errorf("canFallBackToGlobal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A server ignoring scope=global answers from the project instead; tagging that
// reply "global" would claim a skill exists that does not.
func TestServedFromProjectScope(t *testing.T) {
	tests := []struct {
		name   string
		detail platform.PromptDetail
		want   bool
	}{
		{name: "a global skill has neither", detail: platform.PromptDetail{Name: "routines"}},
		{
			name:   "a version means the project answered",
			detail: platform.PromptDetail{Name: "routines", Version: 4},
			want:   true,
		},
		{
			name:   "an id means the project answered",
			detail: platform.PromptDetail{Name: "routines", Id: "cm123"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := servedFromProjectScope(&tt.detail); got != tt.want {
				t.Errorf("servedFromProjectScope() = %v, want %v", got, tt.want)
			}
		})
	}
}
