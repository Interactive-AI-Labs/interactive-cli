package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	reportsOrganization string
	reportsProject      string

	reportsListFolder  string
	reportsListAgent   string
	reportsListColumns []string
	reportsListJSON    bool
	reportsListYAML    bool

	reportsGetJSON bool
	reportsGetYAML bool

	reportsCreateFile   string
	reportsCreateDryRun bool

	reportsUpdateFile   string
	reportsUpdateForce  bool
	reportsUpdateDryRun bool

	reportsRestoreExpectedUpdatedAt string

	reportsDeleteForce bool

	reportsSchemaJSON bool
	reportsSchemaYAML bool
)

const reportWrites = "report changes"

func reportPath(projectID, reportID string) string {
	return fmt.Sprintf("/project/%s/dashboards/%s", projectID, reportID)
}

// findReport reads the report named ref, ignoring case, or with byID the report with id ref.
func findReport(
	ctx context.Context,
	client *platform.APIClient,
	projectID, ref string,
	byID bool,
) (platform.Report, json.RawMessage, error) {
	id, err := findReportID(ctx, client, projectID, ref, byID)
	if err != nil {
		return platform.Report{}, nil, err
	}
	return client.GetReport(ctx, projectID, id)
}

func findReportID(
	ctx context.Context,
	client *platform.APIClient,
	projectID, ref string,
	byID bool,
) (string, error) {
	reports, _, err := client.ListReports(ctx, projectID, platform.ReportListOptions{})
	if err != nil {
		return "", err
	}
	found, err := inputs.ResolveReport(reports, strings.TrimSpace(ref), byID)
	return found.ID, err
}

// reportFolder is the name of the folder a report is in; empty for none.
func reportFolder(
	ctx context.Context,
	client *platform.APIClient,
	projectID string,
	folderID *string,
) (string, error) {
	if folderID == nil {
		return "", nil
	}
	folders, err := client.ListReportFolders(ctx, projectID)
	if err != nil {
		return "", err
	}
	return inputs.ReportFolderName(folderID, inputs.ReportFolderNames(folders)), nil
}

var reportsCmd = &cobra.Command{
	Use:     "reports",
	Aliases: []string{"report"},
	Short:   "Build and change project reports",
	GroupID: groupObserve,
	Long: `Read and change a project's reports. A report is a page of charts, tables and numbers
over the project's traces, usage and costs, laid out in sections of widgets.

A report is saved whole, as one document: print it with 'iai reports get <name> --yaml',
edit it, and save it with 'iai reports update <name> --file', or 'iai reports create --file'
for a new one. Changing a report requires iai login or JWT authentication.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var reportsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List a project's reports",
	Long:    "List a project's reports with their folder, agents and last change.",
	Example: `  iai reports list
  iai reports list --folder Operations
  iai reports list --agent <agent-id> --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		columns := reportsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultReportColumns
		}
		if err := validateTableOnlyColumns(cmd, reportsListJSON, reportsListYAML); err != nil {
			return err
		}
		if !reportsListJSON && !reportsListYAML {
			if err := inputs.ValidateColumns(columns, inputs.AllReportColumns); err != nil {
				return err
			}
		}

		pCtx, client, _, err := resolveProject(cmd.Context(), reportsOrganization, reportsProject)
		if err != nil {
			return err
		}
		raw := reportsListJSON || reportsListYAML
		var folders []platform.ReportFolder
		if reportsListFolder != "" || !raw {
			if folders, err = client.ListReportFolders(cmd.Context(), pCtx.projectId); err != nil {
				return err
			}
		}
		folderID := ""
		if reportsListFolder != "" {
			if folderID, err = inputs.ResolveReportFolder(folders, reportsListFolder); err != nil {
				return err
			}
		}
		reports, body, err := client.ListReports(
			cmd.Context(), pCtx.projectId,
			platform.ReportListOptions{FolderID: folderID, Agent: reportsListAgent},
		)
		if err != nil {
			return err
		}

		if reportsListJSON {
			return output.PrintRawJSON(out, body)
		}
		if reportsListYAML {
			return output.PrintRawYAML(out, body)
		}
		return output.PrintReportList(out, reports, inputs.ReportFolderNames(folders), columns)
	},
}

var reportsGetCmd = &cobra.Command{
	Use:     "get <name|id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Show a report, or print it as a document to save",
	Long: `Show a report's settings, sections and widgets. In the Copilot, while the report is open
in the user's editor, it shows the editor's unsaved draft instead, and --yaml and --json print
it with its draft_revision.`,
	Example: `  iai reports get "Traffic summary"
  iai reports get "Traffic summary" --yaml > report.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		pCtx, client, _, err := resolveProject(cmd.Context(), reportsOrganization, reportsProject)
		if err != nil {
			return err
		}
		id, err := findReportID(cmd.Context(), client, pCtx.projectId, args[0], true)
		if err != nil {
			return err
		}
		report, raw, err := client.GetReportOrDraft(
			cmd.Context(), pCtx.projectId, id, inputs.ReportsReadDraft(clientOrigin),
		)
		if err != nil {
			return err
		}
		folder, err := reportFolder(cmd.Context(), client, pCtx.projectId, report.FolderID)
		if err != nil {
			return err
		}
		if reportsGetJSON || reportsGetYAML {
			doc, err := inputs.ReportDocumentOf(raw, folder, true)
			if err != nil {
				return err
			}
			if reportsGetJSON {
				return output.PrintRawJSON(out, doc)
			}
			return output.PrintRawYAML(out, doc)
		}
		return output.PrintReportSummary(out, report, folder)
	},
}

var reportsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a report from a document",
	Long: `Create a report from a YAML or JSON document, as 'iai reports get --yaml' prints it.

The document's updated_at is ignored, so a printed report saves as a copy under another
name. folder names the report's folder, and a folder that does not exist is created.
Leaving out default_range sets last7Days. Fields only the platform sets, such as
created_at and updated_by, are ignored.`,
	Example: `  iai reports get "Traffic summary" --yaml > report.yaml
  iai reports create --file report.yaml --dry-run
  iai reports create --file - < report.yaml`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !reportsCreateDryRun {
			if err := requireLoginAuth(reportWrites); err != nil {
				return err
			}
		}
		doc, err := inputs.ReadReportDocument(reportsCreateFile, cmd.InOrStdin())
		if err != nil {
			return err
		}
		return saveReport(cmd, doc, "", false, reportsCreateDryRun)
	},
}

var reportsUpdateCmd = &cobra.Command{
	Use:   "update <name|id>",
	Short: "Replace a report with a document",
	Long: `Replace a report with a YAML or JSON document, as 'iai reports get --yaml' prints it.

Refused if the report was saved after the document's updated_at; --force overwrites it.
A document with draft_revision, as 'iai reports get' prints an open draft, is saved into that
draft, not the report, and only that revision guards it (updated_at and --force don't apply);
refused if the draft is past that revision, or was saved or discarded.
The save prints the draft's new draft_revision for the next one.
folder names the report's folder (null for none; a folder that does not exist is created,
except in a draft),
and leaving it out keeps the report where it is. Leaving out description, terms, filters
or agent_ids clears them, and leaving out default_range sets last7Days.`,
	Example: `  iai reports get "Traffic summary" --yaml > report.yaml
  iai reports update "Traffic summary" --file report.yaml --dry-run
  iai reports update "Traffic summary" --file report.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !reportsUpdateDryRun {
			if err := requireLoginAuth(reportWrites); err != nil {
				return err
			}
		}
		doc, err := inputs.ReadReportDocument(reportsUpdateFile, cmd.InOrStdin())
		if err != nil {
			return err
		}
		if doc.UpdatedAt == "" && doc.DraftRevision == nil && !reportsUpdateForce {
			return fmt.Errorf(
				"the document has no updated_at; start from 'iai reports get %q --yaml', or pass --force",
				args[0],
			)
		}
		return saveReport(cmd, doc, args[0], reportsUpdateForce, reportsUpdateDryRun)
	},
}

// saveReport creates doc as a new report (ref empty) or replaces the report ref names;
// dryRun prints the changes instead.
func saveReport(
	cmd *cobra.Command,
	doc inputs.ReportDocument,
	ref string,
	force, dryRun bool,
) error {
	ctx, out := cmd.Context(), cmd.OutOrStdout()
	pCtx, client, _, err := resolveProject(ctx, reportsOrganization, reportsProject)
	if err != nil {
		return err
	}
	var saved *platform.Report
	before := ""
	if ref != "" {
		report, _, err := findReport(ctx, client, pCtx.projectId, ref, true)
		if err != nil {
			return err
		}
		if before, err = reportFolder(ctx, client, pCtx.projectId, report.FolderID); err != nil {
			return err
		}
		saved = &report
	}
	if dryRun {
		after, keep := inputs.ReportFolderChange(doc)
		if keep {
			after = before
		}
		output.PrintReportSavePreview(out, saved, doc, before, after, force)
		return nil
	}
	id := ""
	if saved != nil {
		id = saved.ID
	}
	result, err := client.ApplyReport(
		ctx, pCtx.projectId,
		inputs.ReportBodyOf(doc, id, force, inputs.ReportsUpdatedVia(clientOrigin)),
	)
	if err != nil {
		if ref == "" {
			ref = doc.Name
		}
		return output.ReportWriteError(err, ref, doc.Name)
	}
	output.PrintReportSaved(out, result, reportPath(pCtx.projectId, result.ID))
	return nil
}

var reportsRestoreCmd = &cobra.Command{
	Use:   "restore <name>",
	Short: "Restore a report's previous saved version",
	Long: `Swap a report with the version saved before its last change. Restoring twice brings
the change back.`,
	Example: `  iai reports restore "Traffic summary"
  iai reports restore "Traffic summary" --expected-updated-at 2026-11-10T09:30:00Z`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		if err := requireLoginAuth(reportWrites); err != nil {
			return err
		}
		var expected *string
		if cmd.Flags().Changed("expected-updated-at") {
			expected = &reportsRestoreExpectedUpdatedAt
		}
		pCtx, client, _, err := resolveProject(cmd.Context(), reportsOrganization, reportsProject)
		if err != nil {
			return err
		}
		found, _, err := findReport(cmd.Context(), client, pCtx.projectId, args[0], false)
		if err != nil {
			return err
		}
		body, err := inputs.RestoreReportBodyOf(
			found.UpdatedAt,
			expected,
			inputs.ReportsUpdatedVia(clientOrigin),
		)
		if err != nil {
			return err
		}
		restored, err := client.RestoreReport(cmd.Context(), pCtx.projectId, found.ID, body)
		if err != nil {
			return output.ReportRestoreError(err, found.Name)
		}
		output.PrintReportRestored(out, restored, reportPath(pCtx.projectId, restored.ID))
		return nil
	},
}

var reportsDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm"},
	Short:   "Delete a report",
	Long: `Delete a report. Prints the deleted report as a document, so 'iai reports create'
can create it again.`,
	Example: `  iai reports delete "Traffic summary"
  iai reports delete "Traffic summary" -f > deleted-report.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		if err := requireLoginAuth(reportWrites); err != nil {
			return err
		}
		pCtx, client, _, err := resolveProject(cmd.Context(), reportsOrganization, reportsProject)
		if err != nil {
			return err
		}
		found, _, err := findReport(cmd.Context(), client, pCtx.projectId, args[0], false)
		if err != nil {
			return err
		}
		if !reportsDeleteForce {
			// The prompt goes to stderr: stdout carries the deleted report's document.
			confirmed, err := confirmDeletion(
				cmd.InOrStdin(), cmd.ErrOrStderr(),
				fmt.Sprintf("report %q", output.Printable(found.Name)),
			)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
				return nil
			}
		}
		folder, err := reportFolder(cmd.Context(), client, pCtx.projectId, found.FolderID)
		if err != nil {
			return err
		}
		raw, err := client.DeleteReport(cmd.Context(), pCtx.projectId, found.ID)
		if err != nil {
			return err
		}
		doc, err := inputs.ReportDocumentOf(raw, folder, false)
		if err != nil {
			return err
		}
		return output.PrintDeletedReport(out, found.Name, found.ID, doc)
	},
}

var reportsSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Show the legal widget selections and the report limits",
	Long: `Show the widget rules (the sources, their metrics and breakdowns, the chart types and
every legal selection) and the limits a report must keep. A widget that matches no
selection is refused on save.`,
	Example: `  iai reports schema
  iai reports schema --json | jq '.selections[] | select(.view == "llms")'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		pCtx, client, _, err := resolveProject(cmd.Context(), reportsOrganization, reportsProject)
		if err != nil {
			return err
		}
		rules, raw, err := client.GetReportWidgetRules(cmd.Context(), pCtx.projectId)
		if err != nil {
			return err
		}
		if reportsSchemaJSON {
			return output.PrintRawJSON(out, raw)
		}
		if reportsSchemaYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintReportRules(out, rules)
	},
}

func init() {
	reportsCmd.PersistentFlags().
		StringVarP(&reportsOrganization, "organization", "o", "", "Organization name that owns the project")
	reportsCmd.PersistentFlags().
		StringVarP(&reportsProject, "project", "p", "", "Project name that owns the reports")

	reportsListCmd.Flags().
		StringVar(&reportsListFolder, "folder", "", "Only reports in this folder; none for reports in no folder")
	reportsListCmd.Flags().
		StringVar(&reportsListAgent, "agent", "", "Only reports attached to this agent id")
	reportsListCmd.Flags().
		StringSliceVar(&reportsListColumns, "columns", nil, fmt.Sprintf(
			"Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml. Available: %s",
			strings.Join(inputs.AllReportColumns, ", "),
		))
	reportsListCmd.Flags().
		BoolVar(&reportsListJSON, "json", false, "Output raw API response as JSON")
	reportsListCmd.Flags().
		BoolVar(&reportsListYAML, "yaml", false, "Output raw API response as YAML")
	reportsListCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	reportsGetCmd.Flags().
		BoolVar(&reportsGetJSON, "json", false, "Output the report as a JSON document that 'iai reports update' reads back")
	reportsGetCmd.Flags().
		BoolVar(&reportsGetYAML, "yaml", false, "Output the report as a YAML document that 'iai reports update' reads back")
	reportsGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	reportsCreateCmd.Flags().
		StringVar(&reportsCreateFile, "file", "", "Path to the YAML or JSON report document; - reads stdin")
	reportsCreateCmd.Flags().
		BoolVar(&reportsCreateDryRun, "dry-run", false, "Print the report that would be created; save nothing")
	_ = reportsCreateCmd.MarkFlagRequired("file")
	indexFlag(reportsCreateCmd, "file", "- reads stdin")

	reportsUpdateCmd.Flags().
		StringVar(&reportsUpdateFile, "file", "", "Path to the YAML or JSON report document; - reads stdin")
	reportsUpdateCmd.Flags().
		BoolVar(&reportsUpdateForce, "force", false, "Overwrite the report even if it changed after the document's updated_at")
	reportsUpdateCmd.Flags().
		BoolVar(&reportsUpdateDryRun, "dry-run", false, "Print what would change against the saved report; save nothing")
	_ = reportsUpdateCmd.MarkFlagRequired("file")
	indexFlag(reportsUpdateCmd, "file", "- reads stdin")

	reportsRestoreCmd.Flags().
		StringVar(&reportsRestoreExpectedUpdatedAt, "expected-updated-at", "", "Fail without restoring unless the report's updated_at equals this value ('iai reports get --yaml' prints it); defaults to the updated_at read just before")

	reportsDeleteCmd.Flags().
		BoolVarP(&reportsDeleteForce, "force", "f", false, "Skip the confirmation prompt")

	reportsSchemaCmd.Flags().
		BoolVar(&reportsSchemaJSON, "json", false, "Output raw API response as JSON")
	reportsSchemaCmd.Flags().
		BoolVar(&reportsSchemaYAML, "yaml", false, "Output raw API response as YAML")
	reportsSchemaCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	reportsCmd.AddCommand(reportsListCmd)
	reportsCmd.AddCommand(reportsGetCmd)
	reportsCmd.AddCommand(reportsCreateCmd)
	reportsCmd.AddCommand(reportsUpdateCmd)
	reportsCmd.AddCommand(reportsRestoreCmd)
	reportsCmd.AddCommand(reportsDeleteCmd)
	reportsCmd.AddCommand(reportsSchemaCmd)
	rootCmd.AddCommand(reportsCmd)
}
