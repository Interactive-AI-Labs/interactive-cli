package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintPromptList(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")

	tests := []struct {
		name    string
		prompts []clients.PromptInfo
		want    string
	}{
		{
			name:    "empty list prints message",
			prompts: []clients.PromptInfo{},
			want:    "No prompts found.\n",
		},
		{
			name:    "nil list prints message",
			prompts: nil,
			want:    "No prompts found.\n",
		},
		{
			name: "single prompt",
			prompts: []clients.PromptInfo{
				{
					Name: "welcome-message",

					Labels:        []string{"production"},
					Tags:          []string{"onboarding"},
					LastUpdatedAt: "2025-01-15T10:30:00Z",
				},
			},
			want: "NAME              LABELS       TAGS         UPDATED\n" +
				"welcome-message   production   onboarding   2025-01-15 11:30:00 CET\n",
		},
		{
			name: "multiple prompts",
			prompts: []clients.PromptInfo{
				{
					Name: "escalation",

					Labels:        nil,
					Tags:          []string{"compliance"},
					LastUpdatedAt: "2025-01-10T08:00:00Z",
				},
				{
					Name: "routing",

					Labels:        []string{"production"},
					Tags:          []string{"core", "routing"},
					LastUpdatedAt: "2025-01-20T14:00:00Z",
				},
			},
			want: "NAME         LABELS       TAGS            UPDATED\n" +
				"escalation                compliance      2025-01-10 09:00:00 CET\n" +
				"routing      production   core, routing   2025-01-20 15:00:00 CET\n",
		},
		{
			name: "truncates long labels list",
			prompts: []clients.PromptInfo{
				{
					Name: "my-prompt",

					Labels:        []string{"production", "staging", "dev", "test"},
					Tags:          nil,
					LastUpdatedAt: "2025-03-01T12:00:00Z",
				},
			},
			want: "NAME        LABELS                               TAGS   UPDATED\n" +
				"my-prompt   production, staging, dev (+1 more)          2025-03-01 13:00:00 CET\n",
		},
		{
			name: "folder rows display trailing slash",
			prompts: []clients.PromptInfo{
				{
					Name:    "team-a",
					RowType: "folder",
				},
				{
					Name:          "faq-lookup",
					Labels:        []string{"production", "latest"},
					Tags:          nil,
					LastUpdatedAt: "2025-03-01T12:00:00Z",
				},
			},
			want: "NAME         LABELS               TAGS   UPDATED\n" +
				"team-a/                                  \n" +
				"faq-lookup   production, latest          2025-03-01 13:00:00 CET\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintPromptList(&buf, tt.prompts)
			if err != nil {
				t.Fatalf("PrintPromptList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintPromptVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []int
		want     string
	}{
		{
			name:     "empty list prints message",
			versions: []int{},
			want:     "No versions found.\n",
		},
		{
			name:     "nil list prints message",
			versions: nil,
			want:     "No versions found.\n",
		},
		{
			name:     "single version",
			versions: []int{1},
			want: "VERSION\n" +
				"1\n",
		},
		{
			name:     "multiple versions sorted descending",
			versions: []int{1, 3, 2},
			want: "VERSION\n" +
				"3\n" +
				"2\n" +
				"1\n",
		},
		{
			name:     "already sorted input still sorted descending",
			versions: []int{5, 4, 3, 2, 1},
			want: "VERSION\n" +
				"5\n" +
				"4\n" +
				"3\n" +
				"2\n" +
				"1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintPromptVersions(&buf, tt.versions)
			if err != nil {
				t.Fatalf("PrintPromptVersions() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintPromptDiff(t *testing.T) {
	tests := []struct {
		name     string
		versionA string
		a        *clients.PromptDetail
		versionB string
		b        *clients.PromptDetail
		want     string
	}{
		{
			name:     "identical content shows no differences",
			versionA: "1",
			a:        &clients.PromptDetail{Prompt: json.RawMessage(`"Hello world"`)},
			versionB: "2",
			b:        &clients.PromptDetail{Prompt: json.RawMessage(`"Hello world"`)},
			want:     "No differences found.\n",
		},
		{
			name:     "different string content shows diff",
			versionA: "1",
			a:        &clients.PromptDetail{Prompt: json.RawMessage(`"Hello world"`)},
			versionB: "2",
			b:        &clients.PromptDetail{Prompt: json.RawMessage(`"Hello universe"`)},
			want: "--- version 1\n" +
				"+++ version 2\n" +
				"@@ -1,1 +1,1 @@\n" +
				"-Hello world\n" +
				"+Hello universe\n",
		},
		{
			name:     "structured JSON content shows diff",
			versionA: "3",
			a: &clients.PromptDetail{
				Prompt: json.RawMessage(`{"role":"system","content":"Hi"}`),
			},
			versionB: "4",
			b: &clients.PromptDetail{
				Prompt: json.RawMessage(`{"role":"system","content":"Bye"}`),
			},
			want: "--- version 3\n" +
				"+++ version 4\n" +
				"@@ -1,4 +1,4 @@\n" +
				" {\n" +
				"   \"role\": \"system\",\n" +
				"-  \"content\": \"Hi\"\n" +
				"+  \"content\": \"Bye\"\n" +
				" }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintPromptDiff(&buf, tt.versionA, tt.a, tt.versionB, tt.b)
			if err != nil {
				t.Fatalf("PrintPromptDiff() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintPromptDetail(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")

	tests := []struct {
		name   string
		prompt *clients.PromptDetail
		want   string
	}{
		{
			name: "full detail with all fields",
			prompt: &clients.PromptDetail{
				Name:      "onboarding-flow",
				Type:      "routine",
				Version:   3,
				Labels:    []string{"production", "staging"},
				Tags:      []string{"v2", "experimental"},
				CreatedAt: "2025-01-10T08:00:00Z",
				UpdatedAt: "2025-01-15T10:30:00Z",
				Prompt: json.RawMessage(
					`"steps:\n  - type: action\n    name: greet\n  - type: finish"`,
				),
			},
			want: "Name:         onboarding-flow\n" +
				"Version:      3\n" +
				"Labels:       production, staging\n" +
				"Tags:         v2, experimental\n" +
				"Created At:   2025-01-10 09:00:00 CET\n" +
				"Updated At:   2025-01-15 11:30:00 CET\n" +
				"\n" +
				"Content:\n" +
				"steps:\n" +
				"  - type: action\n" +
				"    name: greet\n" +
				"  - type: finish\n",
		},
		{
			name: "minimal detail without optional fields",
			prompt: &clients.PromptDetail{
				Name:    "welcome-message",
				Type:    "text",
				Version: 1,
				Prompt:  json.RawMessage(`"You are a helpful assistant."`),
			},
			want: "Name:      welcome-message\n" +
				"Version:   1\n" +
				"\n" +
				"Content:\n" +
				"You are a helpful assistant.\n",
		},
		{
			name: "non-string JSON prompt prints raw",
			prompt: &clients.PromptDetail{
				Name:    "structured-prompt",
				Type:    "chat",
				Version: 1,
				Prompt:  json.RawMessage(`[{"role":"system","content":"Hello"}]`),
			},
			want: "Name:      structured-prompt\n" +
				"Version:   1\n" +
				"\n" +
				"Content:\n" +
				`[{"role":"system","content":"Hello"}]` + "\n",
		},
		{
			name: "prompt content ending with newline has no double newline",
			prompt: &clients.PromptDetail{
				Name:    "trailing-nl",
				Type:    "text",
				Version: 1,
				Prompt:  json.RawMessage(`"Already has newline\n"`),
			},
			want: "Name:      trailing-nl\n" +
				"Version:   1\n" +
				"\n" +
				"Content:\n" +
				"Already has newline\n",
		},
		{
			name: "detail without content",
			prompt: &clients.PromptDetail{
				Name:    "empty-prompt",
				Type:    "text",
				Version: 1,
				Labels:  []string{"draft"},
				Prompt:  nil,
			},
			want: "Name:      empty-prompt\n" +
				"Version:   1\n" +
				"Labels:    draft\n",
		},
		{
			name: "skill with config block",
			prompt: &clients.PromptDetail{
				Name:    "summarize-trace",
				Type:    "skill",
				Version: 2,
				Labels:  []string{"production"},
				Config: json.RawMessage(
					`{"skill":{"description":"Summarize a Langfuse trace","intents":["summarize trace","explain trace"]}}`,
				),
				Prompt: json.RawMessage(`"# Summarize Trace\n\nDo the thing."`),
			},
			want: "Name:          summarize-trace\n" +
				"Version:       2\n" +
				"Labels:        production\n" +
				"Description:   Summarize a Langfuse trace\n" +
				"Intents:       summarize trace, explain trace\n" +
				"\n" +
				"Content:\n" +
				"# Summarize Trace\n" +
				"\n" +
				"Do the thing.\n",
		},
		{
			// Existing typed prompts (routines, macros, etc.) often have an
			// empty config object. The renderer must NOT print a Config
			// block in that case — pre-PR behavior must be preserved.
			name: "empty config object is not rendered",
			prompt: &clients.PromptDetail{
				Name:    "no-config-prompt",
				Type:    "macro",
				Version: 1,
				Prompt:  json.RawMessage(`"id: x\ntext: hi"`),
				Config:  json.RawMessage(`{}`),
			},
			want: "Name:      no-config-prompt\n" +
				"Version:   1\n" +
				"\n" +
				"Content:\n" +
				"id: x\n" +
				"text: hi\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintPromptDetail(&buf, tt.prompt)
			if err != nil {
				t.Fatalf("PrintPromptDetail() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
