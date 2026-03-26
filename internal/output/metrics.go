package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var metricsDailyColumnMap = map[string]struct {
	Header string
	Value  func(m *clients.DailyMetric) string
}{
	"date": {"DATE", func(m *clients.DailyMetric) string { return m.Date }},
	"count_traces": {
		"TRACE COUNT",
		func(m *clients.DailyMetric) string { return formatInt(m.CountTraces) },
	},
	"count_observations": {
		"OBSERVATION COUNT",
		func(m *clients.DailyMetric) string { return formatInt(m.CountObservations) },
	},
	"total_cost": {
		"TOTAL COST",
		func(m *clients.DailyMetric) string { return formatCost(m.TotalCost) },
	},
	"total_tokens": {
		"TOTAL TOKENS",
		func(m *clients.DailyMetric) string { return formatInt(m.TotalTokens) },
	},
}

func PrintMetricsDaily(
	out io.Writer,
	metrics []clients.DailyMetric,
	meta clients.PageMeta,
	columns []string,
	showModels bool,
) error {
	if len(metrics) == 0 {
		fmt.Fprintln(out, "No daily metrics found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = metricsDailyColumnMap[col].Header
	}

	rows := make([][]string, len(metrics))
	for i, metric := range metrics {
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = metricsDailyColumnMap[col].Value(&metric)
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	if showModels {
		for _, metric := range metrics {
			if len(metric.Models) == 0 {
				continue
			}

			fmt.Fprintf(out, "\nModel Breakdown for %s\n", metric.Date)
			modelHeaders := []string{"MODEL", "OBSERVATIONS", "TOTAL TOKENS", "TOTAL COST"}
			modelRows := make([][]string, len(metric.Models))
			for i, model := range metric.Models {
				modelRows[i] = []string{
					model.Model,
					formatInt(model.CountObservations),
					formatInt(model.TotalTokens),
					formatCost(model.TotalCost),
				}
			}
			if err := PrintTable(out, modelHeaders, modelRows); err != nil {
				return err
			}
		}
	}

	fmt.Fprintf(
		out,
		"\nPage %d of %d (%d total items)\n",
		meta.Page,
		meta.TotalPages,
		meta.TotalItems,
	)
	return nil
}
