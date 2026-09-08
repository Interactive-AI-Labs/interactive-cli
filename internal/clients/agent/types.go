package agent

// StartRequest is the body of POST /replays on the agent.
type StartRequest struct {
	Dataset      string         `json:"dataset,omitempty"`
	Scenarios    []string       `json:"scenarios,omitempty"`
	ScenarioBody map[string]any `json:"scenario_body,omitempty"`
	Repeat       int            `json:"repeat"`
	Concurrency  int            `json:"concurrency"`
}

// Skipped is a dataset item the agent did not replay, with the reason.
type Skipped struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// StartResponse is the 202 body of POST /replays.
type StartResponse struct {
	RunID   string    `json:"run_id"`
	Skipped []Skipped `json:"skipped"`
}

// Run is GET /replays/{run_id}. The agent returns a suite (dataset runs, with
// Batches) or a batch (inline runs, with Iterations); GetReplay normalises the
// batch form into a suite holding one batch so callers render one shape.
type Run struct {
	RunID       string      `json:"run_id"`
	Dataset     string      `json:"dataset"`
	Scenario    string      `json:"scenario"`
	Status      string      `json:"status"`
	Repeat      int         `json:"repeat"`
	Concurrency int         `json:"concurrency"`
	Passed      int         `json:"passed"`
	Skipped     []Skipped   `json:"skipped"`
	Batches     []Batch     `json:"batches"`
	Iterations  []Iteration `json:"iterations"`
	Error       string      `json:"error"`
	// Raw is the response body exactly as the agent returned it, for --json.
	Raw []byte `json:"-"`
}

// Batch is one scenario replayed Repeat times.
type Batch struct {
	RunID      string      `json:"run_id"`
	Scenario   string      `json:"scenario"`
	Status     string      `json:"status"`
	Repeat     int         `json:"repeat"`
	Passed     int         `json:"passed"`
	Iterations []Iteration `json:"iterations"`
	Error      string      `json:"error"`
}

// Iteration is one replayed conversation and its verdict.
type Iteration struct {
	RunID       string   `json:"run_id"`
	Scenario    string   `json:"scenario"`
	Status      string   `json:"status"`
	SessionKey  string   `json:"session_key"`
	TraceIDs    []string `json:"trace_ids"`
	EvalTraceID string   `json:"eval_trace_id"`
	Turns       int      `json:"turns"`
	Observed    Observed `json:"observed"`
	Diverged    []string `json:"diverged"`
	Judge       *Judge   `json:"judge"`
	Failures    []string `json:"failures"`
	Error       string   `json:"error"`
}

// Observed is what the agent actually did during an iteration.
type Observed struct {
	ToolsCalled []string `json:"tools_called"`
	ToolsDenied []string `json:"tools_denied"`
	Steps       []string `json:"steps"`
	Routines    []string `json:"routines"`
	Policies    []string `json:"policies"`
}

// Judge is the LLM judge's verdict when the scenario declared a rubric.
type Judge struct {
	Score     string `json:"score"`
	Reasoning string `json:"reasoning"`
}

const (
	StatusRunning = "running"
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusError   = "error"
)
