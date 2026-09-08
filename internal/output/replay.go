package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/agent"
)

const maxSkippedShown = 5

// PrintReplayTarget names the agent a replay is about to hit, before anything
// is sent, so a wrong target is visible.
func PrintReplayTarget(w io.Writer, name, version string, revision int, baseURL string) {
	fmt.Fprintf(w, "%s  %s  rev %d    %s\n", name, version, revision, baseURL)
}

// PrintReplaySkipped lists dataset items the agent will not replay. The agent
// reports the whole dataset even for a subset, so with requested names only
// those are shown; a long list collapses to a count.
func PrintReplaySkipped(w io.Writer, skipped []agent.Skipped, requested []string) {
	if len(requested) > 0 {
		want := make(map[string]bool, len(requested))
		for _, r := range requested {
			want[r] = true
		}
		var kept []agent.Skipped
		for _, s := range skipped {
			if want[s.ID] {
				kept = append(kept, s)
			}
		}
		skipped = kept
	}
	if len(skipped) == 0 {
		return
	}
	if len(skipped) > maxSkippedShown {
		fmt.Fprintf(w, "skipped %d dataset items (see --json)\n", len(skipped))
		return
	}
	for _, s := range skipped {
		fmt.Fprintf(w, "skipped %s: %s\n", s.ID, s.Reason)
	}
}

func PrintReplayProgress(w io.Writer, finished, total int) {
	fmt.Fprintf(w, "%d/%d scenarios finished\n", finished, total)
}

// PrintReplayRun renders a finished run. A scenario is expanded iteration by
// iteration when it is the only one or when it failed or errored; passed
// scenarios in a multi-scenario run are one row each.
func PrintReplayRun(out, errOut io.Writer, run *agent.Run) error {
	printReplaySummary(out, run)
	fmt.Fprintln(out)

	single := len(run.Batches) == 1
	for _, b := range run.Batches {
		if single || b.Status != agent.StatusPassed {
			printBatchExpanded(out, b, single)
		} else {
			fmt.Fprintf(
				out,
				"%-7s %-32s %d/%d\n",
				verdictWord(b.Status),
				b.Scenario,
				b.Passed,
				b.Repeat,
			)
		}
	}

	fmt.Fprintln(out)
	printReplayVerdict(out, run)

	if trace := pointerTrace(run); trace != "" {
		fmt.Fprintf(
			errOut,
			"scores: iai scores list --name replay.verdict --columns name,value,trace_id,comment"+
				" · eval trace: iai traces get %s\n",
			trace,
		)
	}
	return nil
}

func printReplaySummary(w io.Writer, run *agent.Run) {
	parts := []string{}
	if run.Dataset != "" {
		parts = append(parts, "dataset "+run.Dataset)
		parts = append(
			parts,
			fmt.Sprintf("%d scenario%s", len(run.Batches), plural(len(run.Batches))),
		)
	} else if run.Scenario != "" {
		parts = append(parts, "scenario "+run.Scenario)
	}
	parts = append(parts, fmt.Sprintf("repeat %d", run.Repeat))
	if run.Concurrency > 0 {
		parts = append(parts, fmt.Sprintf("concurrency %d", run.Concurrency))
	}
	if n := len(run.Skipped); n > 0 {
		parts = append(parts, fmt.Sprintf("skipped %d", n))
	}
	fmt.Fprintln(w, strings.Join(parts, "   "))
}

func printBatchExpanded(w io.Writer, b agent.Batch, single bool) {
	if single {
		fmt.Fprintln(w, b.Scenario)
	} else {
		fmt.Fprintf(w, "%-7s %-32s %d/%d\n", verdictWord(b.Status), b.Scenario, b.Passed, b.Repeat)
	}
	if b.Error != "" {
		fmt.Fprintf(w, "  ERROR          %s\n", b.Error)
		return
	}
	for i, it := range b.Iterations {
		fmt.Fprintf(w, "--- run %d/%d  %s ---\n", i+1, len(b.Iterations), verdictWord(it.Status))
		printIteration(w, it)
	}
}

func printIteration(w io.Writer, it agent.Iteration) {
	if it.Status == agent.StatusError {
		fmt.Fprintf(w, "  ERROR          %s\n", it.Error)
		return
	}
	line := func(label, value string) {
		if value != "" {
			fmt.Fprintf(w, "  %-14s %s\n", label, value)
		}
	}
	line("turns", fmt.Sprintf("%d", it.Turns))
	line("tools called", strings.Join(it.Observed.ToolsCalled, ", "))
	line("tools denied", strings.Join(it.Observed.ToolsDenied, ", "))
	line("steps", strings.Join(it.Observed.Steps, ", "))
	line("routines", strings.Join(it.Observed.Routines, ", "))
	line("policies", strings.Join(it.Observed.Policies, ", "))
	if it.Judge != nil {
		line("judge", it.Judge.Score)
		if it.Judge.Reasoning != "" {
			fmt.Fprintf(w, "    %s\n", it.Judge.Reasoning)
		}
	}
	line("session", it.SessionKey)
	if it.EvalTraceID != "" {
		trace := it.EvalTraceID
		if len(it.TraceIDs) > 0 {
			trace += "        turn-1 trace " + it.TraceIDs[0]
		}
		line("eval trace", trace)
	}
	if len(it.Diverged) > 0 {
		line("DIVERGED", "recorded but not replayed: "+strings.Join(it.Diverged, ", "))
	}
	for _, f := range it.Failures {
		line("FAIL", f)
	}
}

func printReplayVerdict(w io.Writer, run *agent.Run) {
	word := verdictShort(run.Status)
	if len(run.Batches) == 1 {
		b := run.Batches[0]
		fmt.Fprintf(w, "%-6s %d/%d passed     run %s\n", word, b.Passed, b.Repeat, run.RunID)
		return
	}
	passed := 0
	for _, b := range run.Batches {
		if b.Status == agent.StatusPassed {
			passed++
		}
	}
	pct := 0
	if len(run.Batches) > 0 {
		pct = passed * 100 / len(run.Batches)
	}
	fmt.Fprintf(
		w,
		"%-6s %d%% of %d scenarios passed     run %s\n",
		word,
		pct,
		len(run.Batches),
		run.RunID,
	)
}

// pointerTrace is the eval trace worth opening first: the first failing or
// errored iteration's, else the first iteration's.
func pointerTrace(run *agent.Run) string {
	first := ""
	for _, b := range run.Batches {
		for _, it := range b.Iterations {
			if it.EvalTraceID == "" {
				continue
			}
			if first == "" {
				first = it.EvalTraceID
			}
			if it.Status != agent.StatusPassed {
				return it.EvalTraceID
			}
		}
	}
	return first
}

func verdictWord(status string) string {
	switch status {
	case agent.StatusPassed:
		return "PASSED"
	case agent.StatusFailed:
		return "FAILED"
	case agent.StatusError:
		return "ERROR"
	}
	return strings.ToUpper(status)
}

func verdictShort(status string) string {
	switch status {
	case agent.StatusPassed:
		return "PASS"
	case agent.StatusFailed:
		return "FAIL"
	case agent.StatusError:
		return "ERROR"
	}
	return strings.ToUpper(status)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
