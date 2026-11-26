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

func TestPrintAgentActivity(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")
	tests := []struct {
		name string
		body string
		want string
	}{
		{"empty", `{"activity":[]}`, "No activity found.\n"},
		{
			"deploy and a long update",
			`{"activity":[` +
				`{"time":"2026-11-05T09:52:07.013933Z","action":"deploy","success":true,` +
				`"data":{"revision":34},"actor":{"id":"user:1","display":"Ana"}},` +
				`{"time":"2026-11-05T09:50:53Z","action":"update","success":false,` +
				`"data":{"changes":{"context.preamble":{"from":null,"to":{"examples":["one","two","three","four"]}}}},` +
				`"actor":{"id":"system:rollout","display":"System"}}],` +
				`"meta":{"page":1,"limit":2,"total_items":42,"total_pages":21}}`,
			"TIME                      ACTION   ACTOR    SUCCESS   DATA\n" +
				"2026-11-05 10:52:07 CET   deploy   Ana      true      {\"revision\":34}\n" +
				"2026-11-05 10:50:53 CET   update   System   false     " +
				"{\"changes\":{\"context.preamble\":{\"from\":null,\"to\":{\"examples\":[\"one\",\"two\",\"th...\n" +
				"\nPage 1 of 21 (42 total items)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a platform.AgentActivity
			if err := json.Unmarshal([]byte(tt.body), &a); err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			if err := PrintAgentActivity(&buf, a); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestPrintAgentAnalytics(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			"empty",
			`{"cpu_usage":{"chart_data":[]},"latency":{"chart_data":[{"timestamp":1,"values":{}}]}}`,
			"No analytics found.\n",
		},
		{
			"latest and peak per series",
			`{"cpu_usage":{"chart_data":[` +
				`{"timestamp":1,"values":{"bot":0.012,"bot-db":0.5}},{"timestamp":2,"values":{"bot":0.02}}]},` +
				`"memory_usage":{"chart_data":[` +
				`{"timestamp":1,"values":{"bot":104857600}},{"timestamp":2,"values":{"bot":52428800}}]},` +
				`"status_codes":{"chart_data":[` +
				`{"timestamp":1,"values":{"2xx":12,"5xx":1}},{"timestamp":2,"values":{"2xx":30.333,"5xx":0}}]},` +
				`"requests_per_domain":{"chart_data":[` +
				`{"timestamp":1,"values":{"bot":40}},{"timestamp":2,"values":{"bot":30.333}}]},` +
				`"latency":{"chart_data":[` +
				`{"timestamp":1,"values":{"p50":120,"p95":900,"p99":1500}},` +
				`{"timestamp":2,"values":{"p50":80,"p95":400,"p99":2400}}]}}`,
			"METRIC         SERIES   LATEST     PEAK\n" +
				"CPU (cores)    bot      0.020      0.020\n" +
				"CPU (cores)    bot-db   0.500      0.500\n" +
				"Memory         bot      50.0 MiB   100.0 MiB\n" +
				"Requests       bot      30.3       40\n" +
				"Status codes   2xx      30.3       30.3\n" +
				"Status codes   5xx      0          1\n" +
				"Latency        p50      80ms       120ms\n" +
				"Latency        p95      400ms      900ms\n" +
				"Latency        p99      2.40s      2.40s\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a platform.AgentAnalytics
			if err := json.Unmarshal([]byte(tt.body), &a); err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			if err := PrintAgentAnalytics(&buf, a); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
