package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func filesDeleteTestServer(t *testing.T, requests *filesRequestLog) *httptest.Server {
	t.Helper()
	base := "/api/platform/v1/organizations/org-1/projects/proj-1/files"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeFilesTestSessionResponse(w, r) {
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, base):
			requests.add(r.Method + " " + r.URL.Path)
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
	setupFilesCommandTest(t, server, filesDeleteCmd)
	t.Cleanup(func() {
		filesDeleteOrg, filesDeleteProject = "", ""
		filesDeleteVersion, filesDeleteForce = "", false
		for _, name := range []string{"version", "force"} {
			filesDeleteCmd.Flags().Lookup(name).Changed = false
		}
	})
	filesDeleteOrg, filesDeleteProject = "acme", "alunafi"
	filesDeleteCmd.SetIn(strings.NewReader(stdin))
}

func TestFilesDelete_Modes(t *testing.T) {
	const filesPath = "/api/platform/v1/organizations/org-1/projects/proj-1/files/f-1"
	tests := []struct {
		name        string
		stdin       string
		version     string
		force       bool
		wantRequest string
	}{
		{
			name:        "version with closed stdin",
			version:     "v-1",
			wantRequest: "DELETE " + filesPath + "/versions/v-1",
		},
		{name: "whole file declined", stdin: "n\n"},
		{name: "whole file confirmed", stdin: "y\n", wantRequest: "DELETE " + filesPath},
		{name: "whole file forced", force: true, wantRequest: "DELETE " + filesPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := new(filesRequestLog)
			server := filesDeleteTestServer(t, requests)
			setupFilesDeleteTest(t, server, tt.stdin)

			if tt.version != "" {
				if err := filesDeleteCmd.Flags().Set("version", tt.version); err != nil {
					t.Fatalf("set --version: %v", err)
				}
			}
			if tt.force {
				if err := filesDeleteCmd.Flags().Set("force", "true"); err != nil {
					t.Fatalf("set --force: %v", err)
				}
			}

			if err := filesDeleteCmd.RunE(filesDeleteCmd, []string{"f-1"}); err != nil {
				t.Fatalf("files delete: %v", err)
			}

			got := requests.all()
			if tt.wantRequest == "" {
				if len(got) != 0 {
					t.Fatalf("requests = %v, want none", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("requests = %v, want exactly 1", got)
			}
			if got[0] != tt.wantRequest {
				t.Errorf("request = %q, want %q", got[0], tt.wantRequest)
			}
		})
	}
}
