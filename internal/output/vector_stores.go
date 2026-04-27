package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintVectorStoreDescribe(out io.Writer, store *clients.DescribeVectorStoreResponse) error {
	haStr := "No"
	if store.HA {
		haStr = "Yes"
	}

	autoResizeStr := "No"
	if store.Storage.AutoResize {
		autoResizeStr = "Yes"
	}

	backupsStr := "No"
	if store.Backups.Enabled {
		backupsStr = "Yes"
	}

	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", store.VectorStoreName)
	fmt.Fprintf(w, "Status:\t%s\n", store.Status)
	fmt.Fprintf(w, "Engine Version:\t%s\n", store.EngineVersion)
	if store.CreatedAt != "" {
		fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(store.CreatedAt))
	}
	fmt.Fprintf(w, "HA:\t%s\n", haStr)
	fmt.Fprintf(w, "Backups:\t%s\n", backupsStr)
	if store.Backups.Enabled && store.Backups.StartTime != "" {
		fmt.Fprintf(w, "Backup Time:\t%s\n", store.Backups.StartTime)
	}
	if store.SecretName != "" {
		fmt.Fprintf(w, "Secret:\t%s\n", store.SecretName)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Resources:")
	fmt.Fprintf(w, "  CPU:\t%d\n", store.Resources.CPU)
	fmt.Fprintf(w, "  Memory:\t%.2f GB\n", store.Resources.Memory)

	fmt.Fprintln(w, "Storage:")
	fmt.Fprintf(w, "  Size:\t%d GB\n", store.Storage.Size)
	fmt.Fprintf(w, "  Auto Resize:\t%s\n", autoResizeStr)
	if store.Storage.AutoResize {
		fmt.Fprintf(w, "  Auto Resize Limit:\t%d GB\n", store.Storage.AutoResizeLimit)
	}

	return w.Flush()
}

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
