package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// deleteRun is everything a delete command did: its error, the requests it sent
// past session lookup, and what it printed.
type deleteRun struct {
	err      string
	requests string
	out      string
}

func runDeleteCommand(
	t *testing.T,
	command *cobra.Command,
	stdin string,
	args ...string,
) deleteRun {
	t.Helper()
	requests := new(filesRequestLog)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeFilesTestSessionResponse(w, r) {
			return
		}
		entry := r.Method + " " + r.URL.EscapedPath()
		if r.URL.RawQuery != "" {
			entry += "?" + r.URL.RawQuery
		}
		requests.add(entry)
		fmt.Fprint(w, `{"message":"deleted"}`)
	}))
	t.Cleanup(server.Close)

	origDeployHostname := deploymentHostname
	t.Cleanup(func() { deploymentHostname = origDeployHostname })
	deploymentHostname = server.URL

	out := setupFilesCommandTest(t, server, command)
	command.SetIn(strings.NewReader(stdin))

	var run deleteRun
	err := command.ParseFlags(append(args, "-o", "acme", "-p", "alunafi"))
	if err == nil {
		err = command.ValidateFlagGroups()
	}
	if err == nil {
		err = command.RunE(command, command.Flags().Args())
	}
	if err != nil {
		run.err = err.Error()
	}
	run.requests = strings.Join(requests.all(), ", ")
	run.out = out.String()
	return run
}

func TestPromptDelete(t *testing.T) {
	const routines = "DELETE /api/platform/v1/projects/proj-1/prompts/routines/"
	scopedRefusal := func(name, segment string) string {
		return fmt.Sprintf(
			"cannot delete by --version or --label: %q contains the path segment %q, "+
				"which the server would treat as a wider delete", name, segment,
		)
	}

	tests := []struct {
		name    string
		generic bool
		stdin   string
		args    []string
		want    deleteRun
	}{
		{
			name:  "version delete asks first",
			stdin: "n\n",
			args:  []string{"x", "--version", "3"},
			want: deleteRun{
				out: `This will delete version 3 of routine "x". Continue? [y/N] Aborted.` + "\n",
			},
		},
		{
			name:  "label delete asks first",
			stdin: "y\n",
			args:  []string{"x", "--label", "prod"},
			want: deleteRun{
				requests: routines + "x?label=prod",
				out: `This will delete versions labelled "prod" of routine "x". Continue? [y/N] ` +
					`Deleted versions labelled "prod" of routine "x".` + "\n",
			},
		},
		{
			name:  "whole delete asks first",
			stdin: "y\n",
			args:  []string{"x"},
			want: deleteRun{
				requests: routines + "by-name/x",
				out: `This will delete routine "x" and all its versions. Continue? [y/N] ` +
					`Deleted routine "x" and all its versions.` + "\n",
			},
		},
		{
			name: "-f skips the prompt",
			args: []string{"x", "--version", "3", "-f"},
			want: deleteRun{
				requests: routines + "x?version=3",
				out:      `Deleted version 3 of routine "x".` + "\n",
			},
		},
		{
			name:    "generic prompts share the flow",
			generic: true,
			args:    []string{"x", "--version", "2", "-f"},
			want: deleteRun{
				requests: "DELETE /api/platform/v1/projects/proj-1/prompts/x?version=2",
				out:      `Deleted version 2 of prompt "x".` + "\n",
			},
		},
		{
			name: "version 0",
			args: []string{"x", "--version", "0", "-f"},
			want: deleteRun{err: "--version must be 1 or greater"},
		},
		{
			name: "negative version",
			args: []string{"x", "--version", "-1", "-f"},
			want: deleteRun{err: "--version must be 1 or greater"},
		},
		{
			name: "empty label",
			args: []string{"x", "--label", "", "-f"},
			want: deleteRun{err: "--label must not be empty"},
		},
		{
			name: "blank label",
			args: []string{"x", "--label", " ", "-f"},
			want: deleteRun{err: "--label must not be empty"},
		},
		{
			name: "version and label together",
			args: []string{"x", "--version", "3", "--label", "prod", "-f"},
			want: deleteRun{
				err: "if any flags in the group [version label] are set none of the others can be; [label version] were all set",
			},
		},
		{
			name: "folders segment",
			args: []string{"folders/x", "--version", "2", "-f"},
			want: deleteRun{err: scopedRefusal("folders/x", "folders")},
		},
		{
			name: "by-name segment",
			args: []string{"a/by-name", "--label", "prod", "-f"},
			want: deleteRun{err: scopedRefusal("a/by-name", "by-name")},
		},
		{
			name:    "versions segment on generic prompts",
			generic: true,
			args:    []string{"a/versions/b", "--version", "1", "-f"},
			want:    deleteRun{err: scopedRefusal("a/versions/b", "versions")},
		},
		{
			name: "segment-like names are fine",
			args: []string{"folders-x", "--version", "1", "-f"},
			want: deleteRun{
				requests: routines + "folders-x?version=1",
				out:      `Deleted version 1 of routine "folders-x".` + "\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := makeDeleteCmd(
				PromptTypeConfig{TypeName: "routine", RouteSegment: "routines"},
			)
			if tt.generic {
				command = makeGenericDeleteCmd()
			}
			if got := runDeleteCommand(t, command, tt.stdin, tt.args...); got != tt.want {
				t.Errorf("got  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}
