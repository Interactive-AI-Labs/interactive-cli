package output

import (
	"fmt"
	"io"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintOrganizationList(out io.Writer, orgs []clients.Organization, selectedOrg string) error {
	headers := []string{"NAME", "PROJECTS", "ROLE"}
	rows := make([][]string, len(orgs))
	for i, org := range orgs {
		displayName := org.Name
		if selectedOrg != "" && strings.EqualFold(org.Name, selectedOrg) {
			displayName = displayName + " *"
		}
		rows[i] = []string{displayName, fmt.Sprintf("%d", org.ProjectCount), org.Role}
	}

	return PrintTable(out, headers, rows)
}
