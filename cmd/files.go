package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultFilesTimeout = 15 * time.Minute

var (
	filesUploadName    string
	filesUploadTimeout time.Duration
	filesUploadOrg     string
	filesUploadProject string

	filesListLimit   int
	filesListCursor  string
	filesListColumns []string
	filesListJSON    bool
	filesListYAML    bool
	filesListOrg     string
	filesListProject string

	filesDownloadVersion string
	filesDownloadOutput  string
	filesDownloadForce   bool
	filesDownloadTimeout time.Duration
	filesDownloadOrg     string
	filesDownloadProject string
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

var filesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the files a project holds",
	Long:    `List a project's files, one entry per document at its current version.`,
	Example: `  iai files list
  iai files list --limit 20
  iai files list --cursor <cursor>
  iai files list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := filesListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultFileColumns
		}
		if err := validateTableOnlyColumns(cmd, filesListJSON, filesListYAML); err != nil {
			return err
		}
		if !filesListJSON && !filesListYAML {
			if err := inputs.ValidateColumns(columns, inputs.AllFileColumns); err != nil {
				return err
			}
		}

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), filesListOrg, filesListProject)
		if err != nil {
			return err
		}

		opts := platform.FileListOptions{Cursor: filesListCursor}
		if cmd.Flags().Changed("limit") {
			opts.Limit = &filesListLimit
		}
		files, meta, rawJSON, err := apiClient.ListFiles(cmd.Context(), pCtx.orgId, pCtx.projectId, opts)
		if err != nil {
			return err
		}

		if filesListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}
		if filesListYAML {
			return output.PrintRawYAML(out, rawJSON)
		}
		return output.PrintFileList(out, files, meta, columns)
	},
}

var filesDownloadCmd = &cobra.Command{
	Use:   "download <id|name>",
	Short: "Download a file's bytes",
	Long: `Download a file's current version, or an older one with --version.

Without --output, the file is written under the name the server reports,
reduced to a safe local filename; --output - streams to stdout instead.`,
	Example: `  iai files download <id|name>
  iai files download <id|name> --version <version-id>
  iai files download <id|name> --output report.pdf
  iai files download <id|name> --output - > report.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileRef := args[0]

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(), filesDownloadOrg, filesDownloadProject,
			resolveOpts{deployTimeout: defaultHTTPTimeout, filesTimeout: filesDownloadTimeout},
		)
		if err != nil {
			return err
		}

		bar := output.NewProgressBar(cmd.ErrOrStderr(), 0, fmt.Sprintf("Downloading %s", fileRef))
		download := func(dest io.Writer) (string, int64, error) {
			if filesDownloadVersion != "" {
				return apiClient.DownloadFileVersion(
					cmd.Context(), pCtx.orgId, pCtx.projectId, fileRef, filesDownloadVersion, dest, bar.Add,
				)
			}
			return apiClient.DownloadFile(cmd.Context(), pCtx.orgId, pCtx.projectId, fileRef, dest, bar.Add)
		}

		if filesDownloadOutput == "-" {
			_, _, err := download(cmd.OutOrStdout())
			bar.Finish()
			return reportFileRefAmbiguous(cmd, err)
		}

		tmp, err := os.CreateTemp(".", ".iai-files-download-*")
		if err != nil {
			return fmt.Errorf("failed to create temp download file: %w", err)
		}
		tmpPath := tmp.Name()
		rawName, _, err := download(tmp)
		bar.Finish()
		closeErr := tmp.Close()
		if err != nil {
			os.Remove(tmpPath)
			return reportFileRefAmbiguous(cmd, err)
		}
		if closeErr != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write download to disk: %w", closeErr)
		}

		target := filesDownloadOutput
		if target == "" {
			target = safeDownloadFilename(rawName, fileRef)
		}
		if _, err := os.Stat(target); err == nil && !filesDownloadForce {
			os.Remove(tmpPath)
			return fmt.Errorf("%s already exists; use --force to overwrite", target)
		}
		if err := os.Rename(tmpPath, target); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write %s: %w", target, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", target)
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

	filesListCmd.Flags().IntVar(&filesListLimit, "limit", 0, "Results per page (server default: 50)")
	filesListCmd.Flags().StringVar(&filesListCursor, "cursor", "", "Cursor from a previous page's next-page footer")
	filesListCmd.Flags().StringSliceVar(&filesListColumns, "columns", nil, "Columns to display")
	filesListCmd.Flags().BoolVar(&filesListJSON, "json", false, "Output raw API response as JSON")
	filesListCmd.Flags().BoolVar(&filesListYAML, "yaml", false, "Output raw API response as YAML")
	filesListCmd.Flags().StringVarP(&filesListOrg, "organization", "o", "", "Organization name that owns the project")
	filesListCmd.Flags().StringVarP(&filesListProject, "project", "p", "", "Project name")

	filesDownloadCmd.Flags().StringVar(&filesDownloadVersion, "version", "", "Download this specific version instead of the current one")
	filesDownloadCmd.Flags().StringVar(&filesDownloadOutput, "output", "", "Local path to write to (default: the stored name); - for stdout")
	filesDownloadCmd.Flags().BoolVarP(&filesDownloadForce, "force", "f", false, "Overwrite an existing local file")
	filesDownloadCmd.Flags().
		DurationVar(&filesDownloadTimeout, "timeout", defaultFilesTimeout, "HTTP timeout for the download")
	filesDownloadCmd.Flags().
		StringVarP(&filesDownloadOrg, "organization", "o", "", "Organization name that owns the project")
	filesDownloadCmd.Flags().StringVarP(&filesDownloadProject, "project", "p", "", "Project name")

	filesCmd.AddCommand(filesUploadCmd)
	filesCmd.AddCommand(filesListCmd)
	filesCmd.AddCommand(filesDownloadCmd)
	rootCmd.AddCommand(filesCmd)
}

func safeDownloadFilename(raw, fileID string) string {
	base := filepath.Base(raw)
	if base == "" || base == "." || base == ".." || base == string(filepath.Separator) {
		return fileID
	}
	return base
}

// reportFileRefAmbiguous prints the candidate files behind a *platform.FileRefAmbiguousError
// and returns err unchanged, so every files verb can inherit the same rendering by calling this.
func reportFileRefAmbiguous(cmd *cobra.Command, err error) error {
	var ambiguous *platform.FileRefAmbiguousError
	if errors.As(err, &ambiguous) {
		output.PrintFileRefCandidates(cmd.ErrOrStderr(), ambiguous.Ref, ambiguous.Candidates)
	}
	return err
}
