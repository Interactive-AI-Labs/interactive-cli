package inputs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"gopkg.in/yaml.v3"
)

// Values of updated_via the CLI sends.
const (
	ReportsViaCLI     = "cli"
	ReportsViaCopilot = "copilot"
)

var DefaultReportColumns = []string{
	"id",
	"name",
	"folder",
	"agents",
	"updated_at",
	"updated_via",
}

var AllReportColumns = []string{
	"id",
	"name",
	"description",
	"folder",
	"agents",
	"default_range",
	"updated_at",
	"updated_by",
	"updated_via",
	"created_at",
}

// ReportsUpdatedVia is the origin a write records: copilot inside the Copilot sandbox, else cli.
func ReportsUpdatedVia(clientOrigin string) string {
	if clientOrigin == ReportsViaCopilot {
		return ReportsViaCopilot
	}
	return ReportsViaCLI
}

// ReportsReadDraft says whether 'iai reports get' reads the caller's open draft: in the Copilot.
func ReportsReadDraft(clientOrigin string) bool {
	return ReportsUpdatedVia(clientOrigin) == ReportsViaCopilot
}

// reportReadOnlyFields are the report keys the API returns that create and update do not send.
var reportReadOnlyFields = []string{
	"id",
	"project_id",
	"can_undo",
	"created_at",
	"created_by",
	"updated_by",
	"updated_via",
	"draft",
	"draft_revision",
}

// ReportDocument is a report as 'iai reports create' and 'update' read it and
// 'iai reports get --yaml' prints it.
type ReportDocument struct {
	UpdatedAt   string `json:"updated_at,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Folder is a folder name, null for no folder, or absent to keep the report's folder.
	Folder       json.RawMessage   `json:"folder,omitempty"`
	AgentIDs     []string          `json:"agent_ids"`
	Terms        map[string]string `json:"terms"`
	DefaultRange string            `json:"default_range"`
	Filters      []any             `json:"filters"`
	Definition   map[string]any    `json:"definition"`
	// AllAgents and ShowTerms absent keep the report's setting.
	AllAgents *bool `json:"all_agents,omitempty"`
	ShowTerms *bool `json:"show_terms,omitempty"`
	// DraftRevision is the document's draft_revision, as 'iai reports get' prints a draft.
	DraftRevision *int64 `json:"-"`
}

// ReadReportDocument reads a report document from path, or from stdin when path is "-".
func ReadReportDocument(path string, stdin io.Reader) (ReportDocument, error) {
	var (
		data []byte
		err  error
	)
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return ReportDocument{}, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return ParseReportDocument(data)
}

// ParseReportDocument reads a YAML or JSON report document, as 'iai reports get' prints
// it or written by hand. Read-only fields are dropped; any other unknown field is an error.
func ParseReportDocument(data []byte) (ReportDocument, error) {
	var doc ReportDocument
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return doc, fmt.Errorf("the report document is empty")
	}
	var fields map[string]any
	var err error
	// JSON keeps big numbers exact; a YAML flow mapping also starts with '{', so it falls back.
	isJSON := false
	if trimmed[0] == '{' {
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		dec.UseNumber()
		isJSON = dec.Decode(&fields) == nil
	}
	if !isJSON {
		fields = nil
		err = yaml.Unmarshal(trimmed, &fields)
		if err == nil {
			fields, _ = stringifyKeys(fields).(map[string]any)
		}
	}
	if err != nil {
		return doc, fmt.Errorf("failed to parse report document: %s", oneLine(err))
	}
	revision, err := reportDraftRevision(fields["draft_revision"])
	if err != nil {
		return doc, err
	}
	for _, key := range reportReadOnlyFields {
		delete(fields, key)
	}
	normalized, err := json.Marshal(fields)
	if err != nil {
		return doc, fmt.Errorf("failed to parse report document: %s", oneLine(err))
	}
	dec := json.NewDecoder(bytes.NewReader(normalized))
	dec.UseNumber()
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return doc, fmt.Errorf("failed to parse report document: %s", oneLine(err))
	}
	doc.DraftRevision = revision
	if len(doc.Folder) > 0 && string(doc.Folder) != "null" {
		var name string
		if json.Unmarshal(doc.Folder, &name) != nil || strings.TrimSpace(name) == "" {
			return doc, fmt.Errorf("folder must be a folder name, or null for no folder")
		}
	}
	return doc, nil
}

// reportDraftRevision is a document's draft_revision, nil when absent.
func reportDraftRevision(v any) (*int64, error) {
	if v == nil {
		return nil, nil
	}
	data, _ := json.Marshal(v)
	var n int64
	if json.Unmarshal(data, &n) != nil {
		return nil, fmt.Errorf(
			"draft_revision must be a whole number; keep it as 'iai reports get --yaml' prints it",
		)
	}
	return &n, nil
}

func oneLine(err error) string {
	return strings.Join(strings.Fields(err.Error()), " ")
}

// ReportFolderChange is what saving doc does to the report's folder: keep it, or move it
// to name ("" for no folder).
func ReportFolderChange(doc ReportDocument) (name string, keep bool) {
	if len(doc.Folder) == 0 {
		return "", true
	}
	_ = json.Unmarshal(doc.Folder, &name)
	return name, false
}

// ReportFolderName is the name of a report's folder, its id when the folder is not listed,
// and empty for no folder.
func ReportFolderName(folderID *string, names map[string]string) string {
	if folderID == nil {
		return ""
	}
	if name, ok := names[*folderID]; ok {
		return name
	}
	return *folderID
}

// ReportBodyOf is the save request for doc: a new report without id, else a replacement of
// report id guarded by the document's updated_at unless force; with the document's
// draft_revision, the replacement goes into the caller's open draft instead.
func ReportBodyOf(
	doc ReportDocument,
	id string,
	force bool,
	updatedVia string,
) platform.ApplyReportBody {
	body := platform.ApplyReportBody{
		UpdatedVia:   updatedVia,
		Name:         doc.Name,
		Description:  doc.Description,
		AgentIDs:     doc.AgentIDs,
		Terms:        doc.Terms,
		DefaultRange: doc.DefaultRange,
		Definition:   doc.Definition,
		Filters:      doc.Filters,
		AllAgents:    doc.AllAgents,
		ShowTerms:    doc.ShowTerms,
	}
	if id != "" {
		body.ID, body.ExpectedUpdatedAt, body.Force = id, doc.UpdatedAt, force
		body.DraftRevision = doc.DraftRevision
	}
	switch name, keep := ReportFolderChange(doc); {
	case keep:
	case name == "":
		body.FolderID = json.RawMessage("null")
	default:
		body.Folder = name
	}
	return body
}

// ReportDocumentOf is a report the API answered as a document: its folder by name (null for
// none) instead of folder_id, without id, and without updated_at unless withUpdatedAt. With
// the caller's draft (?include=draft), the draft's fields replace the report's, with its
// revision as draft_revision.
func ReportDocumentOf(
	raw json.RawMessage,
	folder string,
	withUpdatedAt bool,
) (json.RawMessage, error) {
	var fields map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&fields); err != nil {
		return nil, fmt.Errorf("failed to decode report: %w", err)
	}
	draft, _ := fields["draft"].(map[string]any)
	delete(fields, "draft")
	if document, ok := draft["document"].(map[string]any); ok {
		maps.Copy(fields, document)
		fields["draft_revision"] = draft["revision"]
	}
	delete(fields, "folder_id")
	fields["folder"] = nil
	if folder != "" {
		fields["folder"] = folder
	}
	delete(fields, "id")
	if !withUpdatedAt {
		delete(fields, "updated_at")
	}
	return json.Marshal(fields)
}

// RestoreReportBodyOf is the restore request: guarded by expected when given, else by current.
func RestoreReportBodyOf(
	current string,
	expected *string,
	updatedVia string,
) (platform.RestoreReportBody, error) {
	body := platform.RestoreReportBody{ExpectedUpdatedAt: current, UpdatedVia: updatedVia}
	if expected != nil {
		body.ExpectedUpdatedAt = strings.TrimSpace(*expected)
		if body.ExpectedUpdatedAt == "" {
			return body, fmt.Errorf(
				"--expected-updated-at is empty; leave it out to use the report's current updated_at",
			)
		}
	}
	return body, nil
}

// ResolveReport finds the report whose name, ignoring case, is name; names are unique per project.
func ResolveReport(reports []platform.Report, name string) (platform.Report, error) {
	for _, r := range reports {
		if strings.EqualFold(r.Name, name) {
			return r, nil
		}
	}
	return platform.Report{}, fmt.Errorf("report %q not found; run 'iai reports list'", name)
}

// ResolveReportFolder finds the folder whose name, ignoring case, is ref; "none" names no
// folder unless a folder is called that.
func ResolveReportFolder(folders []platform.ReportFolder, ref string) (string, error) {
	for _, f := range folders {
		if strings.EqualFold(f.Name, ref) {
			return f.ID, nil
		}
	}
	if strings.EqualFold(strings.TrimSpace(ref), "none") {
		return "none", nil
	}
	return "", fmt.Errorf("report folder %q not found; run 'iai reports list'", ref)
}

func ReportFolderNames(folders []platform.ReportFolder) map[string]string {
	names := make(map[string]string, len(folders))
	for _, f := range folders {
		names[f.ID] = f.Name
	}
	return names
}
