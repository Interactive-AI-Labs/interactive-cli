package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

var reportColumnMap = map[string]struct {
	Header string
	Value  func(r *platform.Report, folders map[string]string) string
}{
	"id": {"ID", func(r *platform.Report, _ map[string]string) string { return r.ID }},
	"name": {
		"NAME",
		func(r *platform.Report, _ map[string]string) string { return Printable(r.Name) },
	},
	"description": {
		"DESCRIPTION",
		func(r *platform.Report, _ map[string]string) string { return Printable(r.Description) },
	},
	"folder": {
		"FOLDER",
		func(r *platform.Report, folders map[string]string) string {
			return Printable(inputs.ReportFolderName(r.FolderID, folders))
		},
	},
	"agents": {
		"AGENTS",
		func(r *platform.Report, _ map[string]string) string { return reportScope(r.AgentIDs) },
	},
	"default_range": {
		"DEFAULT RANGE",
		func(r *platform.Report, _ map[string]string) string { return r.DefaultRange },
	},
	"updated_at": {
		"UPDATED AT",
		func(r *platform.Report, _ map[string]string) string { return LocalTime(r.UpdatedAt) },
	},
	"updated_by": {
		"UPDATED BY",
		func(r *platform.Report, _ map[string]string) string { return Printable(r.UpdatedBy) },
	},
	"updated_via": {
		"UPDATED VIA",
		func(r *platform.Report, _ map[string]string) string { return r.UpdatedVia },
	},
	"created_at": {
		"CREATED AT",
		func(r *platform.Report, _ map[string]string) string { return LocalTime(r.CreatedAt) },
	},
}

func reportScope(agentIDs []string) string {
	if len(agentIDs) == 0 {
		return "project"
	}
	return Printable(TruncateList(agentIDs, 3))
}

func PrintReportList(
	out io.Writer,
	reports []platform.Report,
	folders map[string]string,
	columns []string,
) error {
	if len(reports) == 0 {
		fmt.Fprintln(out, "No reports found.")
		return nil
	}
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = reportColumnMap[col].Header
	}
	rows := make([][]string, len(reports))
	for i := range reports {
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = reportColumnMap[col].Value(&reports[i], folders)
		}
		rows[i] = row
	}
	return PrintTable(out, headers, rows)
}

// PrintReportSummary prints a report's settings and its sections with their widgets;
// draft says the caller has an unsaved draft of it open.
func PrintReportSummary(out io.Writer, r platform.Report, folder string, draft bool) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", r.ID)
	fmt.Fprintf(w, "Name:\t%s\n", Printable(r.Name))
	if r.Description != "" {
		fmt.Fprintf(w, "Description:\t%s\n", Printable(r.Description))
	}
	if folder != "" {
		fmt.Fprintf(w, "Folder:\t%s\n", Printable(folder))
	}
	fmt.Fprintf(w, "Scope:\t%s\n", reportScope(r.AgentIDs))
	fmt.Fprintf(w, "Default Range:\t%s\n", r.DefaultRange)
	if len(r.Terms) > 0 {
		fmt.Fprintf(w, "Terms:\t%s\n", reportJSON(r.Terms))
	}
	if len(r.Filters) > 0 {
		fmt.Fprintf(w, "Filters:\t%s\n", reportJSON(r.Filters))
	}
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(r.CreatedAt))
	fmt.Fprintf(w, "Updated At:\t%s\n", LocalTime(r.UpdatedAt))
	fmt.Fprintf(w, "Updated By:\t%s\n", Printable(r.UpdatedBy))
	fmt.Fprintf(w, "Updated Via:\t%s\n", r.UpdatedVia)
	fmt.Fprintf(w, "Can Restore:\t%s\n", reportYesNo(r.CanUndo))
	if draft {
		fmt.Fprintln(w, "Unsaved Draft:\tyes, open in the editor; --yaml prints it")
	}
	if err := w.Flush(); err != nil {
		return err
	}
	sections := reportSections(r.Definition)
	if len(sections) == 0 {
		fmt.Fprintln(out, "\nNo sections.")
		return nil
	}
	fmt.Fprintln(out, "\nSections:")
	for _, s := range sections {
		fmt.Fprintf(
			out,
			"  %s (%s × %s, %s)\n",
			reportSectionLabel(
				s,
			),
			reportJSON(s.raw["columns"]),
			reportJSON(s.raw["rows"]),
			reportText(s.raw["fill"]),
		)
		for _, widget := range s.widgets {
			fmt.Fprintf(
				out,
				"    - %s  %s\n",
				reportWidgetLabel(widget),
				reportWidgetSelection(widget),
			)
		}
	}
	return nil
}

// PrintDeletedReport prints a deleted report's document as YAML, under a comment naming it.
func PrintDeletedReport(out io.Writer, name, id string, doc json.RawMessage) error {
	fmt.Fprintf(
		out,
		"# Deleted report %q (id %s). Create it again with 'iai reports create --file'.\n",
		Printable(name), id,
	)
	return PrintRawYAML(out, doc)
}

// PrintReportSaved prints the report a create or update saved and where to open it.
func PrintReportSaved(out io.Writer, r platform.Report, path string) {
	if r.Draft {
		fmt.Fprintf(
			out,
			"Added to your open draft of report %q (id %s). Save in the editor to keep it.\n"+
				"Path: %s\n",
			Printable(r.Name), r.ID, path,
		)
		return
	}
	fmt.Fprintf(out, "Saved report %q (id %s).\nPath: %s\n", Printable(r.Name), r.ID, path)
}

// PrintReportRestored prints the report a restore brought back and where to open it.
func PrintReportRestored(out io.Writer, r platform.Report, path string) {
	fmt.Fprintf(
		out, "Restored the previous version of report %q (id %s).\nPath: %s\n",
		Printable(r.Name), r.ID, path,
	)
}

// reportDocumentOf is a saved report as a document.
func reportDocumentOf(r platform.Report) inputs.ReportDocument {
	return inputs.ReportDocument{
		UpdatedAt:    r.UpdatedAt,
		Name:         r.Name,
		Description:  r.Description,
		AgentIDs:     r.AgentIDs,
		Terms:        r.Terms,
		DefaultRange: r.DefaultRange,
		Filters:      r.Filters,
		Definition:   r.Definition,
	}
}

// PrintReportSavePreview prints what saving next would change: over saved, or as a new
// report when saved is nil. before and after are the report's folder names.
func PrintReportSavePreview(
	out io.Writer,
	saved *platform.Report,
	next inputs.ReportDocument,
	before, after string,
	force bool,
) {
	// A new report lists its folder and sections, not every field as added.
	current := next
	current.Definition = nil
	if saved == nil {
		fmt.Fprintf(out, "Would create report %q:\n", Printable(next.Name))
	} else {
		current = reportDocumentOf(*saved)
		if saved.UpdatedAt != next.UpdatedAt && !force {
			fmt.Fprintf(
				out,
				"The report changed after this document was read (saved by %s at %s via %s); "+
					"update will be refused without --force.\n",
				Printable(saved.UpdatedBy), LocalTime(saved.UpdatedAt), saved.UpdatedVia,
			)
		}
		fmt.Fprintf(out, "Would update report %q (id %s):\n", Printable(saved.Name), saved.ID)
	}
	changes := ReportChanges(current, next, before, after)
	if len(changes) == 0 {
		fmt.Fprintln(out, "  No changes.")
	}
	for _, line := range changes {
		fmt.Fprintln(out, "  "+line)
	}
	fmt.Fprintln(out, "Dry run: nothing was saved.")
}

// ReportChanges lists what saving next over saved changes, one line per change; before and
// after are the folder names. Folder names compare ignoring case and a missing
// default_range is the platform's default.
func ReportChanges(saved, next inputs.ReportDocument, before, after string) []string {
	var lines []string
	field := func(name string, a, b any) {
		if reportJSON(a) != reportJSON(b) {
			lines = append(lines, fmt.Sprintf("~ %s: %s → %s", name, reportJSON(a), reportJSON(b)))
		}
	}
	field("name", saved.Name, next.Name)
	field("description", saved.Description, next.Description)
	if !strings.EqualFold(before, after) {
		field("folder", before, after)
	}
	field("agent_ids", reportStrings(saved.AgentIDs), reportStrings(next.AgentIDs))
	field("terms", reportTerms(saved.Terms), reportTerms(next.Terms))
	field(
		"default_range",
		reportDefaultRange(saved.DefaultRange),
		reportDefaultRange(next.DefaultRange),
	)
	field("filters", reportList(saved.Filters), reportList(next.Filters))
	return append(lines, reportDefinitionChanges(saved.Definition, next.Definition)...)
}

func reportDefaultRange(r string) string {
	if r == "" {
		return platform.DefaultReportRange
	}
	return r
}

func reportDefinitionChanges(saved, next map[string]any) []string {
	var lines []string
	if reportJSON(saved["version"]) != reportJSON(next["version"]) && saved != nil {
		lines = append(lines, fmt.Sprintf(
			"~ definition version: %s → %s",
			reportJSON(saved["version"]),
			reportJSON(next["version"]),
		))
	}
	oldSections, newSections := reportSections(saved), reportSections(next)
	oldByID, newByID := reportSectionsByID(oldSections), reportSectionsByID(newSections)

	for _, s := range newSections {
		old, ok := oldByID[s.id]
		if !ok {
			lines = append(lines, "+ section "+reportSectionLabel(s))
			continue
		}
		if changed := reportChangedKeys(old.raw, s.raw); len(changed) > 0 {
			lines = append(lines, fmt.Sprintf("~ section %s: %s", reportSectionLabel(s), changed))
		}
	}
	for _, s := range oldSections {
		if _, ok := newByID[s.id]; !ok {
			lines = append(lines, "- section "+reportSectionLabel(s))
		}
	}
	if kept := reportSectionOrder(
		oldSections,
		newByID,
	); !slices.Equal(
		kept,
		reportSectionOrder(newSections, oldByID),
	) {
		lines = append(lines, "~ section order changed")
	}

	oldWidgets, newWidgets := reportWidgetsByID(oldSections), reportWidgetsByID(newSections)
	for _, s := range newSections {
		for _, widget := range s.widgets {
			id := reportText(widget["id"])
			old, ok := oldWidgets[id]
			switch {
			case !ok:
				lines = append(
					lines,
					fmt.Sprintf(
						"+ widget %s in %s",
						reportWidgetLabel(widget),
						reportSectionLabel(s),
					),
				)
			case old.section.id != s.id:
				lines = append(lines, fmt.Sprintf(
					"~ widget %s: moved from %s to %s",
					reportWidgetLabel(widget),
					reportSectionLabel(old.section),
					reportSectionLabel(s),
				))
			}
			if changed := reportChangedKeys(old.widget, widget); ok && len(changed) > 0 {
				lines = append(
					lines,
					fmt.Sprintf("~ widget %s: %s", reportWidgetLabel(widget), changed),
				)
			}
		}
	}
	for _, s := range oldSections {
		for _, widget := range s.widgets {
			if _, ok := newWidgets[reportText(widget["id"])]; !ok {
				lines = append(
					lines,
					fmt.Sprintf(
						"- widget %s from %s",
						reportWidgetLabel(widget),
						reportSectionLabel(s),
					),
				)
			}
		}
	}
	for _, s := range newSections {
		old, ok := oldByID[s.id]
		if ok &&
			!slices.Equal(
				reportWidgetOrder(old, s.id, newWidgets),
				reportWidgetOrder(s, s.id, oldWidgets),
			) {
			lines = append(
				lines,
				fmt.Sprintf("~ section %s: widget order changed", reportSectionLabel(s)),
			)
		}
	}
	return lines
}

type reportSection struct {
	id      string
	raw     map[string]any // without widgets
	widgets []map[string]any
}

type reportPlacedWidget struct {
	widget  map[string]any
	section reportSection
}

func reportSections(definition map[string]any) []reportSection {
	items, _ := definition["sections"].([]any)
	var sections []reportSection
	for _, item := range items {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		s := reportSection{id: reportText(raw["id"]), raw: map[string]any{}}
		for k, v := range raw {
			if k != "widgets" {
				s.raw[k] = v
			}
		}
		widgets, _ := raw["widgets"].([]any)
		for _, w := range widgets {
			if widget, ok := w.(map[string]any); ok {
				s.widgets = append(s.widgets, widget)
			}
		}
		sections = append(sections, s)
	}
	return sections
}

func reportSectionsByID(sections []reportSection) map[string]reportSection {
	byID := make(map[string]reportSection, len(sections))
	for _, s := range sections {
		byID[s.id] = s
	}
	return byID
}

func reportWidgetsByID(sections []reportSection) map[string]reportPlacedWidget {
	byID := map[string]reportPlacedWidget{}
	for _, s := range sections {
		for _, w := range s.widgets {
			byID[reportText(w["id"])] = reportPlacedWidget{widget: w, section: s}
		}
	}
	return byID
}

// reportSectionOrder is the order of the sections that are also in other.
func reportSectionOrder(sections []reportSection, other map[string]reportSection) []string {
	var ids []string
	for _, s := range sections {
		if _, ok := other[s.id]; ok {
			ids = append(ids, s.id)
		}
	}
	return ids
}

// reportWidgetOrder is the order of s's widgets that sit in section sectionID on the other side.
func reportWidgetOrder(
	s reportSection,
	sectionID string,
	other map[string]reportPlacedWidget,
) []string {
	var ids []string
	for _, w := range s.widgets {
		id := reportText(w["id"])
		if placed, ok := other[id]; ok && placed.section.id == sectionID {
			ids = append(ids, id)
		}
	}
	return ids
}

// reportChangedKeys names each key whose value differs, with both values for scalars.
func reportChangedKeys(a, b map[string]any) string {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	var changed []string
	for k := range keys {
		before, after := reportJSON(a[k]), reportJSON(b[k])
		if before == after {
			continue
		}
		if reportScalar(a[k]) && reportScalar(b[k]) {
			changed = append(changed, fmt.Sprintf("%s %s → %s", k, before, after))
		} else {
			changed = append(changed, k)
		}
	}
	sort.Strings(changed)
	return strings.Join(changed, ", ")
}

func reportScalar(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return false
	}
	return true
}

func reportSectionLabel(s reportSection) string {
	return fmt.Sprintf("%q (%s)", Printable(reportText(s.raw["title"])), Printable(s.id))
}

func reportWidgetLabel(w map[string]any) string {
	return fmt.Sprintf("%q (%s)", Printable(reportText(w["name"])), Printable(reportText(w["id"])))
}

// reportWidgetSelection reads a widget as [source · metrics · breakdown · chart type].
func reportWidgetSelection(w map[string]any) string {
	source := reportText(w["view"])
	if sources, ok := w["sources"].([]any); ok && len(sources) > 0 {
		names := make([]string, len(sources))
		for i, s := range sources {
			names[i] = reportText(s)
		}
		source += " of " + strings.Join(names, "+")
	}
	parts := []string{source}
	var metrics []string
	items, _ := w["metrics"].([]any)
	for _, m := range items {
		if metric, ok := m.(map[string]any); ok {
			metrics = append(
				metrics,
				reportText(metric["measure"])+"/"+reportText(metric["aggregation"]),
			)
		}
	}
	parts = append(parts, strings.Join(metrics, ", "))
	if dims, ok := w["dimensions"].([]any); ok && len(dims) > 0 {
		if dim, ok := dims[0].(map[string]any); ok {
			parts = append(parts, "by "+reportText(dim["field"]))
		}
	}
	parts = append(parts, reportText(w["chartType"]))
	return Printable("[" + strings.Join(parts, " · ") + "]")
}

func reportText(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if v == nil {
		return ""
	}
	return reportJSON(v)
}

func reportJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(data)
}

func reportStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func reportTerms(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

func reportList(s []any) []any {
	if s == nil {
		return []any{}
	}
	return s
}

// ReportWriteError explains a refused report write; ref names the report saved over and name
// the document's name.
func ReportWriteError(err error, ref, name string) error {
	var refusal *platform.ReportError
	if !errors.As(err, &refusal) {
		return err
	}
	details := refusal.Details
	switch {
	case refusal.Code == platform.ReportChangedCode:
		return fmt.Errorf(
			"the report changed after this document was read: saved by %s at %s via %s; "+
				"run 'iai reports get %q --yaml' and make the change on that, "+
				"or re-run with --force to overwrite it",
			Printable(reportText(details["updated_by"])),
			LocalTime(reportText(details["updated_at"])),
			reportText(details["updated_via"]),
			ref,
		)
	case refusal.Code == platform.ReportNameTakenCode:
		return fmt.Errorf(
			"%s; save under another name, or change that report with 'iai reports update %q --file'",
			refusal.Message,
			name,
		)
	case details["widget_id"] != nil && details["rule"] != nil:
		return fmt.Errorf(
			"%s (rule %s); run 'iai reports schema' for the legal selections",
			refusal.Message, reportText(details["rule"]),
		)
	}
	return err
}

// ReportRestoreError explains a refused restore; ref names the report in the hint.
func ReportRestoreError(err error, ref string) error {
	var refusal *platform.ReportError
	if !errors.As(err, &refusal) || refusal.Code != platform.ReportChangedCode {
		return err
	}
	details := refusal.Details
	return fmt.Errorf(
		"the report was saved again after the expected updated_at: by %s at %s via %s; "+
			"run 'iai reports get %q' to look at it before deciding whether to restore",
		Printable(reportText(details["updated_by"])),
		LocalTime(reportText(details["updated_at"])),
		reportText(details["updated_via"]),
		ref,
	)
}

func reportYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// PrintReportRules prints the widget rules: the sources, every legal selection and the limits.
func PrintReportRules(out io.Writer, rules platform.ReportRules) error {
	fmt.Fprintf(out, "Rules version: %s\n\nSources:\n", rules.Version)
	rows := make([][]string, len(rules.Sources))
	for i, s := range rules.Sources {
		rows[i] = []string{
			s.ID,
			Printable(s.Label),
			reportYesNo(s.Combinable),
			reportYesNo(s.AgentScope),
		}
	}
	if err := PrintTable(
		out,
		[]string{"ID", "LABEL", "COMBINABLE", "AGENT SCOPE"},
		rows,
	); err != nil {
		return err
	}
	if c := rules.Combined; c.View != "" {
		fmt.Fprintf(
			out, "\nCombined (view %s): %d to %d of %s\n",
			c.View, c.MinSources, c.MaxSources, strings.Join(c.Sources, ", "),
		)
	}
	charts := make([]string, len(rules.ChartTypes))
	for i, c := range rules.ChartTypes {
		charts[i] = c.ID
	}
	fmt.Fprintf(out, "Chart types: %s\n\nSelections:\n", strings.Join(charts, ", "))
	rows = make([][]string, len(rules.Selections))
	for i, s := range rules.Selections {
		breakdown := "-"
		if s.Breakdown != nil {
			breakdown = *s.Breakdown
		}
		rows[i] = []string{
			s.View,
			strings.Join(s.Sources, "+"),
			s.Measure + "/" + s.Aggregation,
			breakdown,
			strings.Join(s.ChartTypes, ", "),
			reportYesNo(s.AgentScope),
		}
	}
	headers := []string{"VIEW", "SOURCES", "METRIC", "BREAKDOWN", "CHART TYPES", "AGENT SCOPE"}
	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}
	fmt.Fprintln(out, "\nLimits:")
	w := NewDescribeWriter(out)
	for _, name := range slices.Sorted(maps.Keys(rules.Limits)) {
		fmt.Fprintf(w, "  %s:\t%s\n", name, rules.Limits[name])
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(out, "\nDefault ranges: %s\n", strings.Join(rules.DefaultRanges, ", "))
	fmt.Fprintf(out, "Report filter columns: %s\n", strings.Join(rules.ReportFilterColumns, ", "))
	return nil
}
