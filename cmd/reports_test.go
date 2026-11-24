package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

// Only the Copilot reads the open draft, and prints the revision to save it at.
func TestReportsGetReadsTheDraftInTheCopilot(t *testing.T) {
	const reports = "/api/platform/v1/projects/proj-1/reports"
	tests := []struct {
		origin, wantQuery, wantName string
		wantRevision                bool
	}{
		{inputs.ReportsViaCopilot, "include=draft", "name: Traffic v2", true},
		{inputs.ReportsViaCLI, "", "name: Traffic", false},
	}
	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			var query string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if writeFilesTestSessionResponse(w, r) {
					return
				}
				switch r.URL.Path {
				case reports:
					fmt.Fprint(w, `{"success":true,"data":[{"id":"r-1","name":"Traffic"}]}`)
				case reports + "/r-1":
					query = r.URL.RawQuery
					draft := "null"
					if query == "include=draft" {
						draft = `{"revision":3,"document":{"name":"Traffic v2","definition":{}}}`
					}
					fmt.Fprintf(w, `{"success":true,"data":{"id":"r-1","name":"Traffic",`+
						`"definition":{},"draft":%s}}`, draft)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			out := setupFilesCommandTest(t, server, reportsGetCmd)
			origOrigin := clientOrigin
			t.Cleanup(func() {
				clientOrigin = origOrigin
				reportsOrganization, reportsProject, reportsGetYAML = "", "", false
			})
			clientOrigin = tt.origin
			reportsOrganization, reportsProject, reportsGetYAML = "acme", "alunafi", true

			if err := reportsGetCmd.RunE(reportsGetCmd, []string{"Traffic"}); err != nil {
				t.Fatal(err)
			}
			got := out.String()
			if query != tt.wantQuery || !strings.Contains(got, tt.wantName+"\n") ||
				strings.Contains(got, "draft_revision: 3") != tt.wantRevision {
				t.Errorf("query %q, output:\n%s", query, got)
			}
		})
	}
}
