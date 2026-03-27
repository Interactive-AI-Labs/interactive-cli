package output

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var scoreConfigColumnMap = map[string]struct {
	Header string
	Value  func(s *clients.ScoreConfigInfo) string
}{
	"id":        {"ID", func(s *clients.ScoreConfigInfo) string { return s.ID }},
	"name":      {"NAME", func(s *clients.ScoreConfigInfo) string { return s.Name }},
	"data_type": {"DATA TYPE", func(s *clients.ScoreConfigInfo) string { return s.DataType }},
	"is_archived": {
		"ARCHIVED",
		func(s *clients.ScoreConfigInfo) string { return strconv.FormatBool(s.IsArchived) },
	},
	"min_value": {
		"MIN VALUE",
		func(s *clients.ScoreConfigInfo) string { return formatOptionalFloat(s.MinValue) },
	},
	"max_value": {
		"MAX VALUE",
		func(s *clients.ScoreConfigInfo) string { return formatOptionalFloat(s.MaxValue) },
	},
	"description": {
		"DESCRIPTION",
		func(s *clients.ScoreConfigInfo) string { return s.Description },
	},
	"created_at": {
		"CREATED AT",
		func(s *clients.ScoreConfigInfo) string { return LocalTime(s.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(s *clients.ScoreConfigInfo) string { return LocalTime(s.UpdatedAt) },
	},
}

func formatOptionalFloat(v *float64) string {
	if v == nil {
		return "-"
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func PrintScoreConfigList(
	out io.Writer,
	configs []clients.ScoreConfigInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(configs) == 0 {
		fmt.Fprintln(out, "No score configs found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := scoreConfigColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(configs))
	for i, cfg := range configs {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := scoreConfigColumnMap[col]; ok {
				row[j] = def.Value(&cfg)
			}
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	PrintPageMeta(out, meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

func PrintScoreConfigDetail(out io.Writer, cfg *clients.ScoreConfigInfo) error {
	fmt.Fprintf(out, "ID:          %s\n", cfg.ID)
	fmt.Fprintf(out, "Name:        %s\n", cfg.Name)
	fmt.Fprintf(out, "Data Type:   %s\n", cfg.DataType)
	fmt.Fprintf(out, "Archived:    %v\n", cfg.IsArchived)
	if cfg.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", cfg.Description)
	}
	fmt.Fprintf(out, "Min Value:   %s\n", formatOptionalFloat(cfg.MinValue))
	fmt.Fprintf(out, "Max Value:   %s\n", formatOptionalFloat(cfg.MaxValue))
	if len(cfg.Categories) > 0 && string(cfg.Categories) != "null" {
		fmt.Fprintf(out, "Categories:  %s\n", string(cfg.Categories))
	}
	fmt.Fprintf(out, "Created At:  %s\n", LocalTime(cfg.CreatedAt))
	fmt.Fprintf(out, "Updated At:  %s\n", LocalTime(cfg.UpdatedAt))
	return nil
}

func PrintScoreConfigCreateResult(out io.Writer, cfg *clients.ScoreConfigInfo) error {
	fmt.Fprintf(out, "Created score config %q.\n", cfg.Name)
	fmt.Fprintf(out, "ID:        %s\n", cfg.ID)
	fmt.Fprintf(out, "Name:      %s\n", cfg.Name)
	fmt.Fprintf(out, "Data Type: %s\n", cfg.DataType)
	fmt.Fprintf(out, "Created At: %s\n", LocalTime(cfg.CreatedAt))
	return nil
}

func PrintScoreConfigUpdateResult(out io.Writer, cfg *clients.ScoreConfigInfo) error {
	fmt.Fprintf(out, "Updated score config %q.\n", cfg.ID)
	fmt.Fprintf(out, "Name:      %s\n", cfg.Name)
	fmt.Fprintf(out, "Archived:  %v\n", cfg.IsArchived)
	fmt.Fprintf(out, "Updated At: %s\n", LocalTime(cfg.UpdatedAt))
	return nil
}
