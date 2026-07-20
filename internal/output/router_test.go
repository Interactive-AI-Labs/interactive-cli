package output

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewRouterInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "production",
			input: "https://app.interactive.ai",
			want:  "https://app.interactive.ai/api/v1",
		},
		{
			name:  "trailing slash",
			input: "https://dev.interactive.ai/",
			want:  "https://dev.interactive.ai/api/v1",
		},
		{name: "missing scheme", input: "app.interactive.ai", wantErr: true},
		{name: "path", input: "https://app.interactive.ai/root", wantErr: true},
		{name: "query string", input: "https://app.interactive.ai?region=eu", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRouterInfo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewRouterInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got.BaseURL); diff != "" {
				t.Errorf("base URL mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPrintRouterInfo(t *testing.T) {
	info, err := NewRouterInfo("https://app.interactive.ai")
	if err != nil {
		t.Fatalf("NewRouterInfo() error = %v", err)
	}

	tests := []struct {
		name  string
		print func(*bytes.Buffer) error
		want  string
	}{
		{
			name:  "human readable",
			print: func(out *bytes.Buffer) error { return PrintRouterInfo(out, info) },
			want: `Base URL:        https://app.interactive.ai/api/v1
Documentation:   https://docs.interactive.ai/llm-router
Endpoints:
  Chat Completions:
    Method:        POST
    URL:           https://app.interactive.ai/api/v1/chat/completions
    Description:   Generate chat responses, optionally returning them as they are created.
  Responses:
    Method:        POST
    URL:           https://app.interactive.ai/api/v1/responses
    Description:   Generate model responses from text, messages, or tool results.
  Embeddings:
    Method:        POST
    URL:           https://app.interactive.ai/api/v1/embeddings
    Description:   Convert text into numbers that can be compared by meaning.
  Rerank:
    Method:        POST
    URL:           https://app.interactive.ai/api/v1/rerank
    Description:   Sort candidate documents by their relevance to a query.
  Models:
    Method:        GET
    URL:           https://app.interactive.ai/api/v1/models
    Description:   List models available for the request region.
  Health:
    Method:        GET
    URL:           https://app.interactive.ai/api/v1/health/llm-router
    Description:   Check whether the inference router is healthy.
`,
		},
		{
			name:  "JSON",
			print: func(out *bytes.Buffer) error { return PrintStructuredJSON(out, info) },
			want: `{
  "baseUrl": "https://app.interactive.ai/api/v1",
  "endpoints": {
    "chatCompletions": {
      "method": "POST",
      "url": "https://app.interactive.ai/api/v1/chat/completions",
      "description": "Generate chat responses, optionally returning them as they are created."
    },
    "responses": {
      "method": "POST",
      "url": "https://app.interactive.ai/api/v1/responses",
      "description": "Generate model responses from text, messages, or tool results."
    },
    "embeddings": {
      "method": "POST",
      "url": "https://app.interactive.ai/api/v1/embeddings",
      "description": "Convert text into numbers that can be compared by meaning."
    },
    "rerank": {
      "method": "POST",
      "url": "https://app.interactive.ai/api/v1/rerank",
      "description": "Sort candidate documents by their relevance to a query."
    },
    "models": {
      "method": "GET",
      "url": "https://app.interactive.ai/api/v1/models",
      "description": "List models available for the request region."
    },
    "health": {
      "method": "GET",
      "url": "https://app.interactive.ai/api/v1/health/llm-router",
      "description": "Check whether the inference router is healthy."
    }
  },
  "documentationUrl": "https://docs.interactive.ai/llm-router"
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := tt.print(&out); err != nil {
				t.Fatalf("print() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, out.String()); diff != "" {
				t.Errorf("output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
