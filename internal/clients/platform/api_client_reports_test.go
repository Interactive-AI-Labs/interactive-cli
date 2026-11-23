package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-querystring/query"
)

func TestReportListQuery(t *testing.T) {
	tests := []struct {
		opts any
		want string
	}{
		{nil, ""},
		{ReportListOptions{}, ""},
		{ReportListOptions{FolderID: "f-1"}, "folder_id=f-1"},
		{
			ReportListOptions{FolderID: "none", Agent: "support-bot"},
			"agent=support-bot&folder_id=none",
		},
		{reportGetOptions{Include: "draft"}, "include=draft"},
	}
	for _, tt := range tests {
		q, err := query.Values(tt.opts)
		if got := q.Encode(); err != nil || got != tt.want {
			t.Errorf("query.Values(%+v) = %q, %v; want %q", tt.opts, got, err, tt.want)
		}
	}
}

// A get with ?include=draft reads the draft object; an apply reads draft as a flag.
func TestReportDraftDecode(t *testing.T) {
	tests := []struct {
		name, raw, want string
	}{
		{
			"get with a draft",
			`{"id":"r-1","folder_id":"f-1","draft":{"document":{"name":"B","folder_id":"f-2"}}}`,
			`r-1 false f-2`,
		},
		{"get with no draft", `{"id":"r-1","folder_id":"f-1","draft":null}`, `r-1 false <nil>`},
	}
	for _, tt := range tests {
		var got reportWithDraft
		err := json.Unmarshal([]byte(tt.raw), &got)
		folder := "<nil>"
		if got.Draft != nil {
			folder = *got.Draft.Document.FolderID
		}
		if s := fmt.Sprint(got.ID, " ", got.Report.Draft, " ", folder); err != nil || s != tt.want {
			t.Errorf("%s: got %s, %v; want %s", tt.name, s, err, tt.want)
		}
	}
	var applied Report
	if err := json.Unmarshal([]byte(`{"id":"r-1","draft":true}`), &applied); err != nil ||
		!applied.Draft {
		t.Errorf("apply: got %+v, %v; want draft true", applied, err)
	}
}

func TestReportRefusal(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "platform error in detail",
			status: 409,
			body:   `{"detail":{"success":false,"error":{"code":"REPORT_CHANGED","message":"changed","details":{"updated_by":"u-9"}}}}`,
			want:   `{"Code":"REPORT_CHANGED","Message":"changed","Details":{"updated_by":"u-9"}}`,
		},
		{
			name:   "schema errors are listed",
			status: 422,
			body:   `{"detail":{"error":{"code":"REPORT_INVALID","message":"invalid","details":{"schema_errors":[{"path":"name","message":"required"}]}}}}`,
			want:   `{"Code":"REPORT_INVALID","Message":"invalid\n  - name: required","Details":{"schema_errors":[{"message":"required","path":"name"}]}}`,
		},
		{
			name:   "platform error",
			status: 400,
			body:   `{"success":false,"error":{"code":"INVALID_REQUEST","message":"bad"}}`,
			want:   `{"Code":"INVALID_REQUEST","Message":"bad","Details":null}`,
		},
		{
			name:   "fastapi detail",
			status: 404,
			body:   `{"detail":"Report not found"}`,
			want:   `{"Code":"","Message":"Report not found","Details":null}`,
		},
		{
			name:   "no body",
			status: 502,
			body:   ``,
			want:   `{"Code":"","Message":"failed to delete report: server returned 502 Bad Gateway","Details":null}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := json.Marshal(reportRefusal("delete report")(tt.status, []byte(tt.body)))
			if string(got) != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

// The data comes back as sent, and decoded with numbers kept exact.
func TestReportsCallKeepsNumbers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{"n":12345678901234567,"f":0.5}}`))
	}))
	defer server.Close()
	got, raw, err := reportsCall[map[string]any](
		newEvalTestClient(
			t,
			server.URL,
		),
		context.Background(),
		http.MethodGet,
		"/r",
		nil,
		nil,
		"get report",
	)
	decoded, _ := json.Marshal(got)
	if want := `{"f":0.5,"n":12345678901234567} {"n":12345678901234567,"f":0.5} <nil>`; string(
		decoded,
	)+" "+string(
		raw,
	)+" "+fmt.Sprint(
		err,
	) != want {
		t.Errorf("got %s %s %v, want %s", decoded, raw, err, want)
	}
}
