package output

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func PrintRouterModelList(
	out io.Writer,
	models []platform.RouterModel,
	meta platform.PageMeta,
) error {
	if len(models) == 0 {
		fmt.Fprintln(out, "No models found.")
		return nil
	}

	headers := []string{"NAME", "DEVELOPER", "PROVIDERS", "CONTEXT", "REGION", "RECOMMENDED", "ID"}
	rows := make([][]string, len(models))
	for i, m := range models {
		providers := "-"
		if len(m.ProviderNames) > 0 {
			providers = strconv.Itoa(len(m.ProviderNames))
		}
		rows[i] = []string{
			m.ModelName,
			cmp.Or(m.Developer, m.EndpointProvider, "-"),
			providers,
			formatInt(m.ContextLength),
			routerModelRegionName(m.Region),
			formatRecommendations(m.Recommendations),
			m.ID,
		}
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	PrintPageMeta(out, meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

func PrintRouterModelDetail(out io.Writer, m *platform.RouterModelDetail) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", m.ID)
	fmt.Fprintf(w, "Model Name:\t%s\n", m.ModelName)
	if m.MarketingName != "" {
		fmt.Fprintf(w, "Marketing:\t%s\n", m.MarketingName)
	}
	fmt.Fprintf(w, "Developer:\t%s\n", cmp.Or(m.Developer, m.EndpointProvider, "-"))
	fmt.Fprintf(w, "Region:\t%s\n", routerModelRegionLine(m.Region, m.HasOtherRegion))
	if len(m.Recommendations) > 0 {
		fmt.Fprintf(w, "Recommended:\t%s\n", formatRecommendations(m.Recommendations))
	}
	if m.Description != "" {
		fmt.Fprintf(w, "Description:\t%s\n", m.Description)
	}
	if m.ContextLength != nil {
		fmt.Fprintf(w, "Context:\t%s\n", formatInt(m.ContextLength))
	}
	fmt.Fprintf(w, "Latency:\t%s\n", routerModelLatency(m.Latency))
	if len(m.Capabilities) > 0 {
		fmt.Fprintf(w, "Capabilities:\t%s\n", strings.Join(m.Capabilities, ", "))
	}
	if m.TokenizerID != "" {
		fmt.Fprintf(w, "Tokenizer:\t%s\n", m.TokenizerID)
	}
	if m.CreatedAt != "" {
		fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(m.CreatedAt))
	}
	if len(m.Prices) > 0 {
		fmt.Fprintln(w, "Prices:")
		for _, k := range slices.Sorted(maps.Keys(m.Prices)) {
			fmt.Fprintf(w, "  %s:\t%s\n", k, strconv.FormatFloat(m.Prices[k], 'f', -1, 64))
		}
	}
	if len(m.Providers) == 0 {
		fmt.Fprintln(w, "Providers:\tnone")
		return w.Flush()
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(out)

	headers := []string{
		"PROVIDER",
		"TAG",
		"INPUT",
		"OUTPUT",
		"CONTEXT",
		"MAX OUTPUT",
		"UPTIME (1D)",
	}
	rows := make([][]string, len(m.Providers))
	for i, p := range m.Providers {
		rows[i] = []string{
			p.Name,
			cmp.Or(derefString(p.Tag), "-"),
			routerModelPrice(p.Prices, "input"),
			routerModelPrice(p.Prices, "output"),
			formatInt(p.ContextLength),
			formatInt(p.MaxOutputTokens),
			formatFloat(p.UptimeLastDay, "%"),
		}
	}
	return PrintTable(out, headers, rows)
}

func routerModelRegionName(region string) string {
	switch region {
	case platform.RouterModelRegionUS:
		return "Global"
	case platform.RouterModelRegionEU:
		return "EU"
	}
	return region
}

func routerModelRegionLine(region string, hasOtherRegion bool) string {
	name := routerModelRegionName(region)
	if !hasOtherRegion {
		return name
	}
	var other string
	switch region {
	case platform.RouterModelRegionUS:
		other = platform.RouterModelRegionEU
	case platform.RouterModelRegionEU:
		other = platform.RouterModelRegionUS
	default:
		return name
	}
	return fmt.Sprintf("%s (also in %s)", name, routerModelRegionName(other))
}

func routerModelPrice(prices map[string]float64, key string) string {
	v, ok := prices[key]
	if !ok {
		return "-"
	}
	return formatOptionalFloat(&v)
}

func routerModelLatency(l *platform.RouterModelLatency) string {
	if l == nil || (l.TimeToFirstTokenSeconds == nil && l.SecondsPerOutputToken == nil) {
		return "-"
	}
	seconds := func(v *float64) string {
		if v == nil {
			return "-"
		}
		return strconv.FormatFloat(*v, 'g', 3, 64) + " s"
	}
	return fmt.Sprintf(
		"%s to first token, %s per output token (%s, UTC)",
		seconds(l.TimeToFirstTokenSeconds),
		seconds(l.SecondsPerOutputToken),
		l.Date,
	)
}

// Sorted by category so table output is stable across calls.
func formatRecommendations(recs map[string]int) string {
	if len(recs) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(recs))
	for _, k := range slices.Sorted(maps.Keys(recs)) {
		parts = append(parts, k+" #"+strconv.Itoa(recs[k]))
	}
	return strings.Join(parts, ", ")
}
