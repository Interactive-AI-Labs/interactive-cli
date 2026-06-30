package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type jsonSchema struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Type        string                `json:"type"`
	Properties  map[string]jsonSchema `json:"properties"`
	Required    []string              `json:"required"`
	Defs        map[string]jsonSchema `json:"$defs"`
	Ref         string                `json:"$ref"`
	Items       *jsonSchema           `json:"items"`
	AnyOf       []jsonSchema          `json:"anyOf"`
	OneOf       []jsonSchema          `json:"oneOf"`
	Enum        []any                 `json:"enum"`
	Default     any                   `json:"default"`
	Const       any                   `json:"const"`
}

func PrintSchemaPretty(out io.Writer, raw json.RawMessage, version string) error {
	var root jsonSchema
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	fmt.Fprintf(out, "Schema version: %s\n", version)
	fmt.Fprintf(out, "Full documentation: https://docs.interactive.ai/schemas\n")

	if err := printTypeTable(out, root, root.Defs); err != nil {
		return err
	}

	names := make([]string, 0, len(root.Defs))
	for name := range root.Defs {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		def := root.Defs[name]
		def.Title = name
		if err := printTypeTable(out, def, root.Defs); err != nil {
			return err
		}
	}

	return nil
}

func printTypeTable(out io.Writer, s jsonSchema, defs map[string]jsonSchema) error {
	if len(s.Properties) == 0 {
		return nil
	}

	title := s.Title
	if title == "" {
		title = "Schema"
	}

	fmt.Fprintf(out, "\n%s\n", title)
	if s.Description != "" {
		fmt.Fprintf(out, "  %s\n", s.Description)
	}
	fmt.Fprintln(out)

	reqSet := make(map[string]bool, len(s.Required))
	for _, r := range s.Required {
		reqSet[r] = true
	}

	type field struct {
		name, typ, req, def, desc string
	}

	fields := make([]field, 0, len(s.Properties))
	for name, prop := range s.Properties {
		req := "no"
		if reqSet[name] {
			req = "yes"
		}
		fields = append(fields, field{
			name: name,
			typ:  resolveTypeName(prop, defs),
			req:  req,
			def:  resolveDefault(prop),
			desc: prop.Description,
		})
	}

	sort.Slice(fields, func(i, j int) bool {
		if fields[i].req != fields[j].req {
			return fields[i].req == "yes"
		}
		return fields[i].name < fields[j].name
	})

	headers := []string{"FIELD", "TYPE", "REQUIRED", "DEFAULT", "DESCRIPTION"}
	rows := make([][]string, len(fields))
	for i, f := range fields {
		rows[i] = []string{f.name, f.typ, f.req, f.def, f.desc}
	}
	return PrintTable(out, headers, rows)
}

func resolveTypeName(s jsonSchema, defs map[string]jsonSchema) string {
	if s.Ref != "" {
		return refName(s.Ref)
	}

	// Discriminated unions nest a oneOf inside an anyOf; union both alike.
	if alts := append(append([]jsonSchema{}, s.AnyOf...), s.OneOf...); len(alts) > 0 {
		parts := make([]string, 0, len(alts))
		for _, alt := range alts {
			parts = append(parts, resolveTypeName(alt, defs))
		}
		return strings.Join(parts, " | ")
	}

	if len(s.Enum) > 0 {
		vals := make([]string, len(s.Enum))
		for i, v := range s.Enum {
			vals[i] = fmt.Sprintf("%v", v)
		}
		return "enum(" + strings.Join(vals, ", ") + ")"
	}

	if s.Const != nil {
		return fmt.Sprintf("const(%v)", s.Const)
	}

	if s.Type == "array" && s.Items != nil {
		return "list[" + resolveTypeName(*s.Items, defs) + "]"
	}

	if s.Type != "" {
		return s.Type
	}

	return "any"
}

func resolveDefault(s jsonSchema) string {
	if s.Default == nil {
		return "-"
	}
	switch v := s.Default.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%g", v)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func refName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
