package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/buildinfo"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/versioncheck"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	version            = buildinfo.Version
	cfgDirName         = ".interactiveai"
	sessionFileName    = "session_cookies.json"
	defaultHTTPTimeout = 30 * time.Second

	groupAuth        = "auth"
	groupAgents      = "agents"
	groupContext     = "context"
	groupConnectors  = "connectors"
	groupInfra       = "infra"
	groupGovernance  = "governance"
	groupImprovement = "improvement"
)

var (
	hostname           string = "https://app.interactive.ai"
	deploymentHostname string = "https://deployment.interactive.ai"
	token              string
	apiKey             string
	clientOrigin       string
	cfgFilePath        string
	rootCmd            = &cobra.Command{
		Use:     "iai",
		Short:   "InteractiveAI's CLI",
		Version: version,
		Long: `InteractiveAI's CLI to interact with its platform.

Use the subcommands below to manage your organizations, projects, agents, services, secrets, prompts, routines, policies, variables, glossaries, macros, and other components.`,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !strings.HasPrefix(hostname, "http://") && !strings.HasPrefix(hostname, "https://") {
				hostname = "https://" + hostname
			}
			if !strings.HasPrefix(deploymentHostname, "http://") &&
				!strings.HasPrefix(deploymentHostname, "https://") {
				deploymentHostname = "https://" + deploymentHostname
			}
			if apiKey != "" && token != "" {
				fmt.Fprintln(
					cmd.ErrOrStderr(),
					"Warning: the API key (--api-key or INTERACTIVE_API_KEY) is ignored because INTERACTIVE_TOKEN is set; unset INTERACTIVE_TOKEN to use it.",
				)
			}

			// RefreshCache is intentionally gated too: no point keeping the
			// cache warm when the notice can't be shown anyway.
			if cmd.Name() != "update" && updateNoticeAllowed() {
				notifyUpdate()
				go versioncheck.RefreshCache(cfgDirName)
			}
		},
	}
)

// updateNoticeAllowed reports whether the upgrade nudge may be shown: never
// in CI, in scripts (stderr not a terminal), or when the user opted out.
func updateNoticeAllowed() bool {
	if os.Getenv("CI") != "" || os.Getenv("INTERACTIVE_NO_UPDATE_NOTIFIER") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// notifyUpdate queues the upgrade nudge from the on-disk version cache, so a
// new release is announced one run after RefreshCache first stores it.
func notifyUpdate() {
	latest, ok := versioncheck.PendingNotification(cfgDirName, version)
	if !ok {
		return
	}
	updateMessage <- fmt.Sprintf(
		"\nA new version of iai is available: v%s → v%s\nRun \"iai update\" to upgrade.",
		version, latest,
	)
}

// chainRootPersistentPreRun calls the root command's PersistentPreRun manually.
// Cobra doesn't chain PersistentPreRun hooks, so subcommands that define their
// own must call this to preserve URL normalization.
var chainRootPersistentPreRun = func(cmd *cobra.Command, args []string) {
	if root := cmd.Root(); root != nil && root.PersistentPreRun != nil {
		root.PersistentPreRun(cmd, args)
	}
}

var updateMessage = make(chan string, 1)

func Execute() {
	refuseBlankArgs(rootCmd)
	err := rootCmd.Execute()

	select {
	case msg := <-updateMessage:
		fmt.Fprintln(os.Stderr, msg)
	default:
	}

	if err != nil {
		os.Exit(1)
	}
}

// refuseBlankArgs stops a blank name, such as `agents delete " "`, from reaching the API.
func refuseBlankArgs(c *cobra.Command) {
	validate := c.Args
	if validate == nil && !c.HasSubCommands() {
		validate = cobra.ArbitraryArgs
	}
	if validate != nil {
		c.Args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if strings.TrimSpace(arg) == "" {
					return fmt.Errorf("arguments cannot be blank")
				}
			}
			return validate(cmd, args)
		}
	}
	for _, sub := range c.Commands() {
		refuseBlankArgs(sub)
	}
}

func init() {
	envHostname := os.Getenv("INTERACTIVE_HOSTNAME")
	if envHostname != "" {
		hostname = envHostname
	}

	envDeploymentHostname := os.Getenv("INTERACTIVE_DEPLOYMENT_HOSTNAME")
	if envDeploymentHostname != "" {
		deploymentHostname = envDeploymentHostname
	}

	envToken := os.Getenv("INTERACTIVE_TOKEN")
	if envToken != "" {
		token = envToken
	}

	envApiKey := os.Getenv("INTERACTIVE_API_KEY")
	if envApiKey != "" {
		apiKey = envApiKey
	}

	// The Copilot sandbox sets INTERACTIVE_CLIENT_ORIGIN=copilot, so report writes record where they came from.
	clientOrigin = os.Getenv("INTERACTIVE_CLIENT_ORIGIN")

	rootCmd.AddGroup(
		&cobra.Group{ID: groupAuth, Title: "Auth:"},
		&cobra.Group{ID: groupAgents, Title: "Agents:"},
		&cobra.Group{ID: groupContext, Title: "Context:"},
		&cobra.Group{ID: groupConnectors, Title: "Connectors:"},
		&cobra.Group{ID: groupInfra, Title: "Infrastructure:"},
		&cobra.Group{ID: groupGovernance, Title: "Governance:"},
		&cobra.Group{ID: groupImprovement, Title: "Improvement:"},
	)

	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", hostname, "Hostname for the API")
	rootCmd.PersistentFlags().
		StringVar(&deploymentHostname, "deployment-hostname", deploymentHostname, "Hostname for the deployment API")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", apiKey, "API key for authentication")
	rootCmd.PersistentFlags().
		StringVar(&cfgFilePath, "cfg-file", "", "Path to YAML config file with organization, project, and optional service definitions")
}
