package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintDatabaseList(out io.Writer, databases []clients.DatabaseOutput) error {
	if len(databases) == 0 {
		fmt.Fprintln(out, "No databases found.")
		return nil
	}

	headers := []string{"NAME", "REVISION", "STATUS", "UPDATED"}
	rows := make([][]string, len(databases))
	for i, db := range databases {
		rows[i] = []string{
			db.Name,
			fmt.Sprintf("%d", db.Revision),
			db.Status,
			LocalTime(db.Updated),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintDatabaseDescribe(out io.Writer, db *clients.DescribeDatabaseResponse) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", db.Name)
	fmt.Fprintf(w, "Revision:\t%d\n", db.Revision)
	fmt.Fprintf(w, "Status:\t%s\n", db.Status)
	if db.Message != "" {
		fmt.Fprintf(w, "Message:\t%s\n", db.Message)
	}
	if db.Updated != "" {
		fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(db.Updated))
	}
	fmt.Fprintf(w, "PostgreSQL Version:\t%s\n", db.PostgresVersion)
	fmt.Fprintf(w, "Instances:\t%d\n", db.Instances)

	fmt.Fprintln(w, "Resources:")
	fmt.Fprintf(w, "  CPU:\t%s\n", db.Resources.CPU)
	fmt.Fprintf(w, "  Memory:\t%s\n", db.Resources.Memory)

	fmt.Fprintln(w, "Storage:")
	fmt.Fprintf(w, "  Size:\t%s\n", db.Storage.Size)

	if len(db.Extensions) > 0 {
		fmt.Fprintf(w, "Extensions:\t%s\n", strings.Join(db.Extensions, ", "))
	}

	if db.Backup != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Backup:")
		fmt.Fprintf(w, "  Schedule:\t%s\n", db.Backup.Schedule)
		fmt.Fprintf(w, "  Retention:\t%s\n", db.Backup.RetentionPolicy)
	}

	if db.StackId != "" {
		fmt.Fprintf(w, "Stack:\t%s\n", db.StackId)
	}

	if db.CredentialsSecret != "" {
		fmt.Fprintf(w, "Credentials Secret:\t%s\n", db.CredentialsSecret)
	}

	return w.Flush()
}

func PrintDatabaseBackups(out io.Writer, backups []clients.BackupOutput) error {
	if len(backups) == 0 {
		fmt.Fprintln(out, "No backups found.")
		return nil
	}

	headers := []string{"NAME", "PHASE", "STARTED", "STOPPED", "ERROR"}
	rows := make([][]string, len(backups))
	for i, b := range backups {
		rows[i] = []string{
			b.Name,
			b.Phase,
			LocalTime(b.StartedAt),
			LocalTime(b.StoppedAt),
			b.Error,
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintDatabaseBackup(out io.Writer, backup *clients.BackupOutput) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", backup.Name)
	fmt.Fprintf(w, "Phase:\t%s\n", backup.Phase)
	if backup.StartedAt != "" {
		fmt.Fprintf(w, "Started At:\t%s\n", LocalTime(backup.StartedAt))
	}
	if backup.StoppedAt != "" {
		fmt.Fprintf(w, "Stopped At:\t%s\n", LocalTime(backup.StoppedAt))
	}
	if backup.Error != "" {
		fmt.Fprintf(w, "Error:\t%s\n", backup.Error)
	}

	return w.Flush()
}
