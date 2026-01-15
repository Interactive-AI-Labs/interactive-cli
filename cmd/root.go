package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	version            = "0.0.31"
	cfgDirName         = ".interactiveai"
	sessionFileName    = "session_cookies.json"
	defaultHTTPTimeout = 15 * time.Second
)

var (
	hostname           string = "https://app.interactive.ai"
	deploymentHostname string = "https://deployment.interactive.ai"
	apiKey             string
	cfgFilePath        string
	rootCmd            = &cobra.Command{
		Use:     "iai",
		Short:   "InteractiveAI's CLI",
		Version: version,
		Long: `# Install

The CLI is distributed through Go's package manager, so it must first be installed. Click on [this](https://go.dev/doc/install) link and follow the instructions to do so.

To validate the installation run:

` + "```bash" + `
go version
` + "```" + `

Once Go is installed, ensure Go binaries are in your PATH:

` + "```bash" + `
export PATH=$PATH:$(go env GOPATH)/bin
` + "```" + `

Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.) to make it permanent.

Now install InteractiveAI's CLI with the following command:

` + "```bash" + `
go install github.com/Interactive-AI-Labs/interactive-cli/cmd/iai@latest
` + "```" + `

Verify the installation by running:

` + "```bash" + `
iai --help
` + "```" + `

---

InteractiveAI's CLI to interact with its platform`,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !strings.HasPrefix(hostname, "http://") && !strings.HasPrefix(hostname, "https://") {
				hostname = "https://" + hostname
			}
			if !strings.HasPrefix(deploymentHostname, "http://") && !strings.HasPrefix(deploymentHostname, "https://") {
				deploymentHostname = "https://" + deploymentHostname
			}
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
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

	envApiKey := os.Getenv("INTERACTIVE_API_KEY")
	if envApiKey != "" {
		apiKey = envApiKey
	}

	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", hostname, "Hostname for the API")
	rootCmd.PersistentFlags().StringVar(&deploymentHostname, "deployment-hostname", deploymentHostname, "Hostname for the deployment API")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", apiKey, "API key for authentication")
	rootCmd.PersistentFlags().StringVar(&cfgFilePath, "cfg-file", "", "Path to YAML config file with organization, project, and optional service definitions")
}
