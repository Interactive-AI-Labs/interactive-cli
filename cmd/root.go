package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	cfgDirName         = ".interactiveai"
	sessionFileName    = "session_cookies.json"
	defaultHTTPTimeout = 15 * time.Second
)

var rootCmd = &cobra.Command{
	Use:   "interactiveai",
	Short: "InteractiveAI's CLI",
	Long:  `InteractiveAI's CLI to interact with its platform`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
