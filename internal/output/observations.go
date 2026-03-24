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
		headers[i] = observationColumnMap[col].Header
	}

	rows := make([][]string, len(observations))
	for i, o := range observations {
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = observationColumnMap[col].Value(&o)
		}
		rows[i] = row
	}

	return PrintTable(out, headers, rows)
}

func PrintObservationDetail(out io.Writer, obs *clients.ObservationDetail) error {
	// Core
	fmt.Fprintf(out, "ID:                    %s\n", obs.ID)
	fmt.Fprintf(out, "Trace ID:              %s\n", obs.TraceID)
	fmt.Fprintf(out, "Type:                  %s\n", obs.Type)
	fmt.Fprintf(out, "Name:                  %s\n", obs.Name)
	fmt.Fprintf(out, "Start Time:            %s\n", LocalTime(obs.StartTime))
	fmt.Fprintf(out, "End Time:              %s\n", LocalTime(obs.EndTime))
	if obs.ParentObservationID != "" {
		fmt.Fprintf(out, "Parent Observation ID: %s\n", obs.ParentObservationID)
	}
	if obs.Level != "" {
		fmt.Fprintf(out, "Level:                 %s\n", obs.Level)
	}
	if obs.StatusMessage != "" {
		fmt.Fprintf(out, "Status Message:        %s\n", obs.StatusMessage)
	}

	// Model
	fmt.Fprintf(out, "\n--- Model ---\n")
	fmt.Fprintf(out, "Model: %s\n", obs.Model)
	if len(obs.ModelParameters) > 0 && string(obs.ModelParameters) != "null" {
		isTTY := isTerminal(out)
		fmt.Fprintf(
			out,
			"Parameters:\n%s\n",
			indentLines(prettyJSONUnwrapString(obs.ModelParameters, isTTY), "  "),
		)
	}

	// Metrics
	fmt.Fprintf(out, "\n--- Metrics ---\n")
	fmt.Fprintf(out, "Latency:       %s\n", formatFloat(obs.LatencyMs, "ms"))
	fmt.Fprintf(out, "Input Tokens:  %s\n", formatInt(obs.InputTokens))
	fmt.Fprintf(out, "Output Tokens: %s\n", formatInt(obs.OutputTokens))
	fmt.Fprintf(out, "Total Tokens:  %s\n", formatInt(obs.TotalTokens))
	fmt.Fprintf(out, "Total Cost:    %s\n", formatCost(obs.TotalCost))

	// Prompt
	if obs.PromptName != "" || obs.PromptVersion != nil {
		fmt.Fprintf(out, "\n--- Prompt ---\n")
		if obs.PromptName != "" {
			fmt.Fprintf(out, "Prompt Name:    %s\n", obs.PromptName)
		}
		if obs.PromptVersion != nil {
			fmt.Fprintf(out, "Prompt Version: %d\n", *obs.PromptVersion)
		}
	}

	// IO
	isTTY := isTerminal(out)
	const jsonPrefix = "  "

	if len(obs.Input) > 0 && string(obs.Input) != "null" {
		fmt.Fprintf(
			out,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Input:", colorGreen),
			indentLines(prettyJSON(obs.Input, isTTY), jsonPrefix),
		)
	}

	if len(obs.Output) > 0 && string(obs.Output) != "null" {
		fmt.Fprintf(
			out,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Output:", colorRed),
			indentLines(prettyJSON(obs.Output, isTTY), jsonPrefix),
		)
	}

	if len(obs.Metadata) > 0 && string(obs.Metadata) != "null" {
		fmt.Fprintf(
			out,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Metadata:", colorOrange),
			indentLines(prettyJSON(obs.Metadata, isTTY), jsonPrefix),
		)
	}

	return nil
}
