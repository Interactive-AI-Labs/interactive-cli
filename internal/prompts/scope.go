// Package prompts reads one prompt type across the project and global scopes.
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
	Plural       string // names the type in messages: "skills", "routines", …
	Global       bool
	Warn         io.Writer
}

// List returns the project's records followed by the global ones, every row
// labeled with its scope. The scopes are not deduplicated: a name defined in
// both is two rows, and the Copilot runtime is what picks between them.
func (r ScopedReader) List(
	ctx context.Context,
	opts platform.PromptListOptions,
) (*platform.PromptListResponse, error) {
	if !r.Global {
		return r.Client.ListPrompts(ctx, r.ProjectID, r.RouteSegment, opts)
	}

	result, err := r.listWhole(ctx, opts)
	if err != nil {
		return nil, err
	}
	// Global records are project-wide, so no folder holds them.
	if opts.Subfolder != "" {
		return result, nil
	}

	global, err := r.listWhole(ctx, platform.PromptListOptions{Scope: platform.ScopeGlobal})
	if err != nil {
		// A failed global read must not hide the project's own records.
		fmt.Fprintf(r.Warn, "Warning: could not load global %s: %v\n", r.Plural, err)
		return result, nil
	}
	if global.Scope != platform.ScopeGlobal {
		return result, nil
	}

	result.Prompts = mergeScopes(result.Prompts, global.Prompts)
	// The reply now covers both scopes, so only the rows carry one.
	result.Scope = ""
	result.TotalCount = len(result.Prompts)
	return result, nil
}

// mergeScopes labels the rows: a reply carries one scope, the merged listing two.
func mergeScopes(project, global []platform.PromptInfo) []platform.PromptInfo {
	merged := make([]platform.PromptInfo, 0, len(project)+len(global))
	for _, row := range project {
		row.Scope = platform.ScopeProject
		merged = append(merged, row)
	}
	for _, row := range global {
		row.Scope = platform.ScopeGlobal
		merged = append(merged, row)
	}
	return merged
}

// listWhole reads one scope in full: the listing covers two of them, and a page
// of one against a page of the other describes neither.
func (r ScopedReader) listWhole(
	ctx context.Context,
	opts platform.PromptListOptions,
) (*platform.PromptListResponse, error) {
	opts.Limit = platform.PromptScanLimit
	result, err := r.Client.ListPrompts(ctx, r.ProjectID, r.RouteSegment, opts)
	if err != nil {
		return nil, err
	}
	// The limit is far past any real catalogue; say so rather than cut silently.
	if result.TotalCount > len(result.Prompts) {
		fmt.Fprintf(r.Warn, "Warning: showing %d of %d %s\n",
			len(result.Prompts), result.TotalCount, r.Plural)
	}
	return result, nil
}

// Get reads the scope the caller asked for. A name the project does not have is
// not found; the global scope is pointed at rather than searched, so the record
// returned always came from the scope the caller named.
func (r ScopedReader) Get(
	ctx context.Context,
	name string,
	opts platform.PromptGetOptions,
) (*platform.PromptDetail, error) {
	// The project scope is the server's default and predates the parameter.
	if opts.Scope == platform.ScopeProject {
		opts.Scope = ""
	}

	result, err := r.Client.GetPrompt(ctx, r.ProjectID, r.RouteSegment, name, opts)
	if err != nil {
		if suggestsGlobal(r.Global, opts, err) {
			fmt.Fprintf(
				r.Warn,
				"Hint: to look in the global scope, run: iai %s get %s --scope %s\n",
				r.Plural,
				name,
				platform.ScopeGlobal,
			)
		}
		return nil, err
	}
	if opts.Scope == platform.ScopeGlobal && result.Scope != platform.ScopeGlobal {
		return nil, fmt.Errorf(
			"failed to read global %s: the server answered from the project scope",
			r.Plural,
		)
	}
	return result, nil
}

// suggestsGlobal reports whether a failed read should point at the global scope.
// Only a 404 does, and only for a read the global scope could have answered: it
// carries one version of each record, labeled "active".
func suggestsGlobal(global bool, opts platform.PromptGetOptions, err error) bool {
	if !global || opts.Scope == platform.ScopeGlobal {
		return false
	}
	if opts.Version != 0 || (opts.Label != "" && opts.Label != activeLabel) {
		return false
	}
	var notFound *platform.NotFoundError
	return errors.As(err, &notFound)
}
