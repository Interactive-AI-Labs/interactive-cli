package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/google/go-querystring/query"
)

// Error codes of the reports API the CLI explains.
const (
	ReportChangedCode   = "REPORT_CHANGED"
	ReportNameTakenCode = "REPORT_NAME_TAKEN"
)

// Report is a saved report as the platform returns it.
type Report struct {
	ID           string            `json:"id"`
	ProjectID    string            `json:"project_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	FolderID     *string           `json:"folder_id"`
	AgentIDs     []string          `json:"agent_ids"`
	DefaultRange string            `json:"default_range"`
	CanUndo      bool              `json:"can_undo"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
	UpdatedBy    string            `json:"updated_by"`
	UpdatedVia   string            `json:"updated_via"`
	Terms        map[string]string `json:"terms,omitempty"`
	Definition   map[string]any    `json:"definition,omitempty"`
	Filters      []any             `json:"filters,omitempty"`
	// Draft is true when a Copilot write went into the caller's open draft, not the report.
	Draft bool `json:"draft"`
}

// ReportDraft is the caller's unsaved edit of a report, open in the editor.
type ReportDraft struct {
	Document Report `json:"document"`
}

type ReportFolder struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// DefaultReportRange is the range the platform saves when a report names none.
const DefaultReportRange = "last7Days"

// ApplyReportBody is the body of POST …/reports/apply: with an id it replaces that report.
type ApplyReportBody struct {
	ID                string `json:"id,omitempty"`
	ExpectedUpdatedAt string `json:"expected_updated_at,omitempty"`
	Force             bool   `json:"force,omitempty"`
	UpdatedVia        string `json:"updated_via"`
	// FolderID is absent to keep the folder, null to unfile, or a quoted id to move.
	FolderID     json.RawMessage   `json:"folder_id,omitempty"`
	Folder       string            `json:"folder,omitempty"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	AgentIDs     []string          `json:"agent_ids,omitempty"`
	AllAgents    *bool             `json:"all_agents,omitempty"`
	Terms        map[string]string `json:"terms,omitempty"`
	ShowTerms    *bool             `json:"show_terms,omitempty"`
	DefaultRange string            `json:"default_range,omitempty"`
	Definition   map[string]any    `json:"definition"`
	Filters      []any             `json:"filters,omitempty"`
}

// RestoreReportBody is the body of POST …/reports/{id}/undo.
type RestoreReportBody struct {
	ExpectedUpdatedAt string `json:"expected_updated_at"`
	UpdatedVia        string `json:"updated_via"`
}

// ReportError is a refusal of the reports API, with its code and details.
type ReportError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *ReportError) Error() string { return e.Message }

// reportRefusal reads a refusal of action: {"detail": {"error": …}} or {"error": …}.
func reportRefusal(action string) func(status int, body []byte) error {
	return func(status int, body []byte) error {
		type refusal struct {
			Code    string         `json:"code"`
			Details map[string]any `json:"details"`
		}
		var parsed struct {
			Detail struct {
				Error refusal `json:"error"`
			} `json:"detail"`
			Error refusal `json:"error"`
		}
		_ = json.Unmarshal(body, &parsed)
		r := parsed.Detail.Error
		if r.Code == "" {
			r = parsed.Error
		}
		message := clients.ExtractServerMessage(body)
		if message == "" {
			message = fmt.Sprintf(
				"failed to %s: server returned %d %s", action, status, http.StatusText(status),
			)
		}
		return &ReportError{Code: r.Code, Message: message, Details: r.Details}
	}
}

func reportsPath(projectID string, segments ...string) string {
	path := fmt.Sprintf("/api/platform/v1/projects/%s/reports", url.PathEscape(projectID))
	for _, s := range segments {
		path += "/" + url.PathEscape(s)
	}
	return path
}

// reportsCall sends a reports request and answers its data decoded into T, and as sent.
// Numbers keep their exact value, so a document printed back is the one saved.
func reportsCall[T any](
	c *APIClient,
	ctx context.Context,
	method, path string,
	opts any,
	body any,
	action string,
) (T, json.RawMessage, error) {
	var zero T
	req, err := c.newJSONRequest(ctx, method, path, body)
	if err != nil {
		return zero, nil, err
	}
	q, err := query.Values(opts)
	if err != nil {
		return zero, nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}
	req.URL.RawQuery = q.Encode()
	respBody, err := c.doAndReadWith(req, action, reportRefusal(action))
	if err != nil {
		return zero, nil, err
	}
	data, err := decodeSuccess[json.RawMessage](respBody, action)
	if err != nil {
		return zero, nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&zero); err != nil {
		return zero, nil, fmt.Errorf("failed to decode %s response: %w", action, err)
	}
	return zero, data, nil
}

// ReportListOptions filter GET …/reports; FolderID "none" lists those in no folder.
type ReportListOptions struct {
	FolderID string `url:"folder_id,omitempty"`
	Agent    string `url:"agent,omitempty"`
}

func (c *APIClient) ListReports(
	ctx context.Context,
	projectID string,
	opts ReportListOptions,
) ([]Report, json.RawMessage, error) {
	return reportsCall[[]Report](
		c, ctx, http.MethodGet, reportsPath(projectID), opts, nil, "list reports",
	)
}

func (c *APIClient) GetReport(
	ctx context.Context,
	projectID, reportID string,
) (Report, json.RawMessage, error) {
	return reportsCall[Report](
		c, ctx, http.MethodGet, reportsPath(projectID, reportID), nil, nil, "get report",
	)
}

// reportWithDraft is GET …/reports/{id}?include=draft; its Draft shadows Report.Draft.
type reportWithDraft struct {
	Report
	Draft *ReportDraft `json:"draft"`
}

type reportGetOptions struct {
	Include string `url:"include,omitempty"`
}

// GetReportWithDraft reads a report and the caller's draft of it, nil when there is none.
func (c *APIClient) GetReportWithDraft(
	ctx context.Context,
	projectID, reportID string,
) (Report, *ReportDraft, json.RawMessage, error) {
	got, raw, err := reportsCall[reportWithDraft](
		c, ctx, http.MethodGet, reportsPath(projectID, reportID),
		reportGetOptions{Include: "draft"}, nil, "get report",
	)
	return got.Report, got.Draft, raw, err
}

func (c *APIClient) ApplyReport(
	ctx context.Context,
	projectID string,
	body ApplyReportBody,
) (Report, error) {
	report, _, err := reportsCall[Report](
		c, ctx, http.MethodPost, reportsPath(projectID, "apply"), nil, body, "save report",
	)
	return report, err
}

// RestoreReport swaps a report with its previous version.
func (c *APIClient) RestoreReport(
	ctx context.Context,
	projectID, reportID string,
	body RestoreReportBody,
) (Report, error) {
	report, _, err := reportsCall[Report](
		c, ctx, http.MethodPost, reportsPath(projectID, reportID, "undo"), nil, body,
		"restore report",
	)
	return report, err
}

// DeleteReport deletes a report and answers it as it was, as sent.
func (c *APIClient) DeleteReport(
	ctx context.Context,
	projectID, reportID string,
) (json.RawMessage, error) {
	_, raw, err := reportsCall[json.RawMessage](
		c, ctx, http.MethodDelete, reportsPath(projectID, reportID), nil, nil, "delete report",
	)
	return raw, err
}

func (c *APIClient) ListReportFolders(
	ctx context.Context,
	projectID string,
) ([]ReportFolder, error) {
	folders, _, err := reportsCall[[]ReportFolder](
		c, ctx, http.MethodGet, reportsPath(projectID, "folders"), nil, nil, "list report folders",
	)
	return folders, err
}

// ReportRules are the legal widget selections and the report limits.
type ReportRules struct {
	Version json.Number `json:"version"`
	Sources []struct {
		ID         string `json:"id"`
		Label      string `json:"label"`
		AgentScope bool   `json:"agent_scope"`
		Combinable bool   `json:"combinable"`
	} `json:"sources"`
	Combined struct {
		View       string   `json:"view"`
		Sources    []string `json:"sources"`
		MinSources int      `json:"min_sources"`
		MaxSources int      `json:"max_sources"`
	} `json:"combined"`
	ChartTypes []struct {
		ID string `json:"id"`
	} `json:"chart_types"`
	Selections []struct {
		View        string   `json:"view"`
		Sources     []string `json:"sources"`
		Measure     string   `json:"measure"`
		Aggregation string   `json:"aggregation"`
		Breakdown   *string  `json:"breakdown"`
		ChartTypes  []string `json:"chart_types"`
		AgentScope  bool     `json:"agent_scope"`
	} `json:"selections"`
	Limits              map[string]json.Number `json:"limits"`
	DefaultRanges       []string               `json:"default_ranges"`
	ReportFilterColumns []string               `json:"report_filter_columns"`
}

func (c *APIClient) GetReportWidgetRules(
	ctx context.Context,
	projectID string,
) (ReportRules, json.RawMessage, error) {
	return reportsCall[ReportRules](
		c,
		ctx,
		http.MethodGet,
		reportsPath(projectID, "widget-rules"),
		nil,
		nil,
		"get widget rules",
	)
}
