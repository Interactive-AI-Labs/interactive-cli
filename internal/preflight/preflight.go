// Package preflight detects deploy accidents: before a mutating deploy
// command writes, it surfaces the live state the caller is about to replace
// or destroy (revision, content pin deltas, existing image tags, resources
// a sync would delete, env/secret entries a list replacement drops) so the
// operator — human or LLM agent — can notice a stale-based deploy at the
// moment it can still be stopped.
//
// Output contract: deterministic plain text intended for stderr, stable "⚠"
// prefix on warnings, no colors, no prompts. Destruction implied by
// omission (a sync deletion, a pin downgrade/removal, a dropped env/secret
// entry, an occupied image tag) gates: the caller refuses until an explicit
// flag (--allow-delete, --force) names the intent. Everything ambiguous
// only warns, and callers fail open when live state cannot be fetched.
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

// PrintSyncDeletions reports, before that resource type's sync phase writes
// anything, the resources the config file no longer mentions. A stale stack
// file doesn't say "delete X" — it just stops listing X — so deletion is
// refused unless the matching --allow-delete value was passed; the refusal
// names the flag that unblocks it. resource is the singular noun
// ("service", "agent"); allowFlag is the --allow-delete value ("services",
// "agents").
func PrintSyncDeletions(w io.Writer, resource, allowFlag string, names []string, allowed bool) {
	if len(names) == 0 {
		return
	}
	plural := ""
	if len(names) != 1 {
		plural = "s"
	}
	if allowed {
		fmt.Fprintf(
			w,
			"⚠ sync will DELETE %d %s%s not in the config: %s (--allow-delete=%s; %s deletes run after %s creates/updates)\n",
			len(names),
			resource,
			plural,
			strings.Join(names, ", "),
			allowFlag,
			resource,
			resource,
		)
		return
	}
	fmt.Fprintf(
		w,
		"⚠ sync will NOT delete %d %s%s not in the config: %s (a config that omits a resource looks identical to a stale one — pass --allow-delete=%s to delete)\n",
		len(names),
		resource,
		plural,
		strings.Join(names, ", "),
		allowFlag,
	)
}

// PrintDroppedListEntries warns when a full-list replacement flag (--env,
// --secret) drops entries that exist on the live resource, and reports
// whether it did — callers gate on the result. The flags replace, not merge
// — the classic mistake is passing just the one new value and silently
// wiping the rest. Only real disappearances warn: kept and added names
// print nothing. kind is the plural noun ("env vars", "secret refs"); live
// and incoming are entry names.
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

// pinSections maps the agent_config.context keys known to pin content
// versions to their singular display labels. Known sections gate (their
// downgrades/removals block without --force); any other context section
// whose entries are {id, version}-shaped is picked up by shape and warns
// only — never block on guessed semantics.
var pinSections = map[string]string{
	"system_prompt": "system_prompt",
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

// extractPins collects (section, id) → version from an agent config's
// context block. The config schema is server-owned and opaque to the CLI,
// so anything not shaped like a pin is ignored rather than an error.
// Sections not in pinSections are scanned by shape: an entry counts as a
// pin only when it carries both an id and a version, and is marked
// unknown so its changes warn without gating.
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
		return
	}
	if !known {
		if _, hasVersion := entry["version"]; !hasVersion {
			return
		}
	}
	pins[pinKey{section: label, id: id, known: known}] = parseVersion(entry["version"])
}

// PrintPinChanges compares content pins between the live agent config and
// the incoming replacement and prints the delta. A full-config update
// replaces every pin, so a stale manifest silently reverts colleagues'
// work: downgrades and removals are the loud cases. Additions and upgrades
// print quietly, and an unchanged pin set prints nothing at all.
//
// The return value reports whether a known pin section (pinSections)
// downgraded or removed a pin — the caller gates on it and refuses without
// --force. Loud changes in unknown-but-pin-shaped sections warn but never
// gate: their semantics are guessed from shape alone.
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
			// unchanged (numeric compare catches "13" vs 13)
		case from.display != "" && to.display == "":
			// The entry stays but loses its version: same effect as removing
			// the pin, so it gates the same way.
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
