package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

// savedReport is a report as GET …/reports/{id} answers it.
const savedReport = `{"id":"r-1","project_id":"proj-1","name":"Traffic","description":"Requests",
"folder_id":null,"agent_ids":["support-bot"],"all_agents":false,"show_terms":true,
"default_range":"last7Days","can_undo":true,
"created_at":"2026-11-01T10:00:00Z","created_by":"u-1",
"updated_at":"2026-11-10T09:30:00.123456Z","updated_by":"u-2","updated_via":"ui",
"terms":{"traces":"Replies"},"filters":[{"column":"environment","operator":"=","value":"prod"}],
"definition":{"version":1,"sections":[{"id":"s1","title":"Traffic","columns":3,"rows":1,
"fill":"fixed","widgets":[{"id":"w1","name":"Traces","view":"traces","slot":{"x":0,"y":0},
"metrics":[{"measure":"count","aggregation":"count"}],"dimensions":[],"filters":[],
"chartType":"NUMBER","threshold":12345678901234567}]}]}}`

// decode reads a report or rules as the client does, numbers kept exact.
func decode[T any](t *testing.T, raw string) T {
	t.Helper()
	var v T
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func definition(t *testing.T, raw string) map[string]any {
	t.Helper()
	var d map[string]any
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestReportChanges(t *testing.T) {
	const base = `{"version":1,"sections":[
	  {"id":"a","title":"Traffic","columns":3,"rows":1,"fill":"fixed","widgets":[
	    {"id":"w1","name":"Traces","view":"traces","chartType":"NUMBER"},
	    {"id":"w2","name":"Latency","view":"traces","chartType":"LINE_TIME_SERIES"}]},
	  {"id":"b","title":"Cost","columns":2,"rows":1,"fill":"fill","widgets":[
	    {"id":"w3","name":"Spend","view":"llms","chartType":"PIE"}]}]}`
	tests := []struct {
		name   string
		next   string
		folder string
		edit   func(*inputs.ReportDocument)
		want   []string
	}{
		{name: "same", next: base, folder: "Ops"},
		{
			name:   "folder case and the default range",
			next:   base,
			folder: "OPS",
			edit:   func(d *inputs.ReportDocument) { d.DefaultRange = "" },
		},
		{
			name: "fields",
			next: base,
			edit: func(d *inputs.ReportDocument) {
				d.Name, d.AgentIDs = "Traffic 2", []string{"bot"}
			},
			want: []string{
				`~ name: "Traffic" → "Traffic 2"`,
				`~ folder: "Ops" → ""`,
				`~ agent_ids: [] → ["bot"]`,
			},
		},
		{
			name:   "all_agents and show_terms",
			next:   base,
			folder: "Ops",
			edit: func(d *inputs.ReportDocument) {
				d.AllAgents, d.ShowTerms = ptr(true), ptr(true)
			},
			want: []string{`~ all_agents: false → true`},
		},
		{
			name:   "show_terms alone",
			next:   base,
			folder: "Ops",
			edit:   func(d *inputs.ReportDocument) { d.ShowTerms = ptr(false) },
			want:   []string{`~ show_terms: true → false`},
		},
		{
			name:   "all_agents alone, with no agents listed",
			next:   base,
			folder: "Ops",
			edit:   func(d *inputs.ReportDocument) { d.AgentIDs, d.AllAgents = nil, ptr(true) },
			want:   []string{`~ all_agents: false → true`},
		},
		{
			name:   "all_agents and show_terms left out keep the report's",
			next:   base,
			folder: "Ops",
			edit:   func(d *inputs.ReportDocument) { d.AllAgents, d.ShowTerms = nil, nil },
		},
		{
			name:   "sections and widgets",
			folder: "Ops",
			next: `{"version":1,"sections":[
			  {"id":"b","title":"Spend","columns":2,"rows":1,"fill":"fill","widgets":[
			    {"id":"w3","name":"Spend","view":"llms","chartType":"TABLE"},
			    {"id":"w2","name":"Latency","view":"traces","chartType":"LINE_TIME_SERIES"}]},
			  {"id":"c","title":"Errors","columns":1,"rows":1,"fill":"fill","widgets":[
			    {"id":"w4","name":"Errors","view":"traces","chartType":"NUMBER"}]}]}`,
			want: []string{
				`~ section "Spend" (b): title "Cost" → "Spend"`,
				`+ section "Errors" (c)`,
				`- section "Traffic" (a)`,
				`~ widget "Spend" (w3): chartType "PIE" → "TABLE"`,
				`~ widget "Latency" (w2): moved from "Traffic" (a) to "Spend" (b)`,
				`+ widget "Errors" (w4) in "Errors" (c)`,
				`- widget "Traces" (w1) from "Traffic" (a)`,
			},
		},
		{
			name:   "order",
			folder: "Ops",
			next: `{"version":1,"sections":[
			  {"id":"b","title":"Cost","columns":2,"rows":1,"fill":"fill","widgets":[
			    {"id":"w3","name":"Spend","view":"llms","chartType":"PIE"}]},
			  {"id":"a","title":"Traffic","columns":3,"rows":1,"fill":"fixed","widgets":[
			    {"id":"w2","name":"Latency","view":"traces","chartType":"LINE_TIME_SERIES"},
			    {"id":"w1","name":"Traces","view":"traces","chartType":"NUMBER"}]}]}`,
			want: []string{
				`~ section order changed`,
				`~ section "Traffic" (a): widget order changed`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved := inputs.ReportDocument{
				Name:         "Traffic",
				DefaultRange: "last7Days",
				Definition:   definition(t, base),
				AllAgents:    ptr(false),
				ShowTerms:    ptr(true),
			}
			next := saved
			next.Definition = definition(t, tt.next)
			if tt.edit != nil {
				tt.edit(&next)
			}
			got := ReportChanges(saved, next, "Ops", tt.folder)
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf(
					"changes:\n%s\nwant:\n%s",
					strings.Join(got, "\n"),
					strings.Join(tt.want, "\n"),
				)
			}
		})
	}
}

func TestPrintDeletedReport(t *testing.T) {
	var out bytes.Buffer
	doc := json.RawMessage(`{"name":"Traffic","folder":"Ops","threshold":12345678901234567}`)
	if err := PrintDeletedReport(&out, "Traffic", "r-1", doc); err != nil {
		t.Fatal(err)
	}
	want := "# Deleted report \"Traffic\" (id r-1). Create it again with 'iai reports create --file'.\n" +
		"folder: Ops\nname: Traffic\nthreshold: 12345678901234567\n"
	if out.String() != want {
		t.Errorf("got\n%s\nwant\n%s", out.String(), want)
	}
}

func TestPrintReportList(t *testing.T) {
	folder := "f-1"
	reports := []platform.Report{
		{ID: "r-1", Name: "Traffic", FolderID: &folder, UpdatedVia: "cli"},
		{ID: "r-2", Name: "Cost", AgentIDs: []string{"bot-a", "bot-b"}, UpdatedVia: "ui"},
	}
	tests := []struct {
		name    string
		reports []platform.Report
		columns []string
		want    string
	}{
		{name: "empty", columns: inputs.DefaultReportColumns, want: "No reports found.\n"},
		{
			name:    "folder name and scope",
			reports: reports,
			columns: []string{"id", "name", "folder", "agents", "updated_via"},
			want: "ID    NAME      FOLDER   AGENTS         UPDATED VIA\n" +
				"r-1   Traffic   Ops      project        cli\n" +
				"r-2   Cost               bot-a, bot-b   ui\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := PrintReportList(&out, tt.reports, map[string]string{"f-1": "Ops"}, tt.columns)
			if err != nil || out.String() != tt.want {
				t.Errorf("got %v\n%s\nwant\n%s", err, out.String(), tt.want)
			}
		})
	}
}

// overlaid is savedReport read with ?include=draft while openDraft is open.
func overlaid(t *testing.T) json.RawMessage {
	t.Helper()
	raw, err := platform.OverlayDraft(
		json.RawMessage(strings.TrimSuffix(savedReport, "}") + `,"draft":` + openDraft + "}"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// openDraft is the draft of savedReport GET …/reports/{id}?include=draft answers.
const openDraft = `{"id":"d-1","revision":3,"document":{"name":"Traffic v2",
"description":"","folder_id":"f-2","agent_ids":[],"all_agents":true,"show_terms":false,
"default_range":"last30Days","terms":{},"filters":[],
"definition":{"version":1,"sections":[{"id":"s2","title":"Costs","columns":2,"rows":1,
"fill":"fixed","widgets":[{"id":"w2","name":"Spend","view":"traces",
"metrics":[{"measure":"totalCost","aggregation":"sum"}],"chartType":"LINE_TIME_SERIES"}]}]}}}`

func TestPrintReportSummary(t *testing.T) {
	report := decode[platform.Report](t, savedReport)
	draft := decode[platform.Report](t, string(overlaid(t)))

	saved := "ID:              r-1\n" +
		"Name:            Traffic\n" +
		"Description:     Requests\n" +
		"Folder:          Ops\n" +
		"Scope:           support-bot\n" +
		"Default Range:   last7Days\n" +
		"Terms:           {\"traces\":\"Replies\"}\n" +
		"Filters:         [{\"column\":\"environment\",\"operator\":\"=\",\"value\":\"prod\"}]\n" +
		"Created At:      " + LocalTime("2026-11-01T10:00:00Z") + "\n" +
		"Updated At:      " + LocalTime("2026-11-10T09:30:00.123456Z") + "\n" +
		"Updated By:      u-2\n" +
		"Updated Via:     ui\n" +
		"Can Restore:     yes\n" +
		"\nSections:\n" +
		"  \"Traffic\" (s1) (3 × 1, fixed)\n" +
		"    - \"Traces\" (w1)  [traces · count/count · NUMBER]\n"
	tests := []struct {
		name   string
		report platform.Report
		folder string
		want   string
	}{
		{"saved report", report, "Ops", saved},
		{
			"open draft",
			draft,
			"Drafts",
			"ID:              r-1\n" +
				"Name:            Traffic v2\n" +
				"Folder:          Drafts\n" +
				"Scope:           all agents\n" +
				"Default Range:   last30Days\n" +
				"Created At:      " + LocalTime("2026-11-01T10:00:00Z") + "\n" +
				"Updated At:      " + LocalTime("2026-11-10T09:30:00.123456Z") + "\n" +
				"Updated By:      u-2\n" +
				"Updated Via:     ui\n" +
				"Can Restore:     yes\n" +
				"Unsaved Draft:   yes, open in the editor; the settings and sections are the draft's\n" +
				"\nSections:\n" +
				"  \"Costs\" (s2) (2 × 1, fixed)\n" +
				"    - \"Spend\" (w2)  [traces · totalCost/sum · LINE_TIME_SERIES]\n",
		},
	}
	for _, tt := range tests {
		var out bytes.Buffer
		if err := PrintReportSummary(&out, tt.report, tt.folder); err != nil {
			t.Fatal(err)
		}
		if out.String() != tt.want {
			t.Errorf("%s: got\n%s\nwant\n%s", tt.name, out.String(), tt.want)
		}
	}
}

// A draft printed with --yaml saves back into that draft, at the revision it was read at.
func TestReportDraftRoundTrip(t *testing.T) {
	doc, err := inputs.ReportDocumentOf(overlaid(t), "Drafts", true)
	if err != nil {
		t.Fatal(err)
	}
	var yaml bytes.Buffer
	if err := PrintRawYAML(&yaml, doc); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(yaml.String(), "\ndraft_revision: 3\n") ||
		!strings.Contains(yaml.String(), "\nname: Traffic v2\n") {
		t.Errorf("yaml:\n%s", yaml.String())
	}
	parsed, err := inputs.ParseReportDocument(yaml.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(inputs.ReportBodyOf(parsed, "r-1", false, "copilot"))
	want := `{"id":"r-1","updated_via":"copilot",` +
		`"draft_revision":3,"folder":"Drafts","name":"Traffic v2","description":"",` +
		`"all_agents":true,"show_terms":false,` +
		`"default_range":"last30Days","definition":{"sections":[{"columns":2,` +
		`"fill":"fixed","id":"s2","rows":1,"title":"Costs","widgets":[{"chartType":` +
		`"LINE_TIME_SERIES","id":"w2","metrics":[{"aggregation":"sum","measure":"totalCost"}],` +
		`"name":"Spend","view":"traces"}]}],"version":1}}`
	if string(body) != want {
		t.Errorf("body:\ngot  %s\nwant %s", body, want)
	}
}

func TestPrintReportSaved(t *testing.T) {
	revision := int64(4)
	tests := []struct {
		draft *int64
		want  string
	}{
		{nil, "Saved report \"Traffic\" (id r-1).\nPath: /p\n"},
		{
			&revision,
			"Added to your open draft of report \"Traffic\" (id r-1, draft_revision: 4). " +
				"Save in the editor to keep it.\nPath: /p\n",
		},
	}
	for _, tt := range tests {
		var out bytes.Buffer
		r := platform.Report{ID: "r-1", Name: "Traffic", DraftRevision: tt.draft}
		PrintReportSaved(&out, r, "/p")
		if out.String() != tt.want {
			t.Errorf("draft %v: got %q, want %q", tt.draft, out.String(), tt.want)
		}
	}
}

func TestPrintReportSavePreview(t *testing.T) {
	saved := decode[platform.Report](t, savedReport)
	next := reportDocumentOf(saved)
	sectionC := []any{map[string]any{"id": "c", "title": "Cost", "widgets": []any{}}}
	tests := []struct {
		name  string
		saved *platform.Report
		edit  func(*inputs.ReportDocument)
		force bool
		want  string
	}{
		{
			name: "new report",
			edit: func(d *inputs.ReportDocument) {
				d.Definition = map[string]any{"version": 1, "sections": sectionC}
			},
			want: "Would create report \"Traffic\":\n  ~ folder: \"\" → \"ops\"\n  + section \"Cost\" (c)\nDry run: nothing was saved.\n",
		},
		{
			name:  "unchanged",
			saved: &saved,
			want:  "Would update report \"Traffic\" (id r-1):\n  No changes.\nDry run: nothing was saved.\n",
		},
		{
			name:  "changed since read",
			saved: &saved,
			edit: func(d *inputs.ReportDocument) {
				d.UpdatedAt, d.DefaultRange = "2026-11-09T00:00:00Z", "last30Days"
			},
			want: "The report changed after this document was read (saved by u-2 at " +
				LocalTime("2026-11-10T09:30:00.123456Z") + " via ui); " +
				"update will be refused without --force.\n" +
				"Would update report \"Traffic\" (id r-1):\n" +
				"  ~ default_range: \"last7Days\" → \"last30Days\"\n" +
				"Dry run: nothing was saved.\n",
		},
		{
			name:  "into the open draft",
			saved: &saved,
			edit: func(d *inputs.ReportDocument) {
				revision := int64(3)
				d.DraftRevision, d.DefaultRange = &revision, "last30Days"
			},
			want: "The document goes into your open draft (revision 3); " +
				"the changes below are against the saved report.\n" +
				"Would update report \"Traffic\" (id r-1):\n" +
				"  ~ default_range: \"last7Days\" → \"last30Days\"\n" +
				"Dry run: nothing was saved.\n",
		},
		{
			name:  "into the open draft, read before the report was saved: no --force warning",
			saved: &saved,
			edit: func(d *inputs.ReportDocument) {
				revision := int64(3)
				d.DraftRevision, d.UpdatedAt = &revision, "2026-11-09T00:00:00Z"
			},
			want: "The document goes into your open draft (revision 3); " +
				"the changes below are against the saved report.\n" +
				"Would update report \"Traffic\" (id r-1):\n  No changes.\nDry run: nothing was saved.\n",
		},
		{
			name:  "changed since read with --force",
			saved: &saved,
			edit:  func(d *inputs.ReportDocument) { d.UpdatedAt = "2026-11-09T00:00:00Z" },
			force: true,
			want:  "Would update report \"Traffic\" (id r-1):\n  No changes.\nDry run: nothing was saved.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := next
			if tt.edit != nil {
				tt.edit(&doc)
			}
			before := ""
			if tt.saved != nil {
				before = "Ops"
			}
			var out bytes.Buffer
			PrintReportSavePreview(&out, tt.saved, doc, before, "ops", tt.force)
			if out.String() != tt.want {
				t.Errorf("got\n%s\nwant\n%s", out.String(), tt.want)
			}
		})
	}
}

func TestReportWriteErrors(t *testing.T) {
	changed := &platform.ReportError{Code: platform.ReportChangedCode, Details: map[string]any{
		"updated_at": "2026-11-10T09:30:00Z", "updated_by": "u-9", "updated_via": "copilot",
	}}
	at := LocalTime("2026-11-10T09:30:00Z")
	restore := func(err error, ref, _ string) error { return ReportRestoreError(err, ref) }
	tests := []struct {
		name    string
		explain func(err error, ref, name string) error
		err     error
		want    string
	}{
		{
			name:    "save: changed",
			explain: ReportWriteError,
			err:     changed,
			want: "the report changed after this document was read: saved by u-9 at " + at +
				" via copilot; run 'iai reports get \"Traffic\" --yaml' and make the change on that, " +
				"or re-run with --force to overwrite it",
		},
		{
			name:    "save: name taken",
			explain: ReportWriteError,
			err: &platform.ReportError{
				Code:    platform.ReportNameTakenCode,
				Message: "A report named 'Cost' already exists in this project",
				Details: map[string]any{"existing_id": "r-7"},
			},
			want: "A report named 'Cost' already exists in this project; " +
				"save under another name, or change that report with 'iai reports update \"Cost\" --file'",
		},
		{
			name:    "save: widget rejected",
			explain: ReportWriteError,
			err: &platform.ReportError{
				Code:    "INVALID_REQUEST",
				Message: `Widget "Spend" (w2): a Pie needs a breakdown`,
				Details: map[string]any{"widget_id": "w2", "rule": "pie_needs_breakdown"},
			},
			want: `Widget "Spend" (w2): a Pie needs a breakdown (rule pie_needs_breakdown); ` +
				"run 'iai reports schema' for the legal selections",
		},
		{
			name:    "save: other refusal",
			explain: ReportWriteError,
			err:     &platform.ReportError{Code: "INVALID_REQUEST", Message: "too many sections"},
			want:    "too many sections",
		},
		{
			name:    "save: not a refusal",
			explain: ReportWriteError,
			err:     errors.New("failed to save report: timeout"),
			want:    "failed to save report: timeout",
		},
		{
			name:    "restore: changed",
			explain: restore,
			err:     changed,
			want: "the report was saved again after the expected updated_at: by u-9 at " + at +
				" via copilot; run 'iai reports get \"Traffic\"' to look at it before deciding whether to restore",
		},
		{
			name:    "restore: other refusal",
			explain: restore,
			err:     &platform.ReportError{Code: "NOTHING_TO_UNDO", Message: "Nothing to restore"},
			want:    "Nothing to restore",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.explain(tt.err, "Traffic", "Cost").Error(); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestPrintReportRules(t *testing.T) {
	const rules = `{"version":1,
"sources":[{"id":"traces","label":"Traces","agent_scope":true,"combinable":false},
{"id":"llms","label":"LLMs","agent_scope":true,"combinable":true}],
"combined":{"view":"billing","sources":["llms","agents"],"min_sources":2,"max_sources":2},
"chart_types":[{"id":"NUMBER"},{"id":"PIE"}],
"selections":[{"view":"traces","sources":["traces"],"measure":"count","aggregation":"count",
"breakdown":null,"chart_types":["NUMBER"],"agent_scope":true},
{"view":"billing","sources":["llms","agents"],"measure":"cost","aggregation":"sum",
"breakdown":"agent","chart_types":["PIE"],"agent_scope":false}],
"limits":{"sections":20,"body_bytes":262144},
"default_ranges":["last1Day","last7Days"],"report_filter_columns":["tags"]}`
	want := `Rules version: 1

Sources:
ID       LABEL    COMBINABLE   AGENT SCOPE
traces   Traces   no           yes
llms     LLMs     yes          yes

Combined (view billing): 2 to 2 of llms, agents
Chart types: NUMBER, PIE

Selections:
VIEW      SOURCES       METRIC        BREAKDOWN   CHART TYPES   AGENT SCOPE
traces    traces        count/count   -           NUMBER        yes
billing   llms+agents   cost/sum      agent       PIE           no

Limits:
  body_bytes:   262144
  sections:     20

Default ranges: last1Day, last7Days
Report filter columns: tags
`
	var out bytes.Buffer
	if err := PrintReportRules(&out, decode[platform.ReportRules](t, rules)); err != nil {
		t.Fatal(err)
	}
	if out.String() != want {
		t.Errorf("got\n%s\nwant\n%s", out.String(), want)
	}
}

func ptr[T any](v T) *T { return &v }

func TestReportScope(t *testing.T) {
	tests := []struct {
		name   string
		report platform.Report
		want   string
	}{
		{"no agents", platform.Report{}, "project"},
		{"some agents", platform.Report{AgentIDs: []string{"bot"}}, "bot"},
		{"all agents, with an empty list", platform.Report{AllAgents: true}, "all agents"},
	}
	for _, tt := range tests {
		if got := reportScope(&tt.report); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
