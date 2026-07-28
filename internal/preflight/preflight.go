// Package preflight prints deploy-awareness warnings: before a mutating
// deploy command writes, it surfaces the live state the caller is about to
// replace or destroy (revision, content pin deltas, existing image tags,
// resources a sync will delete, env/secret entries a list replacement
// drops) so the operator — human or LLM agent — can notice a stale-based
// deploy at the moment it can still be stopped.
//
// Output contract: deterministic plain text intended for stderr, stable "⚠"
// prefix on warnings, no colors, no prompts. Checks warn but never block;
// callers fail open when live state cannot be fetched.
package preflight

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PrintUpdateBanner prints the live revision a pending update replaces.
// target names the resource (e.g. "agent chat-agent") for multi-resource
// flows like sync; leave it empty when the command line already names it.
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

// PrintFailOpenNote reports that a pre-flight lookup failed and the command
// is proceeding without it. what describes the lookup, e.g. "fetch live state".
func PrintFailOpenNote(w io.Writer, what string, err error) {
	fmt.Fprintf(w, "⚠ could not %s (%v) — proceeding without pre-flight check\n", what, err)
}

// PrintTagOverwriteWarning warns that pushing to an already-existing tag
// replaces the previous image with no way to recover it.
func PrintTagOverwriteWarning(w io.Writer, tag string) {
	fmt.Fprintf(
		w,
		"⚠ tag %s already exists upstream — pushing replaces it; the previous image is unrecoverable.\n",
		tag,
	)
}

// PrintSyncDeletions announces, before that resource type's sync phase
// writes anything, the resources it will delete because the config file
// does not mention them. A stale stack file doesn't say "delete X" — it just
// stops listing X — so this is the only moment decommissioning looks
// different from a forgotten pull. Deletes run after creates/updates for
// that resource type. resource is the singular noun ("service", "agent").
func PrintSyncDeletions(w io.Writer, resource string, names []string) {
	if len(names) == 0 {
		return
	}
	plural := ""
	if len(names) != 1 {
		plural = "s"
	}
	fmt.Fprintf(
		w,
		"⚠ sync will DELETE %d %s%s not in the config: %s (%s deletes run after %s creates/updates — Ctrl-C to abort if the config is stale)\n",
		len(names),
		resource,
		plural,
		strings.Join(names, ", "),
		resource,
		resource,
	)
}

// PrintDroppedListEntries warns when a full-list replacement flag (--env,
// --secret) drops entries that exist on the live resource. The flags
// replace, not merge — the classic mistake is passing just the one new value
// and silently wiping the rest. Only real disappearances warn: kept and
// added names print nothing. kind is the plural noun ("env vars",
// "secret refs"); live and incoming are entry names.
func PrintDroppedListEntries(w io.Writer, kind, flag string, live, incoming []string) {
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
		return
	}
	sort.Strings(dropped)
	fmt.Fprintf(
		w,
		"⚠ this update drops live %s: %s (%s replaces the entire list; pass every value you want to keep)\n",
		kind,
		strings.Join(dropped, ", "),
		flag,
	)
}

// CheckExpectedRevision enforces --expect-revision: the update must not be
// applied when the live revision differs from what the caller expects.
func CheckExpectedRevision(expected, live int) error {
	if live != expected {
		return fmt.Errorf(
			"live revision is %d, expected %d — not applying (--expect-revision)",
			live, expected,
		)
	}
	return nil
}

// formatUpdated renders an RFC3339 timestamp as "2006-01-02 15:04 UTC",
// falling back to the raw string when it doesn't parse.
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

// pinSections maps the agent_config.context keys that pin content versions
// to their singular display labels.
var pinSections = map[string]string{
	"system_prompt": "system_prompt",
	"policies":      "policy",
	"routines":      "routine",
}

type pinKey struct {
	section string
	id      string
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

// extractPins collects (section, id) → version from an agent config's
// context block. The config schema is server-owned and opaque to the CLI,
// so anything not shaped like a pin is ignored rather than an error.
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
	for section, label := range pinSections {
		switch v := ctx[section].(type) {
		case map[string]any:
			addPin(pins, label, v)
		case []any:
			for _, entry := range v {
				if m, ok := entry.(map[string]any); ok {
					addPin(pins, label, m)
				}
			}
		}
	}
	return pins
}

func addPin(pins map[pinKey]pinVersion, label string, entry map[string]any) {
	id, _ := entry["id"].(string)
	if id == "" {
		return
	}
	pins[pinKey{section: label, id: id}] = parseVersion(entry["version"])
}

// PrintPinChanges compares content pins (system_prompt / policies /
// routines) between the live agent config and the incoming replacement and
// prints the delta. A full-config update replaces every pin, so a stale
// manifest silently reverts colleagues' work: downgrades and removals are
// the loud cases. Additions and upgrades print quietly, and an unchanged
// pin set prints nothing at all.
//
// Rollbacks are downgrades by definition, so the downgrade wording is
// informational, never prohibitive — an agent mid-incident must not refuse
// the documented recovery procedure.
func PrintPinChanges(w io.Writer, liveConfig, incomingConfig any) {
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
		case !inLive && inIncoming:
			lines = append(lines, fmt.Sprintf("%s %s: (none) → %s", k.section, k.id, to))
		case from.String() == to.String(),
			from.numeric && to.numeric && from.num == to.num:
			// unchanged (numeric compare catches "13" vs 13)
		case from.numeric && to.numeric && to.num < from.num:
			lines = append(lines, fmt.Sprintf(
				"%s %s: %s → %s  (DOWNGRADE — if this is an intentional rollback, this is expected)",
				k.section,
				k.id,
				from,
				to,
			))
			loud = true
		default:
			lines = append(lines, fmt.Sprintf("%s %s: %s → %s", k.section, k.id, from, to))
		}
	}
	if len(lines) == 0 {
		return
	}

	header := "content pins changed by this update:"
	if loud {
		header = "⚠ " + header
	}
	fmt.Fprintln(w, header)
	for _, l := range lines {
		fmt.Fprintln(w, "    "+l)
	}
}
