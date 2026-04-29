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
