package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	obsTraceID   string
	obsIncludeIO bool
	obsColumns   []string
	obsListJSON  bool
	obsGetJSON   bool
)

var observationsCmd = &cobra.Command{
	Use:     "observations",
	Aliases: []string{"obs", "observation"},
	Short:   "Manage observations",
	Long:    `Manage observations within traces. Works with API key (--api-key or INTERACTIVE_API_KEY) or session from 'iai login'.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if root := cmd.Root(); root != nil && root.PersistentPreRun != nil {
			root.PersistentPreRun(cmd, args)
		}
	},
}

var obsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List observations for a trace",
	Long: `List observations for a trace, showing individual LLM calls, spans, and events.

Uses the platform API with dual authentication (API key or session).

Examples:
  iai observations list --trace-id abc123
  iai observations list --trace-id abc123 --include-io
  iai observations list --trace-id abc123 --columns id,type,name,model,latency_ms
  iai observations list --trace-id abc123 --json | jq '.data.observations[].name'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var org, project string
		if f := cmd.Flags().Lookup("organization"); f != nil {
			org = f.Value.String()
		}
		if f := cmd.Flags().Lookup("project"); f != nil {
			project = f.Value.String()
		}

		traceID := strings.TrimSpace(obsTraceID)
		if traceID == "" {
			return fmt.Errorf("--trace-id is required")
		}
		if err := inputs.ValidateTraceID(traceID); err != nil {
			return err
		}

		columns := obsColumns
		if len(columns) == 0 {
			columns = inputs.DefaultObservationColumns
		}
		if !obsListJSON {
			if err := inputs.ValidateObservationColumns(columns); err != nil {
				return err
			}
		}

		pCtx, err := resolveProject(cmd.Context(), org, project)
		if err != nil {
			return err
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		observations, rawJSON, err := apiClient.ListObservations(
			cmd.Context(), pCtx.orgId, pCtx.projectId, traceID, obsIncludeIO,
		)
		if err != nil {
			return err
		}

		if obsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintObservationList(out, observations, columns)
	},
}

var obsGetCmd = &cobra.Command{
	Use:   "get <observation-id>",
	Short: "Get a specific observation",
	Long: `Get detailed information about a specific observation.

Uses the platform API with dual authentication (API key or session).

Examples:
  iai observations get obs-abc123
  iai observations get obs-abc123 --json | jq '.data.observation'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var org, project string
		if f := cmd.Flags().Lookup("organization"); f != nil {
			org = f.Value.String()
		}
		if f := cmd.Flags().Lookup("project"); f != nil {
			project = f.Value.String()
		}

		observationID := strings.TrimSpace(args[0])
		if err := inputs.ValidateObservationID(observationID); err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), org, project)
		if err != nil {
			return err
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		obs, rawJSON, err := apiClient.GetObservation(
			cmd.Context(), pCtx.orgId, pCtx.projectId, observationID,
		)
		if err != nil {
			return err
		}

		if obsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintObservationDetail(out, obs)
	},
}

func init() {
	// observations list flags
	obsListCmd.Flags().StringVar(&obsTraceID, "trace-id", "", "Trace ID to list observations for (required)")
	obsListCmd.Flags().BoolVar(&obsIncludeIO, "include-io", false, "Include input/output/metadata in response")
	obsListCmd.Flags().BoolVar(&obsListJSON, "json", false, "Output raw API response as JSON")
	obsListCmd.Flags().
		StringSliceVar(&obsColumns, "columns", nil, "Columns to display (comma-separated, default: id,type,name,model,latency_ms,total_cost,total_tokens)\nAvailable: id,trace_id,type,name,start_time,end_time,parent_observation_id,level,status_message,model,input_tokens,output_tokens,total_tokens,total_cost,latency_ms")
	obsListCmd.Flags().StringP("organization", "o", "", "Organization name that owns the project")
	obsListCmd.Flags().StringP("project", "p", "", "Project name")
	_ = obsListCmd.MarkFlagRequired("trace-id")

	// observations get flags
	obsGetCmd.Flags().BoolVar(&obsGetJSON, "json", false, "Output raw API response as JSON")
	obsGetCmd.Flags().StringP("organization", "o", "", "Organization name that owns the project")
	obsGetCmd.Flags().StringP("project", "p", "", "Project name")

	observationsCmd.AddCommand(obsListCmd, obsGetCmd)
	rootCmd.AddCommand(observationsCmd)
}
