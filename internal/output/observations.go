package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var observationColumnMap = map[string]struct {
	Header string
	Value  func(o *clients.ObservationInfo) string
}{
	"id":       {"ID", func(o *clients.ObservationInfo) string { return o.ID }},
	"trace_id": {"TRACE ID", func(o *clients.ObservationInfo) string { return o.TraceID }},
	"type":     {"TYPE", func(o *clients.ObservationInfo) string { return o.Type }},
	"name":     {"NAME", func(o *clients.ObservationInfo) string { return o.Name }},
	"start_time": {
		"START TIME",
		func(o *clients.ObservationInfo) string { return LocalTime(o.StartTime) },
	},
	"end_time": {
		"END TIME",
		func(o *clients.ObservationInfo) string { return LocalTime(o.EndTime) },
	},
	"parent_observation_id": {
		"PARENT ID",
		func(o *clients.ObservationInfo) string { return o.ParentObservationID },
	},
	"level": {"LEVEL", func(o *clients.ObservationInfo) string { return o.Level }},
	"status_message": {
		"STATUS",
		func(o *clients.ObservationInfo) string { return o.StatusMessage },
	},
	"model": {"MODEL", func(o *clients.ObservationInfo) string { return o.Model }},
	"input_tokens": {
		"INPUT TOKENS",
		func(o *clients.ObservationInfo) string { return formatInt(o.InputTokens) },
	},
	"output_tokens": {
		"OUTPUT TOKENS",
		func(o *clients.ObservationInfo) string { return formatInt(o.OutputTokens) },
	},
	"total_tokens": {
		"TOTAL TOKENS",
		func(o *clients.ObservationInfo) string { return formatInt(o.TotalTokens) },
	},
	"total_cost": {
		"COST",
		func(o *clients.ObservationInfo) string { return formatCost(o.TotalCost) },
	},
	"latency_ms": {
		"LATENCY (ms)",
		func(o *clients.ObservationInfo) string { return formatFloat(o.LatencyMs, "ms") },
	},
}

var standaloneObservationColumnMap = map[string]struct {
	Header string
	Value  func(o *clients.StandaloneObservationInfo) string
}{
	"id": {"ID", func(o *clients.StandaloneObservationInfo) string { return o.ID }},
	"trace_id": {
		"TRACE ID",
		func(o *clients.StandaloneObservationInfo) string { return o.TraceID },
	},
	"type":  {"TYPE", func(o *clients.StandaloneObservationInfo) string { return o.Type }},
	"name":  {"NAME", func(o *clients.StandaloneObservationInfo) string { return o.Name }},
	"model": {"MODEL", func(o *clients.StandaloneObservationInfo) string { return o.Model }},
	"environment": {
		"ENVIRONMENT",
		func(o *clients.StandaloneObservationInfo) string { return o.Environment },
	},
	"user_id": {"USER ID", func(o *clients.StandaloneObservationInfo) string { return o.UserID }},
	"version": {"VERSION", func(o *clients.StandaloneObservationInfo) string { return o.Version }},
	"start_time": {
		"START TIME",
		func(o *clients.StandaloneObservationInfo) string { return LocalTime(o.StartTime) },
	},
	"end_time": {
		"END TIME",
		func(o *clients.StandaloneObservationInfo) string { return LocalTime(o.EndTime) },
	},
	"parent_observation_id": {
		"PARENT ID",
		func(o *clients.StandaloneObservationInfo) string { return o.ParentObservationID },
	},
	"level": {"LEVEL", func(o *clients.StandaloneObservationInfo) string { return o.Level }},
	"status_message": {
		"STATUS",
		func(o *clients.StandaloneObservationInfo) string { return o.StatusMessage },
	},
	"input_tokens": {
		"INPUT TOKENS",
		func(o *clients.StandaloneObservationInfo) string { return formatInt(o.InputTokens) },
	},
	"output_tokens": {
		"OUTPUT TOKENS",
		func(o *clients.StandaloneObservationInfo) string { return formatInt(o.OutputTokens) },
	},
	"total_tokens": {
		"TOTAL TOKENS",
		func(o *clients.StandaloneObservationInfo) string { return formatInt(o.TotalTokens) },
	},
	"total_cost": {
		"COST",
		func(o *clients.StandaloneObservationInfo) string { return formatCost(o.TotalCost) },
	},
	"latency_ms": {
		"LATENCY (ms)",
		func(o *clients.StandaloneObservationInfo) string { return formatFloat(o.LatencyMs, "ms") },
	},
}

func PrintObservationList(
	out io.Writer,
	observations []clients.ObservationInfo,
	columns []string,
) error {
	if len(observations) == 0 {
		fmt.Fprintln(out, "No observations found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := observationColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(observations))
	for i, o := range observations {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := observationColumnMap[col]; ok {
				row[j] = def.Value(&o)
			}
		}
		rows[i] = row
	}

	return PrintTable(out, headers, rows)
}

func PrintStandaloneObservationList(
	out io.Writer,
	observations []clients.StandaloneObservationInfo,
	meta clients.CursorMeta,
	columns []string,
) error {
	if len(observations) == 0 {
		fmt.Fprintln(out, "No observations found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := standaloneObservationColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(observations))
	for i, observation := range observations {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := standaloneObservationColumnMap[col]; ok {
				row[j] = def.Value(&observation)
			}
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	if meta.NextCursor != "" {
		fmt.Fprintf(out, "\nNext cursor: %s\n", meta.NextCursor)
	}

	return nil
}

func PrintObservationDetail(out io.Writer, obs *clients.ObservationDetail) error {
	isTTY := isTerminal(out)
	const jsonPrefix = "  "

	w := NewDescribeWriter(out)
	// Core
	fmt.Fprintf(w, "ID:\t%s\n", obs.ID)
	fmt.Fprintf(w, "Trace ID:\t%s\n", obs.TraceID)
	fmt.Fprintf(w, "Type:\t%s\n", obs.Type)
	fmt.Fprintf(w, "Name:\t%s\n", obs.Name)
	fmt.Fprintf(w, "Start Time:\t%s\n", LocalTime(obs.StartTime))
	fmt.Fprintf(w, "End Time:\t%s\n", LocalTime(obs.EndTime))
	if obs.ParentObservationID != "" {
		fmt.Fprintf(w, "Parent Observation ID:\t%s\n", obs.ParentObservationID)
	}
	if obs.Level != "" {
		fmt.Fprintf(w, "Level:\t%s\n", obs.Level)
	}
	if obs.StatusMessage != "" {
		fmt.Fprintf(w, "Status Message:\t%s\n", obs.StatusMessage)
	}

	// Model
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--- Model ---")
	fmt.Fprintf(w, "Model:\t%s\n", obs.Model)
	if len(obs.ModelParameters) > 0 && string(obs.ModelParameters) != "null" {
		fmt.Fprintln(w, "Parameters:")
		fmt.Fprintf(
			w,
			"%s\n",
			indentLines(prettyJSONUnwrapString(obs.ModelParameters, isTTY), jsonPrefix),
		)
	}

	// Metrics
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--- Metrics ---")
	fmt.Fprintf(w, "Latency:\t%s\n", formatFloat(obs.LatencyMs, "ms"))
	fmt.Fprintf(w, "Input Tokens:\t%s\n", formatInt(obs.InputTokens))
	fmt.Fprintf(w, "Output Tokens:\t%s\n", formatInt(obs.OutputTokens))
	fmt.Fprintf(w, "Total Tokens:\t%s\n", formatInt(obs.TotalTokens))
	fmt.Fprintf(w, "Total Cost:\t%s\n", formatCost(obs.TotalCost))

	// Prompt
	if obs.PromptName != "" || obs.PromptVersion != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "--- Prompt ---")
		if obs.PromptName != "" {
			fmt.Fprintf(w, "Prompt Name:\t%s\n", obs.PromptName)
		}
		if obs.PromptVersion != nil {
			fmt.Fprintf(w, "Prompt Version:\t%d\n", *obs.PromptVersion)
		}
	}

	// IO
	if len(obs.Input) > 0 && string(obs.Input) != "null" {
		fmt.Fprintf(
			w,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Input:", colorGreen),
			indentLines(prettyJSON(obs.Input, isTTY), jsonPrefix),
		)
	}

	if len(obs.Output) > 0 && string(obs.Output) != "null" {
		fmt.Fprintf(
			w,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Output:", colorRed),
			indentLines(prettyJSON(obs.Output, isTTY), jsonPrefix),
		)
	}

	if len(obs.Metadata) > 0 && string(obs.Metadata) != "null" {
		fmt.Fprintf(
			w,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Metadata:", colorOrange),
			indentLines(prettyJSON(obs.Metadata, isTTY), jsonPrefix),
		)
	}

	return w.Flush()
}
