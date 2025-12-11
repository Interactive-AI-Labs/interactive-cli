package files

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseEnvFile reads a file from the given path and parses it as KEY=VALUE pairs.
// Empty lines and lines starting with # (comments) are skipped.
func ParseEnvFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	result := make(map[string]string)
	var errors []string
	lineNum := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			errors = append(errors, fmt.Sprintf("  line %d: missing '=' separator", lineNum))
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			errors = append(errors, fmt.Sprintf("  line %d: empty key", lineNum))
			continue
		}

		value := strings.TrimSpace(parts[1])
		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to parse env file: found %d malformed lines:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return result, nil
}
