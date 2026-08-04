package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

const missingRevisionMetadata = "—"

func PrintRevisions(out io.Writer, revisions []clients.RevisionMeta) error {
	if len(revisions) == 0 {
		fmt.Fprintln(out, "No revisions found.")
		return nil
	}

	latest := 0
	for _, revision := range revisions {
		if revision.Revision > latest {
			latest = revision.Revision
		}
	}

	headers := []string{"", "REVISION", "UPDATED", "BY", "SOURCE"}
	rows := make([][]string, len(revisions))
	for i, revision := range revisions {
		marker := ""
		if revision.Revision == latest {
			marker = "*"
		}
		rows[i] = []string{
			marker,
			fmt.Sprintf("%d", revision.Revision),
			LocalTime(revision.Updated),
			formatRevisionActor(revision.Actor),
			formatRevisionSource(revision.Source),
		}
	}

	return PrintTable(out, headers, rows)
}

func printRevisionAttribution(out io.Writer, revision clients.RevisionMeta) {
	fmt.Fprintf(out, "By:\t%s\n", formatRevisionActor(revision.Actor))
	fmt.Fprintf(out, "Source:\t%s\n", formatRevisionSource(revision.Source))
}

func formatRevisionActor(actor *clients.RevisionActor) string {
	if actor == nil {
		return missingRevisionMetadata
	}

	name := actor.DisplayName
	if name == "" {
		name = actor.ID
	}
	if actor.Type == "system" && (name == "" || name == "system") {
		return "system"
	}
	if name == "" {
		return "unknown"
	}

	switch actor.Type {
	case "user":
		return name
	case "api_key":
		return fmt.Sprintf("%s (API key)", name)
	case "service":
		return fmt.Sprintf("%s (service)", name)
	case "system":
		return fmt.Sprintf("%s (system)", name)
	default:
		return "unknown"
	}
}

func formatRevisionSource(source *clients.RevisionSource) string {
	if source == nil {
		return missingRevisionMetadata
	}

	sourceType := source.Type
	if sourceType == "" && source.Version == "" {
		return missingRevisionMetadata
	}
	if sourceType == "cli" {
		sourceType = "iai"
	}
	if sourceType == "" {
		sourceType = "unknown"
	}
	if source.Version == "" {
		return sourceType
	}
	return sourceType + " " + source.Version
}
