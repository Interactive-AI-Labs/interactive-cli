package output

import (
	"cmp"
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func PrintRevisions(out io.Writer, revisions []deployment.RevisionMeta) error {
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
			cmp.Or(LocalTime(revision.Updated), missingMetadata),
			formatRevisionActor(revision.Actor),
			formatRevisionSource(revision.Source),
		}
	}

	return PrintTable(out, headers, rows)
}

func printRevisionAttribution(out io.Writer, revision deployment.RevisionMeta) {
	fmt.Fprintf(out, "By:\t%s\n", formatRevisionActor(revision.Actor))
	fmt.Fprintf(out, "Source:\t%s\n", formatRevisionSource(revision.Source))
}

func formatRevisionActor(actor *deployment.RevisionActor) string {
	if actor == nil {
		return missingMetadata
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
	case "unknown":
		return "unknown"
	default:
		return name
	}
}

func formatRevisionSource(source *deployment.RevisionSource) string {
	if source == nil {
		return missingMetadata
	}

	sourceType := source.Type
	if sourceType == "" && source.Version == "" {
		return missingMetadata
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
