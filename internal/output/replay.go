package output

import (
	"cmp"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

const maxSkippedShown = 5

// PrintReplayTarget names the agent a replay is about to hit, before anything
// is sent, so a wrong target is visible.
func PrintReplayTarget(w io.Writer, name, version string, revision int) {
	fmt.Fprintf(w, "%s  %s  rev %d\n", name, version, revision)
}

// PrintReplaySkipped lists dataset items the agent will not replay. The agent
// reports the whole dataset even for a subset, so with requested names only
// those are shown; a long list collapses to a count.
func PrintReplaySkipped(w io.Writer, skipped []deployment.ReplaySkipped, requested []string) {
	skipped = filterSkipped(skipped, requested)
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

func filterSkipped(
	skipped []deployment.ReplaySkipped,
	requested []string,
) []deployment.ReplaySkipped {
	if len(requested) == 0 {
		return skipped
	}
	want := make(map[string]bool, len(requested))
	for _, r := range requested {
		want[r] = true
	}
	var kept []deployment.ReplaySkipped
	for _, s := range skipped {
		if want[s.ID] {
			kept = append(kept, s)
		}
	}
	return kept
}

func PrintReplayProgress(w io.Writer, finished, total int) {
	fmt.Fprintf(w, "%d/%d scenarios finished\n", finished, total)
}

// PrintReplayRun renders a finished run. A scenario is expanded iteration by
// iteration when it is the only one or when it failed or errored; passed
// scenarios in a multi-scenario run are one row each. A run that errored
// before producing any scenario prints its own error instead.
func PrintReplayRun(out io.Writer, run *deployment.ReplayRun, requested []string) error {
	printReplaySummary(out, run, requested)
	fmt.Fprintln(out)

	if len(run.Batches) == 0 {
		fmt.Fprintf(
			out,
			"  ERROR          %s\n",
			cmp.Or(run.Error, "the run produced no scenarios"),
		)
		fmt.Fprintln(out)
		fmt.Fprintf(out, "%-6s run %s\n", verdictShort(run.Status), run.RunID)
		return nil
	}

	single := len(run.Batches) == 1
	width := nameWidth(run.Batches)
	for _, b := range run.Batches {
		if single {
			fmt.Fprintln(out, b.Scenario)
		} else {
			fmt.Fprintf(
				out,
				"%-7s %-*s %d/%d\n",
				verdictWord(b.Status),
				width,
				b.Scenario,
				b.Passed,
				b.Repeat,
			)
		}
		if single || b.Status != deployment.ReplayStatusPassed {
			printBatchExpanded(out, b)
		}
	}

	fmt.Fprintln(out)
	printReplayVerdict(out, run)
	return nil
}

// PrintReplayPointer tells the reader where the verdict lives on the platform:
// the scores query and the eval trace worth opening first.
func PrintReplayPointer(w io.Writer, run *deployment.ReplayRun) {
	scenario, trace := pointerTrace(run)
	if trace == "" {
		return
	}
	fmt.Fprintf(
		w,
		"scores: iai scores list --name replay.verdict --columns name,trace_id,comment"+
			" · eval trace (%s): iai traces get %s\n",
		scenario,
		trace,
	)
}

func printReplaySummary(w io.Writer, run *deployment.ReplayRun, requested []string) {
	var parts []string
	if run.Dataset != "" {
		parts = append(parts, "dataset "+run.Dataset)
		if n := len(run.Batches); n > 0 {
			parts = append(parts, fmt.Sprintf("%d scenario%s", n, plural(n)))
		}
	} else if run.Scenario != "" {
		parts = append(parts, "scenario "+run.Scenario)
	}
	parts = append(parts, fmt.Sprintf("repeat %d", run.Repeat))
	if run.Concurrency > 0 {
		parts = append(parts, fmt.Sprintf("concurrency %d", run.Concurrency))
	}
	if n := len(filterSkipped(run.Skipped, requested)); n > 0 {
		parts = append(parts, fmt.Sprintf("skipped %d", n))
	}
	fmt.Fprintln(w, strings.Join(parts, "   "))
}

func nameWidth(batches []deployment.ReplayBatch) int {
	width := 0
	for _, b := range batches {
		if len(b.Scenario) > width {
			width = len(b.Scenario)
		}
	}
	return width
}

func printBatchExpanded(w io.Writer, b deployment.ReplayBatch) {
	if b.Error != "" {
		fmt.Fprintf(w, "  ERROR          %s\n", b.Error)
		return
	}
	for i, it := range b.Iterations {
		fmt.Fprintf(w, "--- run %d/%d  %s ---\n", i+1, len(b.Iterations), verdictWord(it.Status))
		printIteration(w, it)
	}
}

func printIteration(w io.Writer, it deployment.ReplayIteration) {
	if it.Status == deployment.ReplayStatusError {
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

func printReplayVerdict(w io.Writer, run *deployment.ReplayRun) {
	word := verdictShort(run.Status)
	if len(run.Batches) == 1 {
		b := run.Batches[0]
		fmt.Fprintf(w, "%-6s %d/%d passed     run %s\n", word, b.Passed, b.Repeat, run.RunID)
		return
	}
	passed := 0
	for _, b := range run.Batches {
		if b.Status == deployment.ReplayStatusPassed {
			passed++
		}
	}
	fmt.Fprintf(
		w,
		"%-6s %d%% of %d scenarios passed     run %s\n",
		word,
		passed*100/len(run.Batches),
		len(run.Batches),
		run.RunID,
	)
}

// pointerTrace is the eval trace worth opening first, with its scenario: the
// first failing or errored iteration's, else the first iteration's.
func pointerTrace(run *deployment.ReplayRun) (scenario, trace string) {
	for _, b := range run.Batches {
		for _, it := range b.Iterations {
			if it.EvalTraceID == "" {
				continue
			}
			if trace == "" {
				scenario, trace = b.Scenario, it.EvalTraceID
			}
			if it.Status != deployment.ReplayStatusPassed {
				return b.Scenario, it.EvalTraceID
			}
		}
	}
	return scenario, trace
}

func verdictWord(status string) string {
	switch status {
	case deployment.ReplayStatusPassed:
		return "PASSED"
	case deployment.ReplayStatusFailed:
		return "FAILED"
	case deployment.ReplayStatusError:
		return "ERROR"
	}
	return strings.ToUpper(status)
}

func verdictShort(status string) string {
	switch status {
	case deployment.ReplayStatusPassed:
		return "PASS"
	case deployment.ReplayStatusFailed:
		return "FAIL"
	case deployment.ReplayStatusError:
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
