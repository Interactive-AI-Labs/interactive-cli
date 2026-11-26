package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func formatFigure(f platform.Figure) string {
	if f.Money == nil {
		return "-"
	}
	amount, err := strconv.ParseFloat(f.Money.Amount, 64)
	if err != nil {
		return f.Money.Amount + " " + f.Money.Currency
	}
	return fmt.Sprintf("%.2f %s", amount, f.Money.Currency)
}

func PrintAgentOverview(out io.Writer, o platform.AgentOverview) error {
	fmt.Fprintf(out, "%s, %s to %s\n\n", Printable(o.DisplayName),
		LocalTime(o.Window.From), LocalTime(o.Window.To))

	p, c := o.Performance, o.Cost
	metric := func(name string, m platform.MetricDelta, format func(*float64) string) []string {
		return []string{name, format(m.Value), format(m.Previous), formatFloat(m.DeltaPct, "%")}
	}
	average := func(v *float64) string { return formatFloat(v, "") }
	cost := func(name string, m platform.CostMetric) []string {
		return []string{
			name,
			formatFigure(m.Value),
			formatFigure(m.Previous),
			formatFloat(m.DeltaPct, "%"),
		}
	}
	rows := [][]string{
		metric("Sessions", p.Sessions, formatOptionalFloat),
		metric("Session latency", p.SessionLatencyMs, formatLatencyMs),
		metric("Traces", p.Traces, formatOptionalFloat),
		metric("Trace latency", p.TraceLatencyMs, formatLatencyMs),
		metric("Traces per session", p.AvgTracesPerSession, average),
		metric("Iterations per trace", p.AvgIterationsPerTrace, average),
		metric("Observations", p.Observations, formatOptionalFloat),
		metric("Scores", p.Scores, formatOptionalFloat),
		cost("Total Cost", c.Agent),
		cost("Interactive Agent", c.InteractiveAgent),
		cost("LLM Cost", c.LLM),
	}
	if c.Tokens != nil {
		rows = append(rows, []string{"Tokens", strconv.Itoa(c.Tokens.Rated), "-", "-"})
	}
	if err := PrintTable(out, []string{"METRIC", "VALUE", "PREVIOUS", "CHANGE"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, u := range []struct {
		kind  string
		items []platform.NameCount
	}{
		{"Document", o.Usage.DocumentsReferenced},
		{"Policy", o.Usage.PoliciesMatched},
		{"Routine", o.Usage.RoutineTriggers},
		{"Tool", o.Usage.ToolCalls},
	} {
		for _, item := range u.items {
			rows = append(rows, []string{u.kind, Printable(item.Name), strconv.Itoa(item.Count)})
		}
	}
	if len(rows) > 0 {
		fmt.Fprintln(out)
		if err := PrintTable(out, []string{"USAGE", "NAME", "COUNT"}, rows); err != nil {
			return err
		}
	}

	rows = nil
	for _, e := range o.Errors.Types {
		rows = append(
			rows,
			[]string{
				"Observation",
				strconv.Itoa(e.ObservationCount),
				Printable(e.ObservationType + ": " + e.Message),
			},
		)
	}
	for _, e := range o.Errors.Traces {
		rows = append(rows, []string{"Trace", strconv.Itoa(e.ErrorCount), e.TraceID})
	}
	for _, e := range o.Errors.Sessions {
		rows = append(rows, []string{"Session", strconv.Itoa(e.ErrorCount), e.SessionID})
	}
	if len(rows) == 0 {
		return nil
	}
	fmt.Fprintln(out) // the message is last so a long one doesn't push the count off screen
	return PrintTable(out, []string{"ERRORS", "COUNT", "NAME"}, rows)
}

func PrintAgentPerformance(out io.Writer, p platform.AgentPerformance) error {
	costs := map[string]platform.Figure{}
	for _, c := range p.AgentCosts {
		costs[c.AgentID] = c.Total
	}
	rows := make([][]string, 0, len(p.Agents))
	for _, a := range p.Agents {
		rows = append(rows, []string{
			Printable(a.Name),
			formatInt(a.Traces),
			formatInt(a.Sessions),
			formatLatencyMs(a.TraceLatency),
			formatLatencyMs(a.SessionLatency),
			formatFigure(costs[a.AgentID]),
		})
	}
	if err := PrintTable(
		out,
		[]string{"NAME", "TRACES", "SESSIONS", "TRACE LATENCY", "SESSION LATENCY", "COST"},
		rows,
	); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "\nTotal Cost: %s\n", formatFigure(p.CostTotal))
	return err
}

func PrintAgentActivity(out io.Writer, a platform.AgentActivity) error {
	if len(a.Activity) == 0 {
		_, err := fmt.Fprintln(out, "No activity found.")
		return err
	}
	rows := make([][]string, 0, len(a.Activity))
	for _, e := range a.Activity {
		data, _ := json.Marshal(e.Data)
		rows = append(rows, []string{
			LocalTime(e.Time),
			Printable(e.Action),
			Printable(e.Actor.Display),
			strconv.FormatBool(e.Success),
			Printable(truncateString(string(data), 80)),
		})
	}
	if err := PrintTable(
		out,
		[]string{"TIME", "ACTION", "ACTOR", "SUCCESS", "DATA"},
		rows,
	); err != nil {
		return err
	}
	PrintPageMeta(out, a.Meta.Page, a.Meta.TotalPages, a.Meta.TotalItems)
	return nil
}
