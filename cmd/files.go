package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultFilesTimeout = 15 * time.Minute

var (
	filesUploadName    string
	filesUploadTimeout time.Duration
	filesUploadOrg     string
	filesUploadProject string
)

var filesCmd = &cobra.Command{
	Use:              "files",
	Aliases:          []string{"file"},
	Short:            "Manage documents stored in a project",
	GroupID:          groupInfra,
	Long:             `Put documents into a project, list what it holds, and fetch them back. Works with API key (--api-key or INTERACTIVE_API_KEY) or session from 'iai login'.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var filesUploadCmd = &cobra.Command{
	Use:   "upload <local-file>",
	Short: "Upload a document into a project",
	Long: `Upload a local file into a project as a new document.

The stored name defaults to the local file's base name; use --name to store
it under a different name.`,
	Example: `  iai files upload ./report.pdf
  iai files upload ./report.pdf --name "Q3 Report.pdf"
  iai files upload ./big.bin --timeout 30m`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		localPath := args[0]

		info, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", localPath, err)
		}

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(), filesUploadOrg, filesUploadProject,
			resolveOpts{deployTimeout: defaultHTTPTimeout, filesTimeout: filesUploadTimeout},
		)
		if err != nil {
			return err
		}

		label := fmt.Sprintf("Uploading %s", filepath.Base(localPath))
		bar := output.NewProgressBar(cmd.ErrOrStderr(), info.Size(), label)
		result, err := apiClient.CreateFile(
			cmd.Context(), pCtx.orgId, pCtx.projectId, localPath, filesUploadName,
			bar.Add,
		)
		bar.Finish()
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "Uploaded %s\n", result.Current.Name)
		fmt.Fprintf(out, "  id:      %s\n", result.Id)
		fmt.Fprintf(out, "  version: %s\n", result.Current.VersionId)
		fmt.Fprintf(out, "  size:    %s\n", output.HumanBytes(result.Current.Size))

		return nil
	},
}

func init() {
	filesUploadCmd.Flags().StringVar(&filesUploadName, "name", "", "Name to store the file under (default: the local file's name)")
	filesUploadCmd.Flags().
		DurationVar(&filesUploadTimeout, "timeout", defaultFilesTimeout, "HTTP timeout for the upload")
	filesUploadCmd.Flags().
		StringVarP(&filesUploadOrg, "organization", "o", "", "Organization name that owns the project")
	filesUploadCmd.Flags().StringVarP(&filesUploadProject, "project", "p", "", "Project name")

	filesCmd.AddCommand(filesUploadCmd)
	rootCmd.AddCommand(filesCmd)
}
