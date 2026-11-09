package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestPrintAgentRoutineRun(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")
	tests := []struct {
		name string
		run  deployment.AgentRoutineRun
		want string
	}{
		{
			name: "accepted receipt",
			run: deployment.AgentRoutineRun{
				RunID:     "run_1",
				RoutineID: "refund",
				Status:    "accepted",
				SessionID: "s1",
				CreatedAt: "2026-11-09T10:00:00Z",
			},
			want: "Run:       run_1\nRoutine:   refund\nStatus:    accepted\nSession:   s1\nCreated:   2026-11-09 11:00:00 CET\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintAgentRoutineRun(&out, &tt.run); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoutineFollowCommand(t *testing.T) {
	tests := []struct {
		name, agent, session, org, project, host, deploy, want string
	}{
		{
			name:    "default hostnames omitted",
			agent:   "test-agent",
			session: "s1",
			org:     "platform-development",
			project: "test",
			want:    "iai agents events list test-agent s1 --organization platform-development --project test --follow",
		},
		{
			name:    "only operator override",
			agent:   "test-agent",
			session: "s1",
			org:     "platform-development",
			project: "test",
			deploy:  "http://localhost:8000",
			want:    "iai --deployment-hostname http://localhost:8000 agents events list test-agent s1 --organization platform-development --project test --follow",
		},
		{
			name:    "only platform override",
			agent:   "test-agent",
			session: "s1",
			org:     "platform-development",
			project: "test",
			host:    "https://dev.interactive.ai",
			want:    "iai --hostname https://dev.interactive.ai agents events list test-agent s1 --organization platform-development --project test --follow",
		},
		{
			name:    "local operator and dev platform",
			agent:   "test-agent",
			session: "s1",
			org:     "platform-development",
			project: "test",
			host:    "https://dev.interactive.ai",
			deploy:  "http://localhost:8000",
			want:    "iai --hostname https://dev.interactive.ai --deployment-hostname http://localhost:8000 agents events list test-agent s1 --organization platform-development --project test --follow",
		},
		{
			name:    "shell metacharacters are literal",
			agent:   "agent",
			session: "s'1",
			org:     "Demo Organization",
			project: "$(touch nope)",
			host:    "https://dev.interactive.ai",
			deploy:  "http://localhost:8000",
			want:    "iai --hostname https://dev.interactive.ai --deployment-hostname http://localhost:8000 agents events list agent 's'\\''1' --organization 'Demo Organization' --project '$(touch nope)' --follow",
		},
		{
			name:    "shell operators and expansions are quoted",
			agent:   "agent",
			session: "s1",
			org:     "a;echo nope",
			project: "`whoami`|*",
			host:    "https://dev.interactive.ai?x=1&y=2",
			deploy:  "http://localhost:8000",
			want:    "iai --hostname 'https://dev.interactive.ai?x=1&y=2' --deployment-hostname http://localhost:8000 agents events list agent s1 --organization 'a;echo nope' --project '`whoami`|*' --follow",
		},
		{
			name:    "empty arguments and whitespace are preserved",
			agent:   "agent",
			session: "s1",
			org:     "",
			project: "line\nbreak\tname",
			host:    "https://dev.interactive.ai",
			deploy:  "http://localhost:8000",
			want:    "iai --hostname https://dev.interactive.ai --deployment-hostname http://localhost:8000 agents events list agent s1 --organization '' --project 'line\nbreak\tname' --follow",
		},
		{
			name:    "IPv6 brackets are quoted",
			agent:   "agent",
			session: "s1",
			org:     "platform-development",
			project: "test",
			host:    "https://dev.interactive.ai",
			deploy:  "http://[::1]:8000",
			want:    "iai --hostname https://dev.interactive.ai --deployment-hostname 'http://[::1]:8000' agents events list agent s1 --organization platform-development --project test --follow",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoutineFollowCommand(
				tt.agent,
				tt.session,
				tt.org,
				tt.project,
				tt.host,
				tt.deploy,
			)
			if got != tt.want {
				t.Errorf("command = %q, want %q", got, tt.want)
			}
		})
	}
}
