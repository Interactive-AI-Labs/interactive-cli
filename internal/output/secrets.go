package output

import (
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintSecretList(out io.Writer, secrets []clients.SecretInfo) error {
	if len(secrets) == 0 {
		fmt.Fprintln(out, "No secrets found.")
		return nil
	}

	headers := []string{"NAME", "TYPE", "CREATED", "KEYS"}
	rows := make([][]string, len(secrets))
	for i, s := range secrets {
		rows[i] = []string{
			s.Name,
			s.Type,
			s.CreatedAt,
			formatSecretKeys(s.Keys, 3),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintSecretData(out io.Writer, data map[string]string) error {
	if len(data) == 0 {
		fmt.Fprintln(out, "No data found in secret.")
		return nil
	}

	headers := []string{"KEYS", "VALUES"}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := make([][]string, 0, len(data))
	for _, k := range keys {
		val := data[k]
		if decoded, err := base64.StdEncoding.DecodeString(val); err == nil {
			val = string(decoded)
		}
		rows = append(rows, []string{k, val})
	}

	return PrintTable(out, headers, rows)
}

func formatSecretKeys(keys []string, maxVisible int) string {
	if len(keys) == 0 {
		return ""
	}
	if len(keys) <= maxVisible {
		return strings.Join(keys, ", ")
	}
	visible := strings.Join(keys[:maxVisible], ", ")
	return fmt.Sprintf("%s (+%d more)", visible, len(keys)-maxVisible)
}
