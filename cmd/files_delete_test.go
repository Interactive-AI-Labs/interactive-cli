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
		flags       []string
		wantErr     string
		wantRequest string
	}{
		{
			name:        "version with closed stdin",
			flags:       []string{"--version", "v-1"},
			wantRequest: "DELETE " + filesPath + "/versions/v-1",
		},
		{
			name:    "empty version",
			flags:   []string{"--version", ""},
			wantErr: "--version must not be empty",
		},
		{
			name:    "blank version is not a whole-file delete",
			stdin:   "y\n",
			flags:   []string{"--version", "  ", "-f"},
			wantErr: "--version must not be empty",
		},
		{name: "whole file declined", stdin: "n\n"},
		{name: "whole file confirmed", stdin: "y\n", wantRequest: "DELETE " + filesPath},
		{name: "whole file forced", flags: []string{"-f"}, wantRequest: "DELETE " + filesPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := new(filesRequestLog)
			server := filesDeleteTestServer(t, requests)
			setupFilesDeleteTest(t, server, tt.stdin)

			if err := filesDeleteCmd.ParseFlags(tt.flags); err != nil {
				t.Fatalf("parse flags: %v", err)
			}

			gotErr := ""
			if err := filesDeleteCmd.RunE(filesDeleteCmd, []string{"f-1"}); err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Fatalf("error = %q, want %q", gotErr, tt.wantErr)
			}

			got := strings.Join(requests.all(), ", ")
			if got != tt.wantRequest {
				t.Errorf("requests = %q, want %q", got, tt.wantRequest)
			}
		})
	}
}
