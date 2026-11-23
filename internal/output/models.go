package output

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"math"
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

	headers := []string{
		"NAME",
		"DEVELOPER",
		"PROVIDERS",
		"CONTEXT",
		"INPUT",
		"OUTPUT",
		"REGION",
		"RECOMMENDED",
		"ID",
	}
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
			formatTokenCount(m.ContextLength),
			formatModalities(m.InputModalities),
			formatModalities(m.OutputModalities),
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
		fmt.Fprintf(w, "Context:\t%s\n", formatTokenCount(m.ContextLength))
	}
	fmt.Fprintf(w, "Input modalities:\t%s\n", formatModalities(m.InputModalities))
	fmt.Fprintf(w, "Output modalities:\t%s\n", formatModalities(m.OutputModalities))
	if m.TokenizerID != "" {
		fmt.Fprintf(w, "Tokenizer:\t%s\n", m.TokenizerID)
	}
	if m.CreatedAt != "" {
		fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(m.CreatedAt))
	}
	printRouterModelPrices(w, m.Prices)
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
		"INPUT ($/1M)",
		"OUTPUT ($/1M)",
		"CACHE READ ($/1M)",
		"CONTEXT",
		"MAX OUTPUT",
		"UPTIME (1D)",
		"LATENCY (P50 30M)",
		"TOK/S (P50 30M)",
	}
	rows := make([][]string, len(m.Providers))
	for i, p := range m.Providers {
		throughput := "-"
		if v := routerModelP50(p.ThroughputLast30m); v != nil {
			throughput = strconv.FormatFloat(math.Round(*v), 'f', 0, 64)
		}
		rows[i] = []string{
			p.Name,
			cmp.Or(derefString(p.Tag), "-"),
			routerModelPrice(p.Prices, "input"),
			routerModelPrice(p.Prices, "output"),
			routerModelPrice(p.Prices, "input_cache_read"),
			formatTokenCount(p.ContextLength),
			formatTokenCount(p.MaxOutputTokens),
			formatFloat(p.UptimeLastDay, "%"),
			formatLatencyMs(routerModelP50(p.LatencyLast30m)),
			throughput,
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

type routerModelPriceUnit struct {
	scale float64
	label string
}

var routerModelPriceUnits = map[string]routerModelPriceUnit{
	"input":                {scale: 1e6},
	"output":               {scale: 1e6},
	"input_cache_read":     {scale: 1e6},
	"input_cache_creation": {scale: 1e6},
	"output_reasoning":     {scale: 1e6},
	"search_units":         {scale: 1, label: "per query"},
	"input_characters":     {scale: 1e6, label: "per 1M characters"},
	"input_seconds":        {scale: 3600, label: "per hour"},
}

func routerModelPriceText(key string, price float64) string {
	unit, known := routerModelPriceUnits[key]
	if !known {
		return formatOptionalFloat(&price)
	}
	// Upstream prices are cut at 16 decimal places, so scaling exposes a tail.
	v := roundSignificant(price*unit.scale, 10)
	out := formatCost(&v)
	if unit.label != "" {
		out += " " + unit.label
	}
	return out
}

func routerModelPrice(prices map[string]float64, key string) string {
	v, ok := prices[key]
	if !ok {
		return "-"
	}
	return routerModelPriceText(key, v)
}

func printRouterModelPrices(w io.Writer, prices map[string]float64) {
	var perToken, other []string
	for _, k := range slices.Sorted(maps.Keys(prices)) {
		if unit, known := routerModelPriceUnits[k]; known && unit.label == "" {
			perToken = append(perToken, k)
		} else {
			other = append(other, k)
		}
	}
	for _, group := range []struct {
		title string
		keys  []string
	}{
		{"List price (per 1M tokens):", perToken},
		{"List price:", other},
	} {
		if len(group.keys) == 0 {
			continue
		}
		fmt.Fprintln(w, group.title)
		for _, k := range group.keys {
			fmt.Fprintf(w, "  %s:\t%s\n", k, routerModelPriceText(k, prices[k]))
		}
	}
}

func routerModelP50(p *platform.RouterModelPercentiles) *float64 {
	if p == nil {
		return nil
	}
	return p.P50
}

func formatModalities(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
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
