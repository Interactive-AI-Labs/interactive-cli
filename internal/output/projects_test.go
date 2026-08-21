package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

func TestPrintProjectList(t *testing.T) {
	tests := []struct {
		name            string
		projects        []api.Project
		selectedProject string
		want            string
	}{
		{
			name:     "empty list prints leading newline and headers",
			projects: []api.Project{},
			want:     "\nNAME   ROLE\n",
		},
		{
			name: "single project no selection",
			projects: []api.Project{
				{Name: "my-project", Role: "admin"},
			},
			selectedProject: "",
			want: "\nNAME         ROLE\n" +
				"my-project   admin\n",
		},
		{
			name: "selected project gets marker",
			projects: []api.Project{
				{Name: "proj-a", Role: "admin"},
				{Name: "proj-b", Role: "viewer"},
			},
			selectedProject: "proj-a",
			want: "\nNAME       ROLE\n" +
				"proj-a *   admin\n" +
				"proj-b     viewer\n",
		},
		{
			name: "selection is case-insensitive",
			projects: []api.Project{
				{Name: "MyProject", Role: "admin"},
			},
			selectedProject: "myproject",
			want: "\nNAME          ROLE\n" +
				"MyProject *   admin\n",
		},
		{
			name: "no match for selected project",
			projects: []api.Project{
				{Name: "proj-a", Role: "admin"},
			},
			selectedProject: "nonexistent",
			want: "\nNAME     ROLE\n" +
				"proj-a   admin\n",
		},
		{
			name: "output starts with newline",
			projects: []api.Project{
				{Name: "proj", Role: "admin"},
			},
			want: "\nNAME   ROLE\n" +
				"proj   admin\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintProjectList(&buf, tt.projects, tt.selectedProject)
			if err != nil {
				t.Fatalf("PrintProjectList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
