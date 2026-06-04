package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// PrintStructuredJSON writes value as pretty-printed JSON.
func PrintStructuredJSON(out io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	_, err = out.Write(append(data, '\n'))
	return err
}

// PrintRawJSON writes pretty-printed JSON to the writer.
func PrintRawJSON(out io.Writer, raw json.RawMessage) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		_, err := out.Write(raw)
		return err
	}
	buf.WriteByte('\n')
	_, err := buf.WriteTo(out)
	return err
}

// PrintStructuredYAML writes value as YAML using its JSON representation.
func PrintStructuredYAML(out io.Writer, value any) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return PrintRawYAML(out, jsonData)
}

// PrintRawYAML writes a raw JSON payload as YAML.
func PrintRawYAML(out io.Writer, raw json.RawMessage) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	data, err := yaml.Marshal(normalizeYAMLValue(normalized))
	if err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	_, err = out.Write(data)
	return err
}

func normalizeYAMLValue(value any) any {
	switch v := value.(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}
		if f, err := v.Float64(); err == nil {
			return f
		}
		return v.String()
	case map[string]any:
		for key, item := range v {
			v[key] = normalizeYAMLValue(item)
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = normalizeYAMLValue(item)
		}
		return v
	default:
		return value
	}
}
