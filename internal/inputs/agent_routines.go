package inputs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

func ValidateRoutineRun(agentName, routineID, callbackURL string) error {
	if strings.TrimSpace(agentName) == "" {
		return fmt.Errorf("agent name must not be empty")
	}
	for _, segment := range strings.Split(routineID, "/") {
		if strings.TrimSpace(segment) == "" || segment == "." || segment == ".." ||
			strings.Contains(segment, `\`) {
			return fmt.Errorf(
				"routine ID must contain nonempty path segments without '.' or '..' segments or backslashes",
			)
		}
	}
	u, err := url.Parse(callbackURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return fmt.Errorf("--callback-url must be an HTTP(S) URL with a host")
	}
	return nil
}

// ReadRoutineInput preserves the input JSON, including exact numbers.
func ReadRoutineInput(path string, stdin io.Reader) (json.RawMessage, error) {
	if path == "" {
		return nil, fmt.Errorf("--input-file must not be empty")
	}
	var raw []byte
	var err error
	if path == "-" {
		raw, err = io.ReadAll(stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read routine input: %w", err)
	}
	var input map[string]json.RawMessage
	if err := json.Unmarshal(raw, &input); err != nil || input == nil {
		return nil, fmt.Errorf("routine input must be a JSON object")
	}
	return raw, nil
}
