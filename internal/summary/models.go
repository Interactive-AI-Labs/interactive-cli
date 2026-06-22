// Package summary turns raw trace/observation data into compact, LLM-readable
// summary models. It performs no I/O and no rendering.
package summary

// Condition is a guideline whose condition matched ("marked true") in a turn.
type Condition struct {
	Text  string
	Score int
}

// ToolCall is a single tool invocation within an iteration.
type ToolCall struct {
	Name    string
	Args    string
	Result  string
	Errored bool
	ErrMsg  string
}

// Iteration is one engine preparation iteration within a turn.
type Iteration struct {
	Number     int
	Conditions []Condition
	Tools      []ToolCall
	KBQueries  []string
}

// TraceSummaryModel is the compact view of a single turn (one trace).
type TraceSummaryModel struct {
	Name       string
	Timestamp  string
	LatencyMs  *float64
	Cost       *float64
	Level      string
	Input      string
	Iterations []Iteration
	Reply      string
	Errors     []string
}

// Turn is one back-and-forth within a session.
type Turn struct {
	Number   int
	Customer string
	Agent    string
	Tools    []string
	Journeys []string
}

// SessionSummaryModel is the overall view of a conversation (one session).
type SessionSummaryModel struct {
	ID        string
	Agent     string
	TurnCount int
	Duration  string
	Cost      *float64
	Turns     []Turn
}
