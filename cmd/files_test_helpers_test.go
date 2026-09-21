package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func writeFilesTestSessionResponse(w http.ResponseWriter, r *http.Request) bool {
	switch r.URL.Path {
	case "/api/v1/session/organizations":
		fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
	case "/api/v1/session/organizations/org-1/projects":
		fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"alunafi"}]}`)
	default:
		return false
	}
	return true
}

func setupFilesCommandTest(
	t *testing.T,
	server *httptest.Server,
	command *cobra.Command,
) *strings.Builder {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	origHostname, origToken, origAPIKey := hostname, token, apiKey
	t.Cleanup(func() {
		hostname, token, apiKey = origHostname, origToken, origAPIKey
		command.SetIn(nil)
		command.SetOut(nil)
		command.SetErr(nil)
	})

	hostname, token, apiKey = server.URL, "test-token", ""
	out := new(strings.Builder)
	command.SetIn(nil)
	command.SetOut(out)
	command.SetErr(new(strings.Builder))
	command.SetContext(context.Background())
	return out
}
