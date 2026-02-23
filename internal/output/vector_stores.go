package output

import (
	"fmt"
	"io"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintVectorStoreList(out io.Writer, stores []clients.VectorStoreInfo) error {
	if len(stores) == 0 {
		fmt.Fprintln(out, "No vector stores found.")
		return nil
	}

	headers := []string{"NAME", "STATUS", "SECRET"}
	rows := make([][]string, len(stores))
	for i, s := range stores {
		rows[i] = []string{
			s.VectorStoreName,
			s.Status,
			s.SecretName,
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintVectorStoreAccepted(out io.Writer, message string) {
	if message != "" {
		fmt.Fprintln(out, message)
	}
}
