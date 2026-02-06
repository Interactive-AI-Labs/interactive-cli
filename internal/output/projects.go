package output

import (
	"fmt"
	"io"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintProjectList(out io.Writer, projects []clients.Project, selectedProject string) error {
	fmt.Fprintln(out)

	headers := []string{"NAME", "ROLE"}
	rows := make([][]string, len(projects))
	for i, proj := range projects {
		displayName := proj.Name
		if selectedProject != "" && strings.EqualFold(proj.Name, selectedProject) {
			displayName = displayName + " *"
		}
		rows[i] = []string{displayName, proj.Role}
	}

	return PrintTable(out, headers, rows)
}
