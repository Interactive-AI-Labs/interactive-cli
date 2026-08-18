package preflight

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

func PrintUpdateBanner(w io.Writer, target string, revision int, updated string) {
	var b strings.Builder
	b.WriteString("Live:")
	if target != "" {
		b.WriteString(" ")
		b.WriteString(target)
	}
	fmt.Fprintf(&b, " revision %d", revision)
	if ts := formatUpdated(updated); ts != "" {
		b.WriteString(", last updated ")
		b.WriteString(ts)
	}
	fmt.Fprintf(&b, " — this update creates revision %d", revision+1)
	fmt.Fprintln(w, b.String())
}

func PrintFailOpenNote(w io.Writer, what string, err error) {
	fmt.Fprintf(w, "⚠ could not %s (%v) — proceeding without pre-flight check\n", what, err)
}

func PrintSyncDeletions(w io.Writer, resource, allowFlag string, names []string, allowed bool) {
	if len(names) == 0 {
		return
	}
	plural := ""
	if len(names) != 1 {
		plural = "s"
	}
	verb := "NOT delete"
	hint := fmt.Sprintf(
		"a config that omits a resource looks identical to a stale one — pass --allow-delete=%s to delete",
		allowFlag,
	)
	if allowed {
		verb = "DELETE"
		hint = fmt.Sprintf(
			"--allow-delete=%s; %s deletes run after %s creates/updates",
			allowFlag, resource, resource,
		)
	}
	fmt.Fprintf(
		w,
		"⚠ sync will %s %d %s%s not in the config: %s (%s)\n",
		verb,
		len(names),
		resource,
		plural,
		strings.Join(names, ", "),
		hint,
	)
}

func PrintDroppedListEntries(w io.Writer, kind, flag string, live, incoming []string) bool {
	keep := make(map[string]bool, len(incoming))
	for _, name := range incoming {
		keep[name] = true
	}
	var dropped []string
	for _, name := range live {
		if !keep[name] {
			dropped = append(dropped, name)
		}
	}
	if len(dropped) == 0 {
		return false
	}
	sort.Strings(dropped)
	fmt.Fprintf(
		w,
		"⚠ this update drops live %s: %s (%s replaces the entire list; pass every value you want to keep)\n",
		kind,
		strings.Join(dropped, ", "),
		flag,
	)
	return true
}

func formatUpdated(ts string) string {
	if ts == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

// description is the legacy system_prompt key (prompt_id).
var pinSections = map[string]string{
	"system_prompt": "system_prompt",
	"description":   "description",
	"policies":      "policy",
	"routines":      "routine",
	"glossaries":    "glossary",
	"macros":        "macro",
}

type pinKey struct {
	section string
	id      string
	known   bool
}

type pinVersion struct {
	display string
	num     float64
	numeric bool
}

func (p pinVersion) String() string {
	if p.display == "" {
		return "(unpinned)"
	}
	return "v" + p.display
}

func parseVersion(v any) pinVersion {
	switch n := v.(type) {
	case nil:
		return pinVersion{}
	case int:
		return pinVersion{display: strconv.Itoa(n), num: float64(n), numeric: true}
	case int64:
		return pinVersion{display: strconv.FormatInt(n, 10), num: float64(n), numeric: true}
	case float64:
		return pinVersion{display: strconv.FormatFloat(n, 'f', -1, 64), num: n, numeric: true}
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return pinVersion{display: n, num: f, numeric: true}
		}
		return pinVersion{display: n}
	default:
		return pinVersion{display: fmt.Sprintf("%v", v)}
	}
}

func extractPins(config any) map[pinKey]pinVersion {
	pins := map[pinKey]pinVersion{}
	cfg, ok := config.(map[string]any)
	if !ok {
		return pins
	}
	ctx, ok := cfg["context"].(map[string]any)
	if !ok {
		return pins
	}
	for section, v := range ctx {
		label, known := pinSections[section]
		if !known {
			label = section
		}
		switch v := v.(type) {
		case map[string]any:
			addPin(pins, label, known, v)
		case []any:
			for _, entry := range v {
				if m, ok := entry.(map[string]any); ok {
					addPin(pins, label, known, m)
				}
			}
		}
	}
	return pins
}

func addPin(pins map[pinKey]pinVersion, label string, known bool, entry map[string]any) {
	id, _ := entry["id"].(string)
	if id == "" {
		id, _ = entry["prompt_id"].(string)
	}
	if id == "" {
		return
	}
	if !known {
		if _, hasVersion := entry["version"]; !hasVersion {
			return
		}
	}
	pins[pinKey{section: label, id: id, known: known}] = parseVersion(entry["version"])
}

func PrintPinChanges(w io.Writer, liveConfig, incomingConfig any) bool {
	livePins := extractPins(liveConfig)
	incomingPins := extractPins(incomingConfig)

	keys := make([]pinKey, 0, len(livePins)+len(incomingPins))
	seen := map[pinKey]bool{}
	for k := range livePins {
		keys = append(keys, k)
		seen[k] = true
	}
	for k := range incomingPins {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].section != keys[j].section {
			return keys[i].section < keys[j].section
		}
		return keys[i].id < keys[j].id
	})

	var lines []string
	loud := false
	blocking := false
	for _, k := range keys {
		from, inLive := livePins[k]
		to, inIncoming := incomingPins[k]
		switch {
		case inLive && !inIncoming:
			lines = append(lines, fmt.Sprintf(
				"%s %s: %s → (none)  (REMOVED — this update drops the pin entirely)",
				k.section, k.id, from,
			))
			loud = true
			blocking = blocking || k.known
		case !inLive && inIncoming:
			lines = append(lines, fmt.Sprintf("%s %s: (none) → %s", k.section, k.id, to))
		case from.String() == to.String(),
			from.numeric && to.numeric && from.num == to.num:
		case from.display != "" && to.display == "":
			// Unpinning has the same effect as removing the pin.
			lines = append(lines, fmt.Sprintf(
				"%s %s: %s → %s  (REMOVED — this update drops the version pin)",
				k.section, k.id, from, to,
			))
			loud = true
			blocking = blocking || k.known
		case from.numeric && to.numeric && to.num < from.num:
			lines = append(lines, fmt.Sprintf(
				"%s %s: %s → %s  (DOWNGRADE — if this is an intentional rollback, this is expected)",
				k.section,
				k.id,
				from,
				to,
			))
			loud = true
			blocking = blocking || k.known
		default:
			lines = append(lines, fmt.Sprintf("%s %s: %s → %s", k.section, k.id, from, to))
		}
	}
	if len(lines) == 0 {
		return false
	}

	header := "content pins changed by this update:"
	if loud {
		header = "⚠ " + header
	}
	fmt.Fprintln(w, header)
	for _, l := range lines {
		fmt.Fprintln(w, "    "+l)
	}
	return blocking
}
