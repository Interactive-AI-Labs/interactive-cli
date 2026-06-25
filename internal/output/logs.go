package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

type logEntry struct {
	Replica   string `json:"replica,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Line      string `json:"line"`
}

var replicaColors = []string{
	"\033[36m",   // cyan
	"\033[33m",   // yellow
	"\033[32m",   // green
	"\033[35m",   // magenta
	"\033[34m",   // blue
	"\033[91m",   // bright red
	"\033[96m",   // bright cyan
	"\033[93m",   // bright yellow
	"\033[1;36m", // bold cyan
	"\033[1;33m", // bold yellow
	"\033[1;32m", // bold green
	"\033[1;35m", // bold magenta
	"\033[1;34m", // bold blue
	"\033[1;91m", // bold bright red
	"\033[1;96m", // bold bright cyan
	"\033[1;93m", // bold bright yellow
}

const (
	colorReset = "\033[0m"
	colorGray  = "\033[90m"
)

var levelColors = map[string]string{
	"info":    "\033[32m", // green
	"log":     "\033[32m", // green
	"notice":  "\033[32m", // green
	"warning": "\033[33m", // orange/yellow
	"warn":    "\033[33m", // orange/yellow
	"error":   "\033[91m", // red
	"fatal":   "\033[91m", // red
	"panic":   "\033[91m", // red
	"debug":   "\033[36m", // cyan
}

type LogsMeta struct {
	Start     string
	End       string
	Truncated bool
	Empty     bool
}

type LogFormatOptions struct {
	Raw        bool
	Decode     bool
	Fields     []string
	AllFields  bool
	CNPGFormat bool
	Timestamps bool
}

// Informational messages are written to stderr so they don't pollute
// log output when piped (e.g. iai services logs httpbin | grep error).
func PrintLogStream(
	out io.Writer,
	r io.Reader,
	showReplica bool,
	meta LogsMeta,
	opts LogFormatOptions,
) error {
	if meta.Empty {
		PrintNoLogsFound(os.Stderr, meta.Start, meta.End)
		return nil
	}

	if meta.Start != "" {
		fmt.Fprintf(os.Stderr, "Showing logs since %s\n\n", LocalTime(meta.Start))
	}

	useColor := isTerminal(out)
	colorMap := make(map[string]string)
	nextColor := 0
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		if opts.Raw {
			rawLine := line
			if opts.Decode {
				rawLine = decodeLineField(line)
			}
			fmt.Fprintln(out, string(rawLine))
			continue
		}

		var entry logEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			fmt.Fprintln(out, string(line))
			continue
		}

		mainLine, extras := formatLogLine(entry.Line, useColor, opts)
		timestampPart := ""
		if opts.Timestamps && entry.Timestamp != "" {
			timestampPart = formatLogTimestamp(entry.Timestamp) + " "
		}
		replicaPart := replicaPrefix(showReplica, entry.Replica, useColor, colorMap, &nextColor)
		prefix := timestampPart + replicaPart
		if extras != "" {
			fmt.Fprintf(out, "%s%s  %s\n", prefix, mainLine, extras)
		} else {
			fmt.Fprintf(out, "%s%s\n", prefix, mainLine)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if meta.Truncated {
		printLogTruncationWarning(os.Stderr)
	}

	return nil
}

var standardFields = map[string]bool{
	"level": true, "msg": true, "message": true,
	"timestamp": true, "ts": true, "time": true, "t": true,
}

var cnpgStandardFields = map[string]bool{
	"record": true, "msg": true,
	"message": true, "error_severity": true,
}

func formatLogLine(
	line string,
	useColor bool,
	opts LogFormatOptions,
) (string, string) {
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil {
		return line, ""
	}

	level := extractString(fields, "level")
	msg := extractString(fields, "msg")
	if msg == "" {
		msg = extractString(fields, "message")
	}

	excluded := standardFields
	if opts.CNPGFormat {
		if rec, ok := fields["record"].(map[string]any); ok {
			if m := extractString(rec, "message"); m != "" {
				msg = m
			}
			if l := extractString(rec, "error_severity"); l != "" {
				level = l
			}
			excluded = cnpgStandardFields
		}
	}

	if level == "" && msg == "" {
		return line, ""
	}

	var extraKeys []string
	for k := range fields {
		if !excluded[k] && !standardFields[k] {
			extraKeys = append(extraKeys, k)
		}
	}
	sort.Strings(extraKeys)

	var b strings.Builder
	if level != "" {
		b.WriteString(formatLevel(level, useColor))
	}
	b.WriteString(msg)

	showFields := opts.Fields
	if opts.AllFields && len(showFields) == 0 {
		showFields = extraKeys
	}
	extras := formatExtras(fields, showFields, useColor)

	return b.String(), extras
}

func replicaPrefix(
	show bool,
	replica string,
	useColor bool,
	colorMap map[string]string,
	nextColor *int,
) string {
	if !show || replica == "" {
		return ""
	}
	displayReplica := trimReplicaSuffix(replica)
	if !useColor {
		return fmt.Sprintf("[%s] ", displayReplica)
	}
	c, ok := colorMap[replica]
	if !ok {
		c = replicaColors[*nextColor%len(replicaColors)]
		colorMap[replica] = c
		*nextColor++
	}
	return fmt.Sprintf("%s[%s]%s ", c, displayReplica, colorReset)
}

func trimReplicaSuffix(replica string) string {
	if i := strings.LastIndex(replica, "-"); i >= 0 && i+1 < len(replica) {
		return replica[i+1:]
	}
	return replica
}

func formatLevel(level string, useColor bool) string {
	tag := strings.ToUpper(level)
	if !useColor {
		return fmt.Sprintf("%-5s ", tag)
	}
	c := levelColors[strings.ToLower(level)]
	if c == "" {
		c = colorReset
	}
	return fmt.Sprintf("%s%-5s%s ", c, tag, colorReset)
}

func formatExtras(fields map[string]any, keys []string, useColor bool) string {
	if len(keys) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		v, ok := fields[k]
		if !ok {
			continue
		}
		encoded, _ := json.Marshal(v)
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, string(encoded)))
	}
	if len(pairs) == 0 {
		return ""
	}
	extrasLine := strings.Join(pairs, " ")
	if useColor {
		return colorGray + extrasLine + colorReset
	}
	return extrasLine
}

func decodeLineField(line []byte) []byte {
	var raw map[string]json.RawMessage
	if json.Unmarshal(line, &raw) != nil {
		return line
	}
	var s string
	if json.Unmarshal(raw["line"], &s) != nil {
		return line
	}
	var inner json.RawMessage
	if json.Unmarshal([]byte(s), &inner) != nil {
		return line
	}
	raw["line"] = inner
	encoded, err := json.Marshal(raw)
	if err != nil {
		return line
	}
	return encoded
}

func formatLogTimestamp(timestamp string) string {
	if t, ok := parseLogUnixTimestamp(timestamp); ok {
		return t.Local().Format("2006-01-02 15:04:05 MST")
	}
	return LocalTime(timestamp)
}

func parseLogUnixTimestamp(timestamp string) (time.Time, bool) {
	if i, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
		if i >= 1e15 || i <= -1e15 {
			return time.Unix(0, i), true
		}
		return time.Unix(i, 0), true
	}

	f, err := strconv.ParseFloat(timestamp, 64)
	if err != nil {
		return time.Time{}, false
	}
	sec, frac := math.Modf(f)
	return time.Unix(int64(sec), int64(frac*1e9)), true
}

func extractString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

type LogField struct {
	Name  string
	Count int
}

func DiscoverLogFields(r io.Reader) ([]LogField, error) {
	counts := make(map[string]int)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry logEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		var inner map[string]any
		if err := json.Unmarshal([]byte(entry.Line), &inner); err != nil {
			continue
		}

		for k := range inner {
			if standardFields[k] {
				continue
			}
			counts[k]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	fields := make([]LogField, 0, len(counts))
	for name, count := range counts {
		fields = append(fields, LogField{Name: name, Count: count})
	}
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].Count != fields[j].Count {
			return fields[i].Count > fields[j].Count
		}
		return fields[i].Name < fields[j].Name
	})
	return fields, nil
}

func PrintLogFields(out io.Writer, fields []LogField) error {
	if len(fields) == 0 {
		fmt.Fprintln(out, "No additional fields found in recent logs.")
		return nil
	}

	headers := []string{"FIELD", "COUNT"}
	rows := make([][]string, len(fields))
	for i, f := range fields {
		rows[i] = []string{f.Name, fmt.Sprintf("%d", f.Count)}
	}
	return PrintTable(out, headers, rows)
}

func PrintNoLogsFound(errOut io.Writer, start, end string) {
	switch {
	case start != "" && end != "":
		fmt.Fprintf(errOut, "No logs found from %s to %s\n", LocalTime(start), LocalTime(end))
	case start != "":
		fmt.Fprintf(errOut, "No logs found since %s\n", LocalTime(start))
	default:
		fmt.Fprintln(errOut, "No logs found in the given timerange")
	}
}

// printLogTruncationWarning warns that the server truncated the log stream.
func printLogTruncationWarning(errOut io.Writer) {
	printWarning(
		errOut,
		"Warning: output was truncated by the server (max 5000 lines). Use --since or --start-time/--end-time to narrow the time range.",
		true,
	)
}

// PrintLogFieldDiscoveryTruncationWarning warns that field discovery may be incomplete.
func PrintLogFieldDiscoveryTruncationWarning(errOut io.Writer) {
	printWarning(
		errOut,
		"Warning: field discovery may be incomplete because the server truncated the log response (max 5000 lines). Use --since to scan a narrower time range.",
		false,
	)
}

func printWarning(out io.Writer, msg string, leadingNewline bool) {
	prefix := ""
	if leadingNewline {
		prefix = "\n"
	}
	if isTerminal(out) {
		fmt.Fprintf(out, "%s\033[91m%s%s\n", prefix, msg, colorReset)
		return
	}
	fmt.Fprintf(out, "%s%s\n", prefix, msg)
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
