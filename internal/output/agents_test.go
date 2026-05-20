package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintAgentDescribe(t *testing.T) {
	tests := []struct {
		name  string
		agent *clients.DescribeAgentResponse
		want  string
	}{
		{
			name: "minimal agent",
			agent: &clients.DescribeAgentResponse{
				Name:     "minimal-agent",
				Id:       "interactive-agent",
				Version:  "0.1.0",
				Revision: 1,
				Status:   "deployed",
			},
			want: "Name:       minimal-agent\n" +
				"Id:         interactive-agent\n" +
				"Version:    0.1.0\n" +
				"Revision:   1\n" +
				"Status:     deployed\n",
		},
		{
			name: "agent with message",
			agent: &clients.DescribeAgentResponse{
				Name:     "msg-agent",
				Id:       "interactive-agent",
				Version:  "0.1.0",
				Revision: 2,
				Status:   "deployed",
				Message:  "rollout in progress",
			},
			want: "Name:       msg-agent\n" +
				"Id:         interactive-agent\n" +
				"Version:    0.1.0\n" +
				"Revision:   2\n" +
				"Status:     deployed\n" +
				"Message:    rollout in progress\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintAgentDescribe(&buf, tt.agent)
			if err != nil {
				t.Fatalf("PrintAgentDescribe() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestPrintAgentList(t *testing.T) {
	tests := []struct {
		name   string
		agents []clients.AgentOutput
		want   string
	}{
		{
			name:   "empty list prints message",
			agents: []clients.AgentOutput{},
			want:   "No agents found.\n",
		},
		{
			name:   "nil list prints message",
			agents: nil,
			want:   "No agents found.\n",
		},
		{
			name: "single agent",
			agents: []clients.AgentOutput{
				{Name: "my-agent", Revision: 3, Status: "deployed", Updated: "2024-06-01"},
			},
			want: "NAME       REVISION   STATUS     UPDATED\n" +
				"my-agent   3          deployed   2024-06-01\n",
		},
		{
			name: "multiple agents",
			agents: []clients.AgentOutput{
				{Name: "agent-a", Revision: 1, Status: "deployed"},
				{Name: "agent-b", Revision: 5, Status: "deploying"},
			},
			want: "NAME      REVISION   STATUS      UPDATED\n" +
				"agent-a   1          deployed    \n" +
				"agent-b   5          deploying   \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintAgentList(&buf, tt.agents)
			if err != nil {
				t.Fatalf("PrintAgentList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintAgentRevisions(t *testing.T) {
	tests := []struct {
		name      string
		revisions []clients.RevisionMeta
		want      string
	}{
		{
			name:      "empty list prints message",
			revisions: []clients.RevisionMeta{},
			want:      "No revisions found.\n",
		},
		{
			name: "single revision gets star marker",
			revisions: []clients.RevisionMeta{
				{Revision: 1, Updated: "2024-01-01"},
			},
			want: "    REVISION   UPDATED\n" +
				"*   1          2024-01-01\n",
		},
		{
			name: "latest revision gets star marker",
			revisions: []clients.RevisionMeta{
				{Revision: 1, Updated: "2024-01-01"},
				{Revision: 3, Updated: "2024-03-01"},
				{Revision: 2, Updated: "2024-02-01"},
			},
			want: "    REVISION   UPDATED\n" +
				"    1          2024-01-01\n" +
				"*   3          2024-03-01\n" +
				"    2          2024-02-01\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintAgentRevisions(&buf, tt.revisions)
			if err != nil {
				t.Fatalf("PrintAgentRevisions() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintAgentRevision(t *testing.T) {
	tests := []struct {
		name string
		rev  *clients.AgentRevisionResponse
		want string
	}{
		{
			name: "minimal revision",
			rev: &clients.AgentRevisionResponse{
				RevisionMeta: clients.RevisionMeta{Revision: 3, Status: "deployed"},
				Id:           "interactive-agent",
				Version:      "1.0.0",
			},
			want: "Revision:   3\n" +
				"Status:     deployed\n" +
				"Id:         interactive-agent\n" +
				"Version:    1.0.0\n",
		},
		{
			name: "revision with endpoint and env",
			rev: &clients.AgentRevisionResponse{
				RevisionMeta: clients.RevisionMeta{
					Revision: 5,
					Status:   "deployed",
					Updated:  "2024-06-01",
				},
				Id:       "interactive-agent",
				Version:  "2.0.0",
				Endpoint: "my-agent.interactive.ai",
				Env: []clients.EnvVar{
					{Name: "LOG_LEVEL", Value: "debug"},
				},
			},
			want: "Revision:   5\n" +
				"Status:     deployed\n" +
				"Updated:    2024-06-01\n" +
				"Id:         interactive-agent\n" +
				"Version:    2.0.0\n" +
				"Endpoint:   my-agent.interactive.ai\n" +
				"\n" +
				"Environment:\n" +
				"  LOG_LEVEL=debug\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintAgentRevision(&buf, tt.rev)
			if err != nil {
				t.Fatalf("PrintAgentRevision() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestPrintAgentCatalog(t *testing.T) {
	tests := []struct {
		name   string
		agents []clients.CatalogAgent
		want   string
	}{
		{
			name:   "empty list prints message",
			agents: []clients.CatalogAgent{},
			want:   "No agents available.\n",
		},
		{
			name:   "nil list prints message",
			agents: nil,
			want:   "No agents available.\n",
		},
		{
			name: "single agent",
			agents: []clients.CatalogAgent{
				{Id: "interactive-agent"},
			},
			want: "AGENT ID\n" +
				"interactive-agent\n",
		},
		{
			name: "multiple agents",
			agents: []clients.CatalogAgent{
				{Id: "interactive-agent"},
				{Id: "other-agent"},
			},
			want: "AGENT ID\n" +
				"interactive-agent\n" +
				"other-agent\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintAgentCatalog(&buf, tt.agents)
			if err != nil {
				t.Fatalf("PrintAgentCatalog() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintCompatibilityMatrix(t *testing.T) {
	tests := []struct {
		name   string
		matrix []clients.CompatibilityEntry
		asJSON bool
		want   string
	}{
		{
			name:   "empty matrix prints message",
			matrix: []clients.CompatibilityEntry{},
			want:   "No compatibility data available.\n",
		},
		{
			name: "table output",
			matrix: []clients.CompatibilityEntry{
				{AgentVersion: "0.1.0", SchemaVersion: "1.0.0"},
				{AgentVersion: "0.2.0", SchemaVersion: "2.0.0"},
			},
			want: "AGENT VERSION   SCHEMA VERSION\n" +
				"0.1.0           1.0.0\n" +
				"0.2.0           2.0.0\n",
		},
		{
			name: "json output",
			matrix: []clients.CompatibilityEntry{
				{AgentVersion: "0.1.0", SchemaVersion: "1.0.0"},
			},
			asJSON: true,
			want: "[\n  {\n    \"agentVersion\": \"0.1.0\",\n" +
				"    \"schemaVersion\": \"1.0.0\"\n  }\n]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintCompatibilityMatrix(&buf, tt.matrix, tt.asJSON)
			if err != nil {
				t.Fatalf("PrintCompatibilityMatrix() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintAgentVersions(t *testing.T) {
	tests := []struct {
		name     string
		agentId  string
		versions []string
		want     string
	}{
		{
			name:     "empty versions prints message",
			agentId:  "interactive-agent",
			versions: []string{},
			want:     "No versions found for agent \"interactive-agent\".\n",
		},
		{
			name:     "nil versions prints message",
			agentId:  "interactive-agent",
			versions: nil,
			want:     "No versions found for agent \"interactive-agent\".\n",
		},
		{
			name:     "single version",
			agentId:  "interactive-agent",
			versions: []string{"0.0.1"},
			want: "VERSION\n" +
				"0.0.1\n",
		},
		{
			name:     "multiple versions",
			agentId:  "interactive-agent",
			versions: []string{"0.0.1", "0.0.2", "0.1.0"},
			want: "VERSION\n" +
				"0.0.1\n" +
				"0.0.2\n" +
				"0.1.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintAgentVersions(&buf, tt.agentId, tt.versions)
			if err != nil {
				t.Fatalf("PrintAgentVersions() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
