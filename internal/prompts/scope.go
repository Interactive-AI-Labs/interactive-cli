// Package prompts resolves a prompt read across the project and global scopes.
// The reply echoes the scope it answered from; a server that does not know the
// parameter ignores it and answers from the project, so the echo tells them apart.
package prompts

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

// The global scope exposes one version, labeled "active".
const activeLabel = "active"

// ScopedReader reads one prompt type for one project.
type ScopedReader struct {
	Client       *platform.APIClient
	ProjectID    string
	RouteSegment string
	Plural       string // names the type in warnings: "skills", "routines", …
	Global       bool
	Warn         io.Writer
}

// List returns the project's page with the global records appended when they apply.
func (r ScopedReader) List(
	ctx context.Context,
	opts platform.PromptListOptions,
) (*platform.PromptListResponse, error) {
	result, err := r.Client.ListPrompts(ctx, r.ProjectID, r.RouteSegment, opts)
	if err != nil {
		return nil, err
	}
	if !wantsGlobalRows(r.Global, opts) {
		return result, nil
	}

	global, err := r.Client.ListPrompts(ctx, r.ProjectID, r.RouteSegment,
		platform.PromptListOptions{Scope: platform.ScopeGlobal})
	if err != nil {
		// A failed global read must not hide the project's own records.
		fmt.Fprintf(r.Warn, "Warning: could not load global %s: %v\n", r.Plural, err)
		return result, nil
	}
	if global.Scope != platform.ScopeGlobal {
		return result, nil
	}

	result.Prompts = mergeGlobalRows(result.Prompts, global.Prompts)
	return result, nil
}

// Get falls back to the global scope for a name the project does not have,
// returning the project's own error when there is no global record either.
func (r ScopedReader) Get(
	ctx context.Context,
	name string,
	opts platform.PromptGetOptions,
) (*platform.PromptDetail, error) {
	result, err := r.Client.GetPrompt(ctx, r.ProjectID, r.RouteSegment, name, opts)
	if err == nil {
		return result, nil
	}
	if !canFallBackToGlobal(r.Global, opts, err) {
		return nil, err
	}

	global, globalErr := r.Client.GetPrompt(ctx, r.ProjectID, r.RouteSegment, name,
		platform.PromptGetOptions{Scope: platform.ScopeGlobal})
	if globalErr != nil {
		// Not found here either is the same answer, not a warning.
		var notFound *platform.NotFoundError
		if !errors.As(globalErr, &notFound) {
			fmt.Fprintf(r.Warn, "Warning: could not check global %s: %v\n", r.Plural, globalErr)
		}
		return nil, err
	}
	if global.Scope != platform.ScopeGlobal {
		return nil, err
	}
	return global, nil
}

// Global records are project-wide and unpaginated, so they belong to the root
// listing's first page. Pages are 0-indexed, so anything <= 0 is that page.
func wantsGlobalRows(global bool, opts platform.PromptListOptions) bool {
	return global && opts.Subfolder == "" && opts.Page <= 0
}

// mergeGlobalRows appends the global records the project does not define. A project
// record hides one of the same name; a folder does not, since it reads as "name/".
func mergeGlobalRows(project, global []platform.PromptInfo) []platform.PromptInfo {
	owned := make(map[string]bool, len(project))
	for _, p := range project {
		if p.RowType != platform.RowTypeFolder {
			owned[p.Name] = true
		}
	}
	merged := project
	for _, row := range global {
		if owned[row.Name] {
			continue
		}
		row.Scope = platform.ScopeGlobal
		merged = append(merged, row)
	}
	return merged
}

// canFallBackToGlobal reports whether a failed project lookup should retry globally.
// Only a 404 qualifies; any other failure is the answer.
func canFallBackToGlobal(global bool, opts platform.PromptGetOptions, err error) bool {
	if !global || opts.Version != 0 || (opts.Label != "" && opts.Label != activeLabel) {
		return false
	}
	var notFound *platform.NotFoundError
	return errors.As(err, &notFound)
}
