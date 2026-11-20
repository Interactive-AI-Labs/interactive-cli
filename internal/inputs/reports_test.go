package inputs

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func TestReportBody(t *testing.T) {
	const def = "definition: {version: 1, sections: []}\n"
	const plain = `{"updated_via":"cli","name":"A","description":"","definition":{"sections":[],"version":1}}`
	const replace = "updated_at: 2026-11-10T09:30:00.123456Z\nname: A\n" + def
	tests := []struct {
		name, in, id string
		force        bool
		want         string
	}{
		{name: "yaml", in: "name: A\n" + def, want: plain},
		{
			name: "json",
			in:   `{"name": "A", "definition": {"version": 1, "sections": []}}`,
			want: plain,
		},
		{
			name: "yaml flow mapping",
			in:   "{name: A, definition: {version: 1, sections: []}}",
			want: plain,
		},
		{name: "create ignores --force", in: "name: A\n" + def, force: true, want: plain},
		{
			name: "replace is guarded by updated_at",
			in:   replace,
			id:   "r-1",
			want: `{"id":"r-1","expected_updated_at":"2026-11-10T09:30:00.123456Z",` +
				`"updated_via":"cli","name":"A","description":"","definition":{"sections":[],"version":1}}`,
		},
		{
			name:  "replace with --force",
			in:    replace,
			id:    "r-1",
			force: true,
			want: `{"id":"r-1","expected_updated_at":"2026-11-10T09:30:00.123456Z","force":true,` +
				`"updated_via":"cli","name":"A","description":"","definition":{"sections":[],"version":1}}`,
		},
		{
			name: "read-only fields are dropped",
			in: "project_id: p\ncan_undo: true\ncreated_at: x\ncreated_by: u\nupdated_by: u\n" +
				"updated_via: ui\nid: r-1\nname: A\n" + def,
			want: plain,
		},
		{
			name: "a folder name moves it",
			in:   "name: A\nfolder: Ops\n" + def,
			want: `{"updated_via":"cli","folder":"Ops","name":"A","description":"",` +
				`"definition":{"sections":[],"version":1}}`,
		},
		{
			name: "folder null unfiles",
			in:   "name: A\nfolder: null\n" + def,
			want: `{"updated_via":"cli","folder_id":null,"name":"A","description":"",` +
				`"definition":{"sections":[],"version":1}}`,
		},
		{name: "no folder keeps it", in: "name: A\n" + def, want: plain},
		{
			name: "a document get --yaml printed",
			in: "agent_ids:\n    - support-bot\ncan_undo: true\ncreated_at: \"2026-11-01T10:00:00Z\"\n" +
				"default_range: last7Days\ndefinition:\n    sections: []\n    threshold: 12345678901234567\n" +
				"    version: 1\ndescription: Requests\nfolder: Ops\nid: r-1\nname: Traffic\n" +
				"project_id: proj-1\nupdated_at: \"2026-11-10T09:30:00.123456Z\"\nupdated_by: u-2\nupdated_via: ui\n",
			id: "r-1",
			want: `{"id":"r-1","expected_updated_at":"2026-11-10T09:30:00.123456Z","updated_via":"cli",` +
				`"folder":"Ops","name":"Traffic","description":"Requests","agent_ids":["support-bot"],` +
				`"default_range":"last7Days","definition":{"sections":[],"threshold":12345678901234567,"version":1}}`,
		},
		{
			name: "an empty folder name",
			in:   "name: A\nfolder: ' '\n" + def,
			want: "folder must be a folder name, or null for no folder",
		},
		{
			name: "a folder that is not a name",
			in:   "name: A\nfolder: [Ops]\n" + def,
			want: "folder must be a folder name, or null for no folder",
		},
		{
			name: "yaml keys that are not strings",
			in:   "name: A\nterms: {2024: Year}\n" + def,
			want: `{"updated_via":"cli","name":"A","description":"","terms":{"2024":"Year"},` +
				`"definition":{"sections":[],"version":1}}`,
		},
		{
			name: "folder_id is no longer read",
			in:   "name: A\nfolder_id: f-1\n" + def,
			want: `failed to parse report document: json: unknown field "folder_id"`,
		},
		{
			name: "unknown json field",
			in:   `{"name": "A", "widgets": [], "definition": {}}`,
			want: `failed to parse report document: json: unknown field "widgets"`,
		},
		{
			name: "wrong type",
			in:   "name: [A]\n" + def,
			want: "failed to parse report document: json: cannot unmarshal array into " +
				"Go struct field ReportDocument.name of type string",
		},
		{name: "not a mapping", in: "- a\n", want: "failed to parse report document: yaml: " +
			"unmarshal errors: line 1: cannot unmarshal !!seq into map[string]interface {}"},
		{name: "empty", in: " \n", want: "the report document is empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ""
			doc, err := ParseReportDocument([]byte(tt.in))
			if err != nil {
				got = err.Error()
			} else {
				data, _ := json.Marshal(ReportBodyOf(doc, tt.id, tt.force, "cli"))
				got = string(data)
			}
			if got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestReportFolderChange(t *testing.T) {
	tests := []struct {
		in, want string
		keep     bool
	}{
		{"folder: Sales", "Sales", false},
		{"folder: null", "", false},
		{"folder:", "", false},
		{"description: x", "", true},
	}
	for _, tt := range tests {
		doc, err := ParseReportDocument([]byte("name: A\ndefinition: {}\n" + tt.in + "\n"))
		if err != nil {
			t.Fatalf("%s: %v", tt.in, err)
		}
		if got, keep := ReportFolderChange(doc); got != tt.want || keep != tt.keep {
			t.Errorf(
				"ReportFolderChange(%s) = %q, %v; want %q, %v",
				tt.in,
				got,
				keep,
				tt.want,
				tt.keep,
			)
		}
	}
}

func TestReportDocumentOf(t *testing.T) {
	const raw = `{"id":"r-1","updated_at":"t","name":"A","folder_id":"f-1","n":12345678901234567}`
	tests := []struct {
		name, folder  string
		withUpdatedAt bool
		want          string
	}{
		{
			"get names the folder",
			"Ops",
			true,
			`{"folder":"Ops","n":12345678901234567,"name":"A","updated_at":"t"}`,
		},
		{
			"no folder is null",
			"",
			true,
			`{"folder":null,"n":12345678901234567,"name":"A","updated_at":"t"}`,
		},
		{
			"delete drops updated_at",
			"Ops",
			false,
			`{"folder":"Ops","n":12345678901234567,"name":"A"}`,
		},
	}
	for _, tt := range tests {
		got, err := ReportDocumentOf(json.RawMessage(raw), tt.folder, tt.withUpdatedAt)
		if string(got) != tt.want || err != nil {
			t.Errorf("%s: got %s, %v; want %s", tt.name, got, err, tt.want)
		}
	}
}

func TestRestoreReportBodyOf(t *testing.T) {
	ptr := func(s string) *string { return &s }
	tests := []struct {
		name     string
		expected *string
		want     string
	}{
		{
			"default is the current updated_at",
			nil,
			`{"expected_updated_at":"2026-11-10T09:30:00Z","updated_via":"cli"}`,
		},
		{
			"as typed",
			ptr("2026-11-09T09:00:00+01:00"),
			`{"expected_updated_at":"2026-11-09T09:00:00+01:00","updated_via":"cli"}`,
		},
		{
			"empty",
			ptr(""),
			"--expected-updated-at is empty; leave it out to use the report's current updated_at",
		},
		{
			"blank",
			ptr("  "),
			"--expected-updated-at is empty; leave it out to use the report's current updated_at",
		},
	}
	for _, tt := range tests {
		got := ""
		body, err := RestoreReportBodyOf("2026-11-10T09:30:00Z", tt.expected, "cli")
		if err != nil {
			got = err.Error()
		} else {
			data, _ := json.Marshal(body)
			got = string(data)
		}
		if got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.name, got, tt.want)
		}
	}
}

func TestResolveReport(t *testing.T) {
	reports := []platform.Report{{ID: "r-1", Name: "Traffic"}, {ID: "r-2", Name: "Cost"}}
	tests := []struct {
		name, want string
	}{
		{"traffic", "r-1"},
		{"COST", "r-2"},
		{"r-1", `report "r-1" not found; run 'iai reports list'`},
		{"Latency", `report "Latency" not found; run 'iai reports list'`},
	}
	for _, tt := range tests {
		got, err := ResolveReport(reports, tt.name)
		if err != nil {
			got.ID = err.Error()
		}
		if got.ID != tt.want {
			t.Errorf("ResolveReport(%q) = %q, want %q", tt.name, got.ID, tt.want)
		}
	}
}

func TestResolveReportFolder(t *testing.T) {
	folders := []platform.ReportFolder{{ID: "f-1", Name: "Ops"}}
	tests := []struct {
		ref, want string
	}{
		{"ops", "f-1"},
		{"f-1", `report folder "f-1" not found; run 'iai reports list'`},
		{"none", "none"},
		{"None", "none"},
		{"Sales", `report folder "Sales" not found; run 'iai reports list'`},
	}
	for _, tt := range tests {
		got, err := ResolveReportFolder(folders, tt.ref)
		if err != nil {
			got = err.Error()
		}
		if got != tt.want {
			t.Errorf("ResolveReportFolder(%q) = %q, want %q", tt.ref, got, tt.want)
		}
	}
}

func TestReportsUpdatedVia(t *testing.T) {
	for origin, want := range map[string]string{"": "cli", "copilot": "copilot", "ui": "cli"} {
		if got := ReportsUpdatedVia(origin); got != want {
			t.Errorf("ReportsUpdatedVia(%q) = %q, want %q", origin, got, want)
		}
	}
}

func TestReportColumns(t *testing.T) {
	tests := []struct {
		columns []string
		want    string
	}{
		{DefaultReportColumns, ""},
		{AllReportColumns, ""},
		{
			[]string{"name", "owner"},
			`unknown column "owner" (available: id, name, description, folder, agents, ` +
				`default_range, updated_at, updated_by, updated_via, created_at)`,
		},
	}
	for _, tt := range tests {
		got := ""
		if err := ValidateColumns(tt.columns, AllReportColumns); err != nil {
			got = err.Error()
		}
		if got != tt.want {
			t.Errorf("ValidateColumns(%v) = %q, want %q", tt.columns, got, tt.want)
		}
	}
}
