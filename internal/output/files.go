package output

import (
	"fmt"
	"io"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

var fileColumnMap = map[string]struct {
	Header string
	Value  func(f *platform.FileMetadata) string
}{
	"id":   {"ID", func(f *platform.FileMetadata) string { return f.ID }},
	"name": {"NAME", func(f *platform.FileMetadata) string { return Printable(f.Name) }},
	"size": {"SIZE", func(f *platform.FileMetadata) string { return HumanBytes(f.Current.Size) }},
	"uploaded_by": {
		"UPLOADED BY",
		func(f *platform.FileMetadata) string { return uploadedBy(f.Current) },
	},
	"created_at": {
		"CREATED AT",
		func(f *platform.FileMetadata) string {
			return LocalTime(f.Current.CreatedAt.Format(time.RFC3339))
		},
	},
	"content_type": {
		"CONTENT TYPE",
		func(f *platform.FileMetadata) string { return Printable(f.Current.ContentType) },
	},
	"version_id": {
		"VERSION ID",
		func(f *platform.FileMetadata) string { return f.Current.VersionID },
	},
}

func uploadedBy(v platform.FileVersion) string {
	if v.CreatedByEmail != nil {
		return Printable(*v.CreatedByEmail)
	}
	return Printable(v.CreatedBy)
}

func PrintFileList(
	out io.Writer,
	files []platform.FileMetadata,
	meta platform.FileListMeta,
	columns []string,
	includeVersions bool,
) error {
	if len(files) == 0 {
		fmt.Fprintln(out, "No files found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := fileColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(files))
	for i, f := range files {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := fileColumnMap[col]; ok {
				row[j] = def.Value(&f)
			}
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	if includeVersions {
		for _, f := range files {
			fmt.Fprintf(out, "\n%s versions:\n", Printable(f.Name))
			if err := PrintFileVersions(out, f.Versions); err != nil {
				return err
			}
		}
	}

	if meta.Cursor != nil && *meta.Cursor != "" {
		fmt.Fprintf(out, "\nMore results — next page: --cursor %s\n", Printable(*meta.Cursor))
	}
	return nil
}

func PrintFileVersions(out io.Writer, versions []platform.FileVersion) error {
	headers := []string{"VERSION ID", "NAME", "SIZE", "UPLOADED BY", "CREATED AT", "CURRENT"}
	rows := make([][]string, len(versions))
	for i, v := range versions {
		current := ""
		if v.IsCurrent {
			current = "yes"
		}
		rows[i] = []string{
			v.VersionID, Printable(v.Name), HumanBytes(v.Size), uploadedBy(v),
			LocalTime(v.CreatedAt.Format(time.RFC3339)), current,
		}
	}
	return PrintTable(out, headers, rows)
}

func PrintFileDetail(out io.Writer, f *platform.FileMetadata) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", f.ID)
	fmt.Fprintf(w, "Name:\t%s\n", Printable(f.Name))
	fmt.Fprintf(w, "Current Version:\t%s\n", f.Current.VersionID)
	fmt.Fprintf(w, "Size:\t%s\n", HumanBytes(f.Current.Size))
	fmt.Fprintf(w, "Uploaded By:\t%s\n", uploadedBy(f.Current))
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(f.Current.CreatedAt.Format(time.RFC3339)))
	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Fprintln(out, "\nVersions:")
	return PrintFileVersions(out, f.Versions)
}

func PrintFileRefCandidates(
	out io.Writer,
	ref string,
	candidates []clients.FileRefCandidate,
) error {
	if _, err := fmt.Fprintf(out, "%q matches more than one file:\n", ref); err != nil {
		return err
	}

	headers := []string{"ID", "NAME", "SIZE", "CREATED AT"}
	rows := make([][]string, len(candidates))
	for i, c := range candidates {
		rows[i] = []string{
			c.FileID,
			Printable(c.Name),
			HumanBytes(c.Size),
			LocalTime(c.CreatedAt.Format(time.RFC3339)),
		}
	}
	return PrintTable(out, headers, rows)
}
