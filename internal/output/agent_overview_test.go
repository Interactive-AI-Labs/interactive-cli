package output

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func usd(amount string) string {
	return `{"kind":"amount","money":{"amount":"` + amount + `","currency":"USD"}}`
}

func TestPrintAgentOverviewAndPerformance(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")
	window := `"displayName":"cigar","window":{"key":"7d","from":"2026-11-19T10:00:00Z","to":"2026-11-26T10:00:00Z"}`
	cost := `"cost":{"agent":{"value":` + usd(
		"0.13",
	) + `,"previous":` + usd(
		"0.10",
	) + `,"deltaPct":30},` +
		`"interactiveAgent":{"value":` + usd(
		"0.1",
	) + `,"previous":` + usd(
		"0",
	) + `},` +
		`"llm":{"value":{"kind":"unavailable","reason":"no rate"},"previous":` + usd(
		"0",
	) + `}`
	tests := []struct {
		name  string
		print func(io.Writer, []byte) error
		body  string
		want  string
	}{
		{
			name:  "overview with usage and errors",
			print: printOverview,
			body: `{` + window + `,"performance":{"sessions":{"value":4,"previous":2,"deltaPct":100},` +
				`"sessionLatencyMs":{"value":7000},"traceLatencyMs":{"value":750.4},` +
				`"avgTracesPerSession":{"value":1.775}},` + cost + `,"tokens":{"rated":38375}},` +
				`"usage":{"policiesMatched":[{"name":"tone","count":3}],"toolCalls":[{"name":"read_pdf","count":1}]},` +
				`"errors":{"types":[{"observationType":"TOOL","message":"bad file","observationCount":2}],` +
				`"traces":[{"traceId":"t1","errorCount":2}],"sessions":[{"sessionId":"s1","errorCount":1}]}}`,
			want: "cigar, 2026-11-19 11:00:00 CET to 2026-11-26 11:00:00 CET\n\n" +
				"METRIC                 VALUE      PREVIOUS   CHANGE\n" +
				"Sessions               4          2          100.00%\n" +
				"Session latency        7.00s      -          -\n" +
				"Traces                 -          -          -\n" +
				"Trace latency          750ms      -          -\n" +
				"Traces per session     1.77       -          -\n" +
				"Iterations per trace   -          -          -\n" +
				"Observations           -          -          -\n" +
				"Scores                 -          -          -\n" +
				"Total Cost             0.13 USD   0.10 USD   30.00%\n" +
				"Interactive Agent      0.10 USD   0.00 USD   -\n" +
				"LLM Cost               -          0.00 USD   -\n" +
				"Tokens                 38375      -          -\n" +
				"\nUSAGE    NAME       COUNT\n" +
				"Policy   tone       3\n" +
				"Tool     read_pdf   1\n" +
				"\nERRORS        COUNT   NAME\n" +
				"Observation   2       TOOL: bad file\n" +
				"Trace         2       t1\n" +
				"Session       1       s1\n",
		},
		{
			name:  "overview without tokens, usage or errors",
			print: printOverview,
			body:  `{` + window + `,` + cost + `}}`,
			want: "cigar, 2026-11-19 11:00:00 CET to 2026-11-26 11:00:00 CET\n\n" +
				"METRIC                 VALUE      PREVIOUS   CHANGE\n" +
				"Sessions               -          -          -\n" +
				"Session latency        -          -          -\n" +
				"Traces                 -          -          -\n" +
				"Trace latency          -          -          -\n" +
				"Traces per session     -          -          -\n" +
				"Iterations per trace   -          -          -\n" +
				"Observations           -          -          -\n" +
				"Scores                 -          -          -\n" +
				"Total Cost             0.13 USD   0.10 USD   30.00%\n" +
				"Interactive Agent      0.10 USD   0.00 USD   -\n" +
				"LLM Cost               -          0.00 USD   -\n",
		},
		{
			name:  "performance prices each agent by id",
			print: printPerformance,
			body: `{"agents":[{"agent_id":"a","name":"A","traces":3,"sessions":2,"trace_latency":1500,"session_latency":900},` +
				`{"agent_id":"b","name":"B"}],` +
				`"agentCosts":[{"agent_id":"b","total":` + usd(
				"0.131878207",
			) + `}],"costTotal":` + usd(
				"0.1319",
			) + `}`,
			want: "NAME   TRACES   SESSIONS   TRACE LATENCY   SESSION LATENCY   COST\n" +
				"A      3        2          1.50s           900ms             -\n" +
				"B      -        -          -               -                 0.13 USD\n" +
				"\nTotal Cost: 0.13 USD\n",
		},
		{
			name:  "performance without agents",
			print: printPerformance,
			body:  `{"agents":[],"costTotal":` + usd("0") + `}`,
			want:  "NAME   TRACES   SESSIONS   TRACE LATENCY   SESSION LATENCY   COST\n\nTotal Cost: 0.00 USD\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tt.print(&buf, []byte(tt.body)); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func printOverview(out io.Writer, body []byte) error {
	var o platform.AgentOverview
	if err := json.Unmarshal(body, &o); err != nil {
		return err
	}
	return PrintAgentOverview(out, o)
}

func printPerformance(out io.Writer, body []byte) error {
	var p platform.AgentPerformance
	if err := json.Unmarshal(body, &p); err != nil {
		return err
	}
	return PrintAgentPerformance(out, p)
}
