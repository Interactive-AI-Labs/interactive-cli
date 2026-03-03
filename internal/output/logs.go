package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

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

const colorReset = "\033[0m"

type LogsMeta struct {
	Since     string
	Truncated bool
	Empty     bool
}

// Informational messages are written to stderr so they don't pollute
// log output when piped (e.g. iai services logs httpbin | grep error).
func PrintLogStream(out io.Writer, r io.Reader, showReplica bool, meta LogsMeta) error {
	if meta.Empty {
		fmt.Fprintln(os.Stderr, "No logs found, check if replica exists")
		return nil
	}

	if meta.Since != "" {
		fmt.Fprintf(os.Stderr, "Showing logs since %s\n\n", LocalTime(meta.Since))
	}

	useColor := showReplica && isTerminal(out)
	colorMap := make(map[string]string)
	nextColor := 0

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry logEntry
		// Not valid JSON; print the raw line as-is.
		if err := json.Unmarshal(line, &entry); err != nil {
			fmt.Fprintln(out, string(line))
			continue
		}

		// No replica prefix needed; print the log line only.
		if !showReplica || entry.Replica == "" {
			fmt.Fprintln(out, entry.Line)
			continue
		}

		// Plain text replica prefix without color.
		if !useColor {
			fmt.Fprintf(out, "[%s] %s\n", entry.Replica, entry.Line)
			continue
		}

		c, ok := colorMap[entry.Replica]
		if !ok {
			c = replicaColors[nextColor%len(replicaColors)]
			colorMap[entry.Replica] = c
			nextColor++
		}
		fmt.Fprintf(out, "%s[%s]%s %s\n", c, entry.Replica, colorReset, entry.Line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if meta.Truncated {
		msg := "Warning: output was truncated by the server (max 5000 lines). Use --since or --start-time to narrow the time range."
		if isTerminal(os.Stderr) {
			fmt.Fprintf(os.Stderr, "\n\033[91m%s%s\n", msg, colorReset)
		} else {
			fmt.Fprintf(os.Stderr, "\n%s\n", msg)
		}
	}

	return nil
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
