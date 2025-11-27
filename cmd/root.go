package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	cfgDirName         = ".interactiveai"
	sessionFileName    = "session_cookies.json"
	defaultHTTPTimeout = 15 * time.Second
)

var (
	hostname string = "https://app.interactive.ai"
	rootCmd         = &cobra.Command{
		Use:   "interactiveai",
		Short: "InteractiveAI's CLI",
		Long:  `InteractiveAI's CLI to interact with its platform`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !strings.HasPrefix(hostname, "http://") && !strings.HasPrefix(hostname, "https://") {
				hostname = "https://" + hostname
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
	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", hostname, "Hostname for the API")
}
