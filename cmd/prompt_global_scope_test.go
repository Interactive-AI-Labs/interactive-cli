package cmd

import (
	"encoding/json"
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
			name:    "global rows are appended and marked",
			project: []platform.PromptInfo{{Name: "own"}},
			global:  []platform.PromptInfo{{Name: "shared"}},
			want: []platform.PromptInfo{
				{Name: "own"},
				{Name: "shared", Source: "general"},
			},
		},
		{
			// The Copilot loads the project's version over the shared one, so a
			// listing that showed both would misreport what runs.
			name:    "a project name shadows the shared one",
			project: []platform.PromptInfo{{Name: "routines"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want:    []platform.PromptInfo{{Name: "routines"}},
		},
		{
			// The case the folder exclusion exists for: a folder and a skill of the
			// same name are different things, the folder already renders as
			// "routines/", and letting it shadow would hide a skill Copilot loads.
			name:    "a folder does not shadow a shared skill of the same name",
			project: []platform.PromptInfo{{Name: "routines", RowType: "folder"}},
			global:  []platform.PromptInfo{{Name: "routines"}},
			want: []platform.PromptInfo{
				{Name: "routines", RowType: "folder"},
				{Name: "routines", Source: "general"},
			},
		},
		{
			name:    "an unrelated folder is left alone",
			project: []platform.PromptInfo{{Name: "team", RowType: "folder"}},
			global:  []platform.PromptInfo{{Name: "shared"}},
			want: []platform.PromptInfo{
				{Name: "team", RowType: "folder"},
				{Name: "shared", Source: "general"},
			},
		},
		{
			// A server that does not serve the shared scope ignores the parameter and
			// answers with the project's own page. Both reads send the same --limit
			// so the pages match, and the merge has to erase the echo completely
			// rather than relabel it "general".
			name:    "a server echoing the project page adds nothing",
			project: []platform.PromptInfo{{Name: "own"}, {Name: "other"}},
			global:  []platform.PromptInfo{{Name: "own"}, {Name: "other"}},
			want:    []platform.PromptInfo{{Name: "own"}, {Name: "other"}},
		},
		{
			name:    "no shared rows leaves the listing untouched",
			project: []platform.PromptInfo{{Name: "own"}},
			global:  nil,
			want:    []platform.PromptInfo{{Name: "own"}},
		},
		{
			name:    "an empty project still lists the shared rows",
			project: nil,
			global:  []platform.PromptInfo{{Name: "shared"}},
			want:    []platform.PromptInfo{{Name: "shared", Source: "general"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeGlobalRows(tt.project, tt.global)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("rows mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

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
			// `skills get --label active` is what the help text recommends, and a
			// global record only ever has that version, so it must still resolve.
			name:        "the active label still resolves",
			globalScope: true,
			label:       "active",
			err:         notFound,
			want:        true,
		},
		{
			name:        "another label cannot be satisfied",
			globalScope: true,
			label:       "staging",
			err:         notFound,
		},
		{
			name:        "a pinned version belongs to the project",
			globalScope: true,
			version:     3,
			err:         notFound,
		},
		{name: "types without a global scope never fall back", err: notFound},
		{
			name:        "a transport failure is the answer",
			globalScope: true,
			err:         errors.New("connection refused"),
		},
		{
			// errors.As has to see through the wrapping the command layer adds.
			name: "a wrapped 404 still qualifies", globalScope: true,
			err: fmt.Errorf("failed to get skill: %w", notFound), want: true,
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

func TestServedFromProjectScope(t *testing.T) {
	tests := []struct {
		name   string
		detail platform.PromptDetail
		want   bool
	}{
		{
			name: "a global record has neither a version nor an id",
			detail: platform.PromptDetail{
				Name:   "routines",
				Labels: []string{"active"},
				Tags:   []string{"copilot"},
				Prompt: json.RawMessage(`"# Routines"`),
			},
		},
		{
			// A server that does not serve the shared scope ignores the parameter
			// and answers with the project's own record. Relabelling that "general"
			// would tell the caller a shared skill exists when none does.
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
