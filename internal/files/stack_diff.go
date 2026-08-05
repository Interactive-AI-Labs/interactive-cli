package files

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// StackDiff holds the result of comparing two StackConfig values.
type StackDiff struct {
	StackID   string           `json:"stackId"`
	Services  ResourceTypeDiff `json:"services"`
	Agents    ResourceTypeDiff `json:"agents"`
	Databases ResourceTypeDiff `json:"databases"`
}

// ResourceTypeDiff records which resources were created, updated, or deleted.
type ResourceTypeDiff struct {
	Created []string         `json:"created"`
	Updated []ResourceChange `json:"updated"`
	Deleted []string         `json:"deleted"`
}

// ResourceChange describes a resource that differs between local and live.
type ResourceChange struct {
	Name    string               `json:"name"`
	Changes map[string]FieldDiff `json:"changes,omitempty"`
}

// FieldDiff describes a single field-level change.
type FieldDiff struct {
	Old string `json:"old,omitempty"`
	New string `json:"new,omitempty"`
}

type fieldChange struct {
	path string
	old  string
	new  string
}

func DiffStackConfigs(local, live *StackConfig) *StackDiff {
	d := &StackDiff{StackID: live.StackId}
	d.Services = diffResourceMap(
		local.Services, live.Services,
		func(a, b ServiceConfig) []fieldChange { return diffFields(a, b) },
	)
	d.Agents = diffResourceMap(
		local.Agents, live.Agents,
		func(a, b AgentConfig) []fieldChange { return diffFields(a, b) },
	)
	d.Databases = diffResourceMap(
		local.Databases, live.Databases,
		func(a, b DatabaseConfig) []fieldChange { return diffFields(a, b) },
	)
	return d
}

func (d *StackDiff) HasChanges() bool {
	return len(d.Services.Created)+len(d.Services.Updated)+len(d.Services.Deleted)+
		len(d.Agents.Created)+len(d.Agents.Updated)+len(d.Agents.Deleted)+
		len(d.Databases.Created)+len(d.Databases.Updated)+len(d.Databases.Deleted) > 0
}

func diffResourceMap[T any](
	local, live map[string]T,
	fieldDiffs func(a, b T) []fieldChange,
) ResourceTypeDiff {
	var d ResourceTypeDiff

	for name := range local {
		if _, ok := live[name]; !ok {
			d.Created = append(d.Created, name)
		}
	}
	sort.Strings(d.Created)

	for name := range live {
		if _, ok := local[name]; !ok {
			d.Deleted = append(d.Deleted, name)
		}
	}
	sort.Strings(d.Deleted)

	for name := range local {
		if liveVal, ok := live[name]; ok {
			changes := fieldDiffs(liveVal, local[name])
			if len(changes) > 0 {
				rc := ResourceChange{Name: name, Changes: make(map[string]FieldDiff)}
				for _, ch := range changes {
					rc.Changes[ch.path] = FieldDiff{Old: ch.old, New: ch.new}
				}
				d.Updated = append(d.Updated, rc)
			}
		}
	}
	sort.Slice(d.Updated, func(i, j int) bool {
		return d.Updated[i].Name < d.Updated[j].Name
	})

	return d
}

func PrintStackDiffDetailed(
	out io.Writer,
	local, live *StackConfig,
	d *StackDiff,
) error {
	if !d.HasChanges() {
		fmt.Fprintln(out, "No differences found.")
		return nil
	}

	fmt.Fprintf(out, "Stack: %s\n", d.StackID)

	// Re-computes diffFields for display ordering; the sorted
	// []fieldChange gives stable human output unlike the map in ResourceChange.
	printSection(out, "service", d.Services, func(name string) []fieldChange {
		return diffFields(live.Services[name], local.Services[name])
	})
	printSection(out, "agent", d.Agents, func(name string) []fieldChange {
		return diffFields(live.Agents[name], local.Agents[name])
	})
	printSection(out, "database", d.Databases, func(name string) []fieldChange {
		return diffFields(live.Databases[name], local.Databases[name])
	})

	return nil
}

func printSection(
	out io.Writer,
	kind string,
	d ResourceTypeDiff,
	changes func(name string) []fieldChange,
) {
	for _, name := range d.Created {
		fmt.Fprintf(out, "\n  + %s %s (create)\n", kind, name)
	}
	for _, rc := range d.Updated {
		fmt.Fprintf(out, "\n  ~ %s %s (update)\n", kind, rc.Name)
		for _, ch := range changes(rc.Name) {
			switch {
			case ch.old == "":
				fmt.Fprintf(out, "    + %s: %s\n", ch.path, ch.new)
			case ch.new == "":
				fmt.Fprintf(out, "    - %s\n", ch.path)
			default:
				fmt.Fprintf(out, "    %s: %s → %s\n", ch.path, ch.old, ch.new)
			}
		}
	}
	for _, name := range d.Deleted {
		fmt.Fprintf(out, "\n  - %s %s (delete)\n", kind, name)
	}
}

// diffFields returns a sorted list of field-level changes between two config
// values (live → local). An empty string for old means the field is new; an
// empty string for new means the field was removed.
func diffFields(live, local any) []fieldChange {
	liveMap := toFlatMap(live)
	localMap := toFlatMap(local)

	var changes []fieldChange

	for k, v := range localMap {
		oldV, exists := liveMap[k]
		if !exists {
			changes = append(changes, fieldChange{path: k, new: v})
		} else if oldV != v {
			changes = append(changes, fieldChange{path: k, old: oldV, new: v})
		}
	}
	for k := range liveMap {
		if _, exists := localMap[k]; !exists {
			changes = append(changes, fieldChange{path: k, old: liveMap[k]})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].path < changes[j].path
	})
	return changes
}

// toFlatMap marshals a value to JSON, then flattens it to a key→value map
// where keys are dot-separated paths (e.g. "image.tag", "resources.cpu").
func toFlatMap(v any) map[string]string {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	result := make(map[string]string)
	flatten(result, "", raw)
	return result
}

func flatten(out map[string]string, prefix string, v any) {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			childPrefix := k
			if prefix != "" {
				childPrefix = prefix + "." + k
			}
			flatten(out, childPrefix, val[k])
		}
	case []any:
		for i, item := range val {
			childPrefix := fmt.Sprintf("%s[%d]", prefix, i)
			flatten(out, childPrefix, item)
		}
	case bool:
		if val {
			out[prefix] = "true"
		} else {
			out[prefix] = "false"
		}
	case nil:
		out[prefix] = "null"
	default:
		out[prefix] = fmt.Sprintf("%v", val)
	}
}
