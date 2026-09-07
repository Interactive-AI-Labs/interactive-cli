package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func filesDeleteTestServer(t *testing.T, requests *[]string) *httptest.Server {
	t.Helper()
	base := "/api/platform/v1/organizations/org-1/projects/proj-1/files"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/session/organizations":
			fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
		case r.URL.Path == "/api/v1/session/organizations/org-1/projects":
			fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"alunafi"}]}`)
		case strings.HasPrefix(r.URL.Path, base):
			*requests = append(*requests, r.Method+" "+r.URL.Path)
			fmt.Fprint(w, `{"success":true,"data":{"id":"f-1","versionId":null}}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func setupFilesDeleteTest(t *testing.T, server *httptest.Server, stdin string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	origHostname, origToken, origApiKey := hostname, token, apiKey
	t.Cleanup(func() {
		hostname, token, apiKey = origHostname, origToken, origApiKey
		filesDeleteOrg, filesDeleteProject = "", ""
		filesDeleteVersion, filesDeleteForce = "", false
		for _, name := range []string{"version", "force"} {
			filesDeleteCmd.Flags().Lookup(name).Changed = false
		}
		filesDeleteCmd.SetIn(nil)
		filesDeleteCmd.SetOut(nil)
	})
	hostname, token, apiKey = server.URL, "test-token", ""
	filesDeleteOrg, filesDeleteProject = "acme", "alunafi"
	filesDeleteCmd.SetIn(strings.NewReader(stdin))
	filesDeleteCmd.SetOut(new(strings.Builder))
	filesDeleteCmd.SetContext(context.Background())
}

func TestFilesDelete_VersionWithClosedStdin(t *testing.T) {
	var requests []string
	server := filesDeleteTestServer(t, &requests)
	setupFilesDeleteTest(t, server, "")

	if err := filesDeleteCmd.Flags().Set("version", "v-1"); err != nil {
		t.Fatalf("set --version: %v", err)
	}

	if err := filesDeleteCmd.RunE(filesDeleteCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files delete --version (closed stdin): %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("requests = %v, want exactly 1", requests)
	}
	if requests[0] != "DELETE /api/platform/v1/organizations/org-1/projects/proj-1/files/f-1/versions/v-1" {
		t.Errorf("request = %q, want a versioned DELETE", requests[0])
	}
}

func TestFilesDelete_NoVersionAnswerNo(t *testing.T) {
	var requests []string
	server := filesDeleteTestServer(t, &requests)
	setupFilesDeleteTest(t, server, "n\n")

	if err := filesDeleteCmd.RunE(filesDeleteCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files delete (answer n): %v", err)
	}

	if len(requests) != 0 {
		t.Fatalf("requests = %v, want none", requests)
	}
}

func TestFilesDelete_NoVersionAnswerYes(t *testing.T) {
	var requests []string
	server := filesDeleteTestServer(t, &requests)
	setupFilesDeleteTest(t, server, "y\n")

	if err := filesDeleteCmd.RunE(filesDeleteCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files delete (answer y): %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("requests = %v, want exactly 1", requests)
	}
	if requests[0] != "DELETE /api/platform/v1/organizations/org-1/projects/proj-1/files/f-1" {
		t.Errorf("request = %q, want a whole-file DELETE", requests[0])
	}
}

func TestFilesDelete_Force(t *testing.T) {
	var requests []string
	server := filesDeleteTestServer(t, &requests)
	setupFilesDeleteTest(t, server, "")

	if err := filesDeleteCmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("set --force: %v", err)
	}

	if err := filesDeleteCmd.RunE(filesDeleteCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files delete -f: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("requests = %v, want exactly 1", requests)
	}
	if requests[0] != "DELETE /api/platform/v1/organizations/org-1/projects/proj-1/files/f-1" {
		t.Errorf("request = %q, want a whole-file DELETE", requests[0])
	}
}
