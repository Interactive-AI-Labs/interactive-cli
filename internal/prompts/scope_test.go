package prompts

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
)

func TestMergeScopes(t *testing.T) {
	tests := []struct {
		name    string
		project []platform.PromptInfo
		global  []platform.PromptInfo
		want    []platform.PromptInfo
	}{
		{
			name:    "a name held by both scopes stays two rows",
			project: []platform.PromptInfo{{Name: "routines"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "routines", Scope: "project"},
				{Name: "routines", Scope: "global"},
			},
		},
		{
			name:    "a folder is labeled like any other row",
			project: []platform.PromptInfo{{Name: "team", RowType: "folder"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "team", RowType: "folder", Scope: "project"},
				{Name: "routines", Scope: "global"},
			},
		},
		{
			name:   "an empty project still lists the global rows",
			global: []platform.PromptInfo{{Name: "routines"}, {Name: "glossaries"}},
			want: []platform.PromptInfo{
				{Name: "routines", Scope: "global"},
				{Name: "glossaries", Scope: "global"},
			},
		},
		{
			name: "no global rows leaves the project's own",
			project: []platform.PromptInfo{
				{Name: "summarize", Versions: []int{1, 2}, Labels: []string{"active"}},
			},
			want: []platform.PromptInfo{
				{
					Name:     "summarize",
					Versions: []int{1, 2},
					Labels:   []string{"active"},
					Scope:    "project",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, mergeScopes(tt.project, tt.global)); diff != "" {
				t.Errorf("rows mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSuggestsGlobal(t *testing.T) {
	notFound := &platform.NotFoundError{Message: "no such prompt"}

	tests := []struct {
		name   string
		global bool
		opts   platform.PromptGetOptions
		err    error
		want   bool
	}{
		{name: "404 on a global-scoped type", global: true, err: notFound, want: true},
		{
			name:   "a wrapped 404 qualifies",
			global: true,
			err:    fmt.Errorf("failed to get skill: %w", notFound),
			want:   true,
		},
		{
			name:   "the active label resolves globally",
			global: true,
			opts:   platform.PromptGetOptions{Label: "active"},
			err:    notFound,
			want:   true,
		},
		{
			name:   "the project scope, asked for by name",
			global: true,
			opts:   platform.PromptGetOptions{Scope: "project"},
			err:    notFound,
			want:   true,
		},
		{
			name:   "the global scope was the one that missed",
			global: true,
			opts:   platform.PromptGetOptions{Scope: "global"},
			err:    notFound,
		},
		{
			name:   "another label cannot",
			global: true,
			opts:   platform.PromptGetOptions{Label: "staging"},
			err:    notFound,
		},
		{
			name:   "a pinned version belongs to the project",
			global: true,
			opts:   platform.PromptGetOptions{Version: 3},
			err:    notFound,
		},
		{name: "types without a global scope never suggest it", err: notFound},
		{name: "a transport failure is the answer", global: true, err: errors.New("refused")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := suggestsGlobal(tt.global, tt.opts, tt.err); got != tt.want {
				t.Errorf("suggestsGlobal() = %v, want %v", got, tt.want)
			}
		})
	}
}
