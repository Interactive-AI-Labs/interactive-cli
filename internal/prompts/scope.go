// Package prompts reads one prompt type from the scope the caller named.
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
	Plural       string // names the type in the hint: "skills", "routines", …
	Global       bool
	Warn         io.Writer
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
