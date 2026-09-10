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
		w := NewDescribeWriter(out)
		fmt.Fprintf(w, "  Error:\t%s\n", cmp.Or(run.Error, "the run produced no scenarios"))
		if err := w.Flush(); err != nil {
			return err
		}
		fmt.Fprintln(out)
		printReplayVerdict(out, run)
		return nil
	}

	single := len(run.Batches) == 1
	if !single {
		rows := make([][]string, 0, len(run.Batches))
		for _, b := range run.Batches {
			rows = append(rows, []string{
				verdictWord(b.Status), b.Scenario, fmt.Sprintf("%d/%d", b.Passed, b.Repeat),
			})
		}
		if err := PrintTable(out, nil, rows); err != nil {
			return err
		}
	}

	for _, b := range run.Batches {
		if !single && b.Status == deployment.ReplayStatusPassed {
			continue
		}
		// The summary already left a blank line; the table needs one per group.
		if !single {
			fmt.Fprintln(out)
		}
		fmt.Fprintln(out, b.Scenario)
		if err := printBatchExpanded(out, b); err != nil {
			return err
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
	fmt.Fprintln(w, joinHeader(parts...))
}

// printBatchExpanded gives each iteration its own describe writer, so one
// iteration's long label cannot pad another's.
func printBatchExpanded(out io.Writer, b deployment.ReplayBatch) error {
	if b.Error != "" {
		w := NewDescribeWriter(out)
		fmt.Fprintf(w, "  Error:\t%s\n", b.Error)
		return w.Flush()
	}
	for i, it := range b.Iterations {
		fmt.Fprintf(out, "--- run %d/%d  %s ---\n", i+1, len(b.Iterations), verdictWord(it.Status))
		w := NewDescribeWriter(out)
		printIteration(w, it)
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func printIteration(w io.Writer, it deployment.ReplayIteration) {
	if it.Status == deployment.ReplayStatusError {
		fmt.Fprintf(w, "  Error:\t%s\n", it.Error)
		return
	}
	line := func(label, value string) {
		if value != "" {
			fmt.Fprintf(w, "  %s:\t%s\n", label, value)
		}
	}
	line("Turns", fmt.Sprintf("%d", it.Turns))
	line("Tools Called", strings.Join(it.Observed.ToolsCalled, ", "))
	line("Tools Denied", strings.Join(it.Observed.ToolsDenied, ", "))
	line("Steps", strings.Join(it.Observed.Steps, ", "))
	line("Routines", strings.Join(it.Observed.Routines, ", "))
	line("Policies", strings.Join(it.Observed.Policies, ", "))
	if it.Judge != nil {
		line("Judge", it.Judge.Score)
		// Continuation rows: an empty label cell keeps every line of the
		// reasoning in the value column, and the block's alignment.
		for _, l := range strings.Split(it.Judge.Reasoning, "\n") {
			if l != "" {
				fmt.Fprintf(w, "  \t%s\n", l)
			}
		}
	}
	line("Session", it.SessionKey)
	line("Eval Trace", it.EvalTraceID)
	if len(it.TraceIDs) > 0 {
		line("Turn-1 Trace", it.TraceIDs[0])
	}
	if len(it.Diverged) > 0 {
		line("Diverged", "recorded but not replayed: "+strings.Join(it.Diverged, ", "))
	}
	for _, f := range it.Failures {
		line("Failure", f)
	}
}

func printReplayVerdict(w io.Writer, run *deployment.ReplayRun) {
	word := verdictShort(run.Status)
	// A run that errored before producing any scenario has nothing to count.
	if len(run.Batches) == 0 {
		fmt.Fprintf(w, "%-6s run %s\n", word, run.RunID)
		return
	}
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
