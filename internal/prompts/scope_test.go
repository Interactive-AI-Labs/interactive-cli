package prompts

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

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
