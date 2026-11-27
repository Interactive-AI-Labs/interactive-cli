package output

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"strconv"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func formatFigure(f platform.Figure) string {
	if f.Money == nil {
		return "-"
	}
	return f.Money.Amount + " " + f.Money.Currency // the platform rounds it for display
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

// PrintAgentAnalytics prints each series' latest and peak value, not every time point, then
// each workload's CPU and memory quota.
func PrintAgentAnalytics(out io.Writer, a platform.AgentAnalytics) error {
	count := func(v float64) string { return strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64) }
	cores := func(v float64) string { return fmt.Sprintf("%.3f", v) }
	memory := func(v float64) string { return HumanBytes(int64(v)) }
	metrics := []struct {
		name   string
		chart  platform.AgentChart
		format func(float64) string
	}{
		{"CPU (cores)", a.CPUUsage.AgentChart, cores},
		{"Memory", a.MemoryUsage.AgentChart, memory},
		{"Requests", a.RequestsPerDomain, count},
		{"Status codes", a.StatusCodes, count},
		{"Latency", a.Latency, func(v float64) string { return formatLatencyMs(&v) }},
	}
	var rows [][]string
	for _, m := range metrics {
		latest, peak := map[string]float64{}, map[string]float64{}
		for _, point := range m.chart.ChartData {
			for series, v := range point.Values {
				if p, ok := peak[series]; !ok || v > p {
					peak[series] = v
				}
				latest[series] = v
			}
		}
		for _, series := range slices.Sorted(maps.Keys(latest)) {
			rows = append(rows, []string{
				m.name, Printable(series), m.format(latest[series]), m.format(peak[series]),
			})
		}
	}

	optional := func(v *float64, format func(float64) string) string {
		if v == nil {
			return "-"
		}
		return format(*v)
	}
	replicas := func(v float64) string { return strconv.FormatFloat(math.Round(v), 'f', -1, 64) }
	var quota [][]string
	for _, q := range a.CPUUsage.Quota {
		quota = append(quota, []string{
			"CPU (cores)", Printable(q.Workload), optional(q.Replicas, replicas),
			optional(q.CPUAllocated, cores), cores(q.CPUUsage),
			formatFloat(q.CPUUsagePercent, "%"), formatFloat(q.CPUMaxUsagePercent, "%"),
		})
	}
	for _, q := range a.MemoryUsage.Quota {
		quota = append(quota, []string{
			"Memory", Printable(q.Workload), optional(q.Replicas, replicas),
			optional(q.MemoryAllocatedBytes, memory), memory(q.MemoryUsageBytes),
			formatFloat(q.MemoryUsagePercent, "%"), formatFloat(q.MemoryMaxUsagePercent, "%"),
		})
	}

	if len(rows)+len(quota) == 0 {
		_, err := fmt.Fprintln(out, "No analytics found.")
		return err
	}
	if len(rows) > 0 {
		if err := PrintTable(
			out,
			[]string{"METRIC", "SERIES", "LATEST", "PEAK"},
			rows,
		); err != nil {
			return err
		}
		if len(quota) == 0 {
			return nil
		}
		fmt.Fprintln(out)
	}
	return PrintTable(out, []string{
		"QUOTA", "WORKLOAD", "REPLICAS", "ALLOCATED", "USAGE", "USAGE %", "MAX USAGE %",
	}, quota)
}
