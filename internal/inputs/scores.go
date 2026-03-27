package inputs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var DefaultScoreColumns = []string{
	"id",
	"name",
	"data_type",
	"value",
	"source",
	"timestamp",
	"trace_id",
}

var AllScoreColumns = []string{
	"id",
	"name",
	"data_type",
	"value",
	"source",
	"timestamp",
	"trace_id",
	"observation_id",
	"session_id",
	"environment",
	"config_id",
	"user_id",
	"comment",
}

type ScoreCreateInput struct {
	ID            string
	Name          string
	TraceID       string
	ObservationID string
	SessionID     string
	DataType      string
	Value         string
	Comment       string
	MetadataJSON  string
	Environment   string
	ConfigID      string
	QueueID       string
}

func PrepareScoreListOptions(opts clients.ScoreListOptions) (clients.ScoreListOptions, error) {
	if err := validateTimestamp(opts.FromTimestamp, "from-timestamp"); err != nil {
		return opts, err
	}
	if err := validateTimestamp(opts.ToTimestamp, "to-timestamp"); err != nil {
		return opts, err
	}
	if opts.Limit < 0 {
		return opts, fmt.Errorf("limit must be non-negative, got %d", opts.Limit)
	}
	if opts.Value != "" && strings.TrimSpace(opts.Operator) == "" {
		opts.Operator = "="
	}
	if opts.Value != "" && (opts.MinValue != nil || opts.MaxValue != nil) {
		return opts, fmt.Errorf(
			"--value/--operator cannot be combined with --min-value/--max-value",
		)
	}
	if opts.MinValue != nil && opts.MaxValue != nil && *opts.MinValue > *opts.MaxValue {
		return opts, fmt.Errorf("--min-value cannot be greater than --max-value")
	}
	return opts, nil
}

func BuildScoreCreateBody(input ScoreCreateInput) (clients.ScoreCreateBody, error) {
	targetCount := 0
	if strings.TrimSpace(input.TraceID) != "" {
		targetCount++
	}
	if strings.TrimSpace(input.ObservationID) != "" {
		targetCount++
	}
	if strings.TrimSpace(input.SessionID) != "" {
		targetCount++
	}
	if targetCount != 1 {
		return clients.ScoreCreateBody{}, fmt.Errorf(
			"exactly one of --trace-id, --observation-id, or --session-id is required",
		)
	}

	valueRaw := strings.TrimSpace(input.Value)

	dataType := strings.ToUpper(strings.TrimSpace(input.DataType))
	if dataType == "" {
		dataType = "NUMERIC"
	}

	parsedValue, err := parseScoreValue(dataType, valueRaw)
	if err != nil {
		return clients.ScoreCreateBody{}, err
	}

	body := clients.ScoreCreateBody{
		ID:            strings.TrimSpace(input.ID),
		Name:          strings.TrimSpace(input.Name),
		TraceID:       strings.TrimSpace(input.TraceID),
		ObservationID: strings.TrimSpace(input.ObservationID),
		SessionID:     strings.TrimSpace(input.SessionID),
		DataType:      dataType,
		Value:         parsedValue,
		Comment:       strings.TrimSpace(input.Comment),
		Environment:   strings.TrimSpace(input.Environment),
		ConfigID:      strings.TrimSpace(input.ConfigID),
		QueueID:       strings.TrimSpace(input.QueueID),
	}

	if strings.TrimSpace(input.MetadataJSON) != "" {
		metadata, err := parseJSONObject(
			input.MetadataJSON,
			"--metadata-json",
		)
		if err != nil {
			return clients.ScoreCreateBody{}, err
		}
		body.Metadata = metadata
	}

	return body, nil
}

func parseScoreValue(dataType, raw string) (any, error) {
	switch dataType {
	case "BOOLEAN":
		switch strings.ToLower(raw) {
		case "true", "1", "yes":
			return 1, nil
		case "false", "0", "no":
			return 0, nil
		default:
			return nil, fmt.Errorf(
				"invalid BOOLEAN score value %q; use true, false, 1, 0, yes, or no",
				raw,
			)
		}
	case "NUMERIC":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid NUMERIC score value %q", raw)
		}
		return value, nil
	default:
		return raw, nil
	}
}

