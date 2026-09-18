package prompts

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
)

func TestMergeGlobalRows(t *testing.T) {
	tests := []struct {
		name    string
		project []platform.PromptInfo
		global  []platform.PromptInfo
		want    []platform.PromptInfo
	}{
		{
			name:    "global records are appended and tagged",
			project: []platform.PromptInfo{{Name: "summarize"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "summarize"},
				{Name: "routines", Scope: "global"},
			},
		},
		{
			name:    "a project record hides the global one",
			project: []platform.PromptInfo{{Name: "routines"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want:    []platform.PromptInfo{{Name: "routines"}},
		},
		{
			name:    "a folder does not hide the global record of that name",
			project: []platform.PromptInfo{{Name: "routines", RowType: "folder"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "routines", RowType: "folder"},
				{Name: "routines", Scope: "global"},
			},
		},
		{
			name:    "an unrelated folder is left alone",
			project: []platform.PromptInfo{{Name: "team", RowType: "folder"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "team", RowType: "folder"},
				{Name: "routines", Scope: "global"},
			},
		},
		{
			name:    "no global records leaves the listing untouched",
			project: []platform.PromptInfo{{Name: "summarize"}},
			want:    []platform.PromptInfo{{Name: "summarize"}},
		},
		{
			name:   "an empty project still lists the global records",
			global: []platform.PromptInfo{{Name: "routines"}},
			want:   []platform.PromptInfo{{Name: "routines", Scope: "global"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, MergeGlobalRows(tt.project, tt.global)); diff != "" {
				t.Errorf("rows mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCanFallBackToGlobal(t *testing.T) {
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
			name:   "the active label resolves",
			global: true,
			opts:   platform.PromptGetOptions{Label: "active"},
			err:    notFound,
			want:   true,
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
		{name: "types without a global scope never fall back", err: notFound},
		{name: "a transport failure is the answer", global: true, err: errors.New("refused")},
		{
			name:   "a wrapped 404 qualifies",
			global: true,
			err:    fmt.Errorf("failed to get skill: %w", notFound),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanFallBackToGlobal(tt.global, tt.opts, tt.err); got != tt.want {
				t.Errorf("CanFallBackToGlobal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWantsGlobalRows(t *testing.T) {
	tests := []struct {
		name   string
		global bool
		opts   platform.PromptListOptions
		want   bool
	}{
		{name: "root listing, first page", global: true, want: true},
		{
			name:   "a negative page is still the first",
			global: true,
			opts:   platform.PromptListOptions{Page: -1},
			want:   true,
		},
		{name: "a later page", global: true, opts: platform.PromptListOptions{Page: 1}},
		{
			name:   "inside a folder",
			global: true,
			opts:   platform.PromptListOptions{Subfolder: "team"},
		},
		{name: "types without a global scope"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wantsGlobalRows(tt.global, tt.opts); got != tt.want {
				t.Errorf("wantsGlobalRows() = %v, want %v", got, tt.want)
			}
		})
	}
}
