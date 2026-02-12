package output

import (
	"bytes"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintOrganizationList(t *testing.T) {
	tests := []struct {
		name        string
		orgs        []clients.Organization
		selectedOrg string
		want        string
	}{
		{
			name: "empty list prints headers only",
			orgs: []clients.Organization{},
			want: "NAME   PROJECTS   ROLE\n",
		},
		{
			name: "single org no selection",
			orgs: []clients.Organization{
				{Name: "acme", ProjectCount: 5, Role: "owner"},
			},
			selectedOrg: "",
			want: "NAME   PROJECTS   ROLE\n" +
				"acme   5          owner\n",
		},
		{
			name: "selected org gets marker",
			orgs: []clients.Organization{
				{Name: "org-a", ProjectCount: 3, Role: "admin"},
				{Name: "org-b", ProjectCount: 1, Role: "viewer"},
			},
			selectedOrg: "org-a",
			want: "NAME      PROJECTS   ROLE\n" +
				"org-a *   3          admin\n" +
				"org-b     1          viewer\n",
		},
		{
			name: "selection is case-insensitive",
			orgs: []clients.Organization{
				{Name: "MyOrg", ProjectCount: 2, Role: "admin"},
			},
			selectedOrg: "myorg",
			want: "NAME      PROJECTS   ROLE\n" +
				"MyOrg *   2          admin\n",
		},
		{
			name: "project count zero",
			orgs: []clients.Organization{
				{Name: "new-org", ProjectCount: 0, Role: "owner"},
			},
			want: "NAME      PROJECTS   ROLE\n" +
				"new-org   0          owner\n",
		},
		{
			name: "no match for selected org",
			orgs: []clients.Organization{
				{Name: "org-a", ProjectCount: 1, Role: "admin"},
			},
			selectedOrg: "nonexistent",
			want: "NAME    PROJECTS   ROLE\n" +
				"org-a   1          admin\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintOrganizationList(&buf, tt.orgs, tt.selectedOrg)
			if err != nil {
				t.Fatalf("PrintOrganizationList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
