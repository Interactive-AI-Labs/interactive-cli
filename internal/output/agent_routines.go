package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func PrintAgentRoutineRun(out io.Writer, run *deployment.AgentRoutineRun) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Run:\t%s\nRoutine:\t%s\nStatus:\t%s\nSession:\t%s\nCreated:\t%s\n",
		run.RunID, run.RoutineID, run.Status, run.SessionID, run.CreatedAt)
	return w.Flush()
}

// RoutineFollowCommand quotes shell arguments and includes only nonempty hostname overrides.
func RoutineFollowCommand(
	agentName, sessionID, org, project, hostname, deploymentHostname string,
) string {
	var args []string
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	if deploymentHostname != "" {
		args = append(args, "--deployment-hostname", deploymentHostname)
	}
	args = append(args, "agents", "events", "list", agentName, sessionID,
		"--organization", org, "--project", project, "--follow")
	return "iai " + shellJoin(args)
}
