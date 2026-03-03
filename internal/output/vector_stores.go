package output

import (
	"fmt"
	"io"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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

	fmt.Fprintf(out, "Name:            %s\n", store.VectorStoreName)
	fmt.Fprintf(out, "Status:          %s\n", store.Status)
	fmt.Fprintf(out, "Engine Version:  %s\n", store.EngineVersion)
	if store.CreatedAt != "" {
		fmt.Fprintf(out, "Created At:      %s\n", LocalTime(store.CreatedAt))
	}
	fmt.Fprintf(out, "HA:              %s\n", haStr)
	backupsStr := "No"
	if store.Backups.Enabled {
		backupsStr = "Yes"
	}
	fmt.Fprintf(out, "Backups:         %s\n", backupsStr)
	if store.Backups.Enabled && store.Backups.StartTime != "" {
		fmt.Fprintf(out, "Backup Time:     %s\n", store.Backups.StartTime)
	}
	if store.SecretName != "" {
		fmt.Fprintf(out, "Secret:          %s\n", store.SecretName)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Resources:")
	fmt.Fprintf(out, "  CPU:               %d\n", store.Resources.CPU)
	fmt.Fprintf(out, "  Memory:            %.2f GB\n", store.Resources.Memory)

	fmt.Fprintln(out, "Storage:")
	fmt.Fprintf(out, "  Size:              %d GB\n", store.Storage.Size)
	fmt.Fprintf(out, "  Auto Resize:       %s\n", autoResizeStr)
	if store.Storage.AutoResize {
		fmt.Fprintf(out, "  Auto Resize Limit: %d GB\n", store.Storage.AutoResizeLimit)
	}

	return nil
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
