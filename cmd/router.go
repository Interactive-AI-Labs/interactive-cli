package cmd

import (
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	routerInfoJSON bool
	routerInfoYAML bool
)

var routerCmd = &cobra.Command{
	Use:     "router",
	Short:   "Inspect the inference router, keys, and models",
	GroupID: groupInfra,
	Long: `Inspect the InteractiveAI inference router, manage its API keys, and browse available models.

Use "iai router info" to display inference endpoint URLs.`,
}

var routerInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display router endpoint information",
	Long:  `Display the inference router base URL, endpoints, and documentation URL.`,
	Example: `  iai router info
  iai router info --json
  iai router info --yaml
  iai router info --hostname https://dev.interactive.ai`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := output.NewRouterInfo(hostname)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if routerInfoJSON {
			return output.PrintStructuredJSON(out, info)
		}
		if routerInfoYAML {
			return output.PrintStructuredYAML(out, info)
		}
		return output.PrintRouterInfo(out, info)
	},
}

func init() {
	routerInfoCmd.Flags().
		BoolVar(&routerInfoJSON, "json", false, "Output router information as JSON")
	routerInfoCmd.Flags().
		BoolVar(&routerInfoYAML, "yaml", false, "Output router information as YAML")
	routerInfoCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	routerCmd.AddCommand(routerInfoCmd)
	rootCmd.AddCommand(routerCmd)
}
