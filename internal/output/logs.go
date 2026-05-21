package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/term"
)

type logEntry struct {
	Replica string `json:"replica,omitempty"`
	Line    string `json:"line"`
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
	Since     string
	Truncated bool
	Empty     bool
}

type LogFormatOptions struct {
	Raw       bool
	Fields    []string
	AllFields bool
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
		fmt.Fprintln(os.Stderr, "No logs found")
		return nil
	}

	if meta.Since != "" {
		fmt.Fprintf(os.Stderr, "Showing logs since %s\n\n", LocalTime(meta.Since))
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
			fmt.Fprintln(out, string(line))
			continue
		}

		var entry logEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			fmt.Fprintln(out, string(line))
			continue
		}

		prefix := ""
		if showReplica && entry.Replica != "" {
			if useColor {
				c, ok := colorMap[entry.Replica]
				if !ok {
					c = replicaColors[nextColor%len(replicaColors)]
					colorMap[entry.Replica] = c
					nextColor++
				}
				prefix = fmt.Sprintf("%s[%s]%s ", c, entry.Replica, colorReset)
			} else {
				prefix = fmt.Sprintf("[%s] ", entry.Replica)
			}
		}

		mainLine, extras, _ := formatLogLine(entry.Line, useColor, opts.Fields, opts.AllFields)
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
		msg := "Warning: output was truncated by the server (max 5000 lines). Use --since or --start-time/--end-time to narrow the time range."
		if isTerminal(os.Stderr) {
			fmt.Fprintf(os.Stderr, "\n\033[91m%s%s\n", msg, colorReset)
		} else {
			fmt.Fprintf(os.Stderr, "\n%s\n", msg)
		}
	}

	return nil
}

var standardFields = map[string]bool{
	"level": true, "msg": true, "message": true,
	"timestamp": true, "ts": true, "time": true, "t": true,
	"record": true,
}

// formatLogLine returns the main display line, an optional extras string, and
// the list of extra (non-standard) field keys found in this line.
func formatLogLine(
	line string,
	useColor bool,
	fieldNames []string,
	allFields bool,
) (string, string, []string) {
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil {
		return line, "", nil
	}

	level := extractString(fields, "level")
	msg := extractString(fields, "msg")
	if msg == "" {
		msg = extractString(fields, "message")
	}

	if rec, ok := fields["record"].(map[string]any); ok {
		if recMsg := extractString(rec, "message"); recMsg != "" {
			msg = recMsg
		}
		if recLevel := extractString(rec, "error_severity"); recLevel != "" {
			level = recLevel
		}
	}

	if level == "" && msg == "" {
		return line, "", nil
	}

	var extraKeys []string
	for k := range fields {
		if !standardFields[k] {
			extraKeys = append(extraKeys, k)
		}
	}
	sort.Strings(extraKeys)

	var b strings.Builder

	if level != "" {
		tag := strings.ToUpper(level)
		if useColor {
			c := levelColors[strings.ToLower(level)]
			if c == "" {
				c = colorReset
			}
			fmt.Fprintf(&b, "%s%-5s%s ", c, tag, colorReset)
		} else {
			fmt.Fprintf(&b, "%-5s ", tag)
		}
	}

	b.WriteString(msg)

	showFields := fieldNames
	if allFields && len(showFields) == 0 {
		showFields = extraKeys
	}

	var extras string
	if len(showFields) > 0 {
		pairs := make([]string, 0, len(showFields))
		for _, k := range showFields {
			v, ok := fields[k]
			if !ok {
				continue
			}
			encoded, _ := json.Marshal(v)
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, string(encoded)))
		}
		if len(pairs) > 0 {
			extrasLine := strings.Join(pairs, " ")
			if useColor {
				extras = fmt.Sprintf("%s%s%s", colorGray, extrasLine, colorReset)
			} else {
				extras = extrasLine
			}
		}
	}

	return b.String(), extras, extraKeys
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

func DiscoverLogFields(r io.Reader) []LogField {
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
			if !standardFields[k] {
				counts[k]++
			}
		}
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
	return fields
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

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
