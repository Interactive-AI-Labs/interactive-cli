package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
)

func TestMergeGlobalRows(t *testing.T) {
	tests := []struct {
		name           string
		project        []platform.PromptInfo
		global         []platform.PromptInfo
		totalCount     int
		want           []platform.PromptInfo
		wantTotalCount int
	}{
		{
			name:       "global rows are appended and marked",
			project:    []platform.PromptInfo{{Name: "own"}},
			global:     []platform.PromptInfo{{Name: "shared"}},
			totalCount: 1,
			want: []platform.PromptInfo{
				{Name: "own"},
				{Name: "shared", Source: "general"},
			},
			wantTotalCount: 2,
		},
		{
			// The Copilot loads the project's version over the shared one, so a
			// listing that showed both would misreport what runs.
			name:           "a project name shadows the shared one",
			project:        []platform.PromptInfo{{Name: "routines"}},
			global:         []platform.PromptInfo{{Name: "routines"}},
			totalCount:     1,
			want:           []platform.PromptInfo{{Name: "routines"}},
			wantTotalCount: 1,
		},
		{
			name:       "folders never collide with shared names",
			project:    []platform.PromptInfo{{Name: "team", RowType: "folder"}},
			global:     []platform.PromptInfo{{Name: "shared"}},
			totalCount: 1,
			want: []platform.PromptInfo{
				{Name: "team", RowType: "folder"},
				{Name: "shared", Source: "general"},
			},
			wantTotalCount: 2,
		},
		{
			name:           "no shared rows leaves the listing untouched",
			project:        []platform.PromptInfo{{Name: "own"}},
			global:         nil,
			totalCount:     1,
			want:           []platform.PromptInfo{{Name: "own"}},
			wantTotalCount: 1,
		},
		{
			name:           "an empty project still lists the shared rows",
			project:        nil,
			global:         []platform.PromptInfo{{Name: "shared"}},
			totalCount:     0,
			want:           []platform.PromptInfo{{Name: "shared", Source: "general"}},
			wantTotalCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotTotal := mergeGlobalRows(tt.project, tt.global, tt.totalCount)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("rows mismatch (-want +got):\n%s", diff)
			}
			if gotTotal != tt.wantTotalCount {
				t.Errorf("totalCount = %d, want %d", gotTotal, tt.wantTotalCount)
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
