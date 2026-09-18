package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

// ConfigFlagBuilder assembles the payload's "config" field; nil if no flags set.
type ConfigFlagBuilder func() map[string]any

type PromptTypeConfig struct {
	TypeName     string   // singular name, e.g. "routine"
	Plural       string   // plural name used as command, e.g. "routines"
	Aliases      []string // command aliases, e.g. ["routine"]
	Short        string   // short description for the parent command
	Long         string   // long description for the parent command
	RouteSegment string   // API URL segment for type-specific routes, e.g. "routines"
	HasSchema    bool     // whether this type supports the schema subcommand
	GroupID      string   // command group shown in iai --help; defaults to groupContext
	// BindPromptConfigFlags registers type-specific flags and returns a config builder.
	BindPromptConfigFlags func(cmd *cobra.Command) ConfigFlagBuilder
	// GlobalScope makes list and get also read scope=global on the same routes:
	// the shared, read-only records Interactive serves to every project.
	GlobalScope   bool
	CreateLong    string // long description for the create subcommand
	ListLong      string // long description for the list subcommand
	GetLong       string // long description for the describe subcommand
	UpdateLong    string // long description for the update subcommand
	DeleteLong    string // long description for the delete subcommand
	CreateExample string // usage examples for the create subcommand
	ListExample   string // usage examples for the list subcommand
	GetExample    string // usage examples for the describe subcommand
	UpdateExample string // usage examples for the update subcommand
	DeleteExample string // usage examples for the delete subcommand
}

func registerPromptType(ptCfg PromptTypeConfig) {
	parentCmd := &cobra.Command{
		Use:     ptCfg.Plural,
		Aliases: ptCfg.Aliases,
		Short:   ptCfg.Short,
		Long:    ptCfg.Long,
		GroupID: func() string {
			if ptCfg.GroupID != "" {
				return ptCfg.GroupID
			}
			return groupContext
		}(),
	}

	createCmd := makeCreateCmd(ptCfg)
	listCmd := makeListCmd(ptCfg)
	getCmd := makeGetCmd(ptCfg)
	updateCmd := makeUpdateCmd(ptCfg)
	deleteCmd := makeDeleteCmd(ptCfg)

	versionsCmd := makeVersionsCmd(ptCfg)
	diffCmd := makeDiffCmd(ptCfg)

	parentCmd.AddCommand(createCmd, listCmd, getCmd, updateCmd, deleteCmd, versionsCmd, diffCmd)

	if ptCfg.HasSchema {
		schemaCmd := makeSchemaCmd(ptCfg)
		parentCmd.AddCommand(schemaCmd)
	}

	rootCmd.AddCommand(parentCmd)
}

func makeSchemaCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		schemaVersion string
		asJSON        bool
		asYAML        bool
	)

	cmd := &cobra.Command{
		Use:   "schema",
		Short: fmt.Sprintf("Display the JSON Schema for %s", ptCfg.Plural),
		Long: fmt.Sprintf(`Fetch and display the JSON Schema for %s from the backend API.

Use --schema-version to request a specific schema version (defaults to latest stable).
Use --json or --yaml to output the schema response in a structured format.

This is a public endpoint and does not require authentication.`, ptCfg.Plural),
		Example: fmt.Sprintf(`  iai %s schema
  iai %s schema --schema-version 0.0.1
  iai %s schema --json`, ptCfg.Plural, ptCfg.Plural, ptCfg.Plural),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			result, err := platform.GetPromptSchema(
				cmd.Context(), hostname, defaultHTTPTimeout, ptCfg.TypeName, schemaVersion,
			)
			if err != nil {
				return err
			}

			if asJSON {
				return output.PrintStructuredJSON(out, result)
			}
			if asYAML {
				return output.PrintStructuredYAML(out, result)
			}

			return output.PrintSchemaPretty(out, result.Schema, result.SchemaVersion)
		},
	}

	cmd.Flags().
		StringVar(&schemaVersion, "schema-version", "", "Schema version to fetch (defaults to latest stable)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output schema response as JSON")
	cmd.Flags().BoolVar(&asYAML, "yaml", false, "Output schema response as YAML")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func makeCreateCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		file          string
		labels        []string
		tags          []string
		project       string
		org           string
		schemaVersion string
		message       string
	)

	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   fmt.Sprintf("Create a %s", ptCfg.TypeName),
		Long:    ptCfg.CreateLong,
		Example: ptCfg.CreateExample,
		Args:    cobra.ExactArgs(1),
	}

	var configBuilder ConfigFlagBuilder
	if ptCfg.BindPromptConfigFlags != nil {
		configBuilder = ptCfg.BindPromptConfigFlags(cmd)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", file, err)
		}

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
		if err != nil {
			return err
		}

		payload := platform.CreatePromptBody{
			Name:          name,
			Prompt:        string(content),
			Labels:        labels,
			Tags:          tags,
			SchemaVersion: schemaVersion,
			CommitMessage: message,
		}
		if configBuilder != nil {
			payload.Config = configBuilder()
		}

		fmt.Fprintln(out)
		fmt.Fprintf(out, "Creating %s %q...\n", ptCfg.TypeName, name)

		result, err := apiClient.CreatePrompt(
			cmd.Context(),
			pCtx.projectId,
			ptCfg.RouteSegment,
			payload,
		)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		return output.PrintPromptDetail(out, result)
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to the file containing the prompt content")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().
		StringSliceVar(&labels, "labels", nil, "Labels for the prompt version (comma-separated)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the prompt (comma-separated)")
	cmd.Flags().
		StringVar(&schemaVersion, "schema-version", "", "Schema version to validate against (defaults to latest stable)")
	cmd.Flags().
		StringVarP(&message, "message", "m", "", "Commit message describing the change (stored on the new version)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeListCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		page    int
		limit   int
		folder  string
		project string
		org     string
		asJSON  bool
		asYAML  bool
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   fmt.Sprintf("List %s in a project", ptCfg.Plural),
		Long:    ptCfg.ListLong,
		Example: ptCfg.ListExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			opts := platform.PromptListOptions{
				Page:  page,
				Limit: limit,
			}
			if folder != "" {
				opts.Subfolder = strings.TrimSpace(folder)
				if strings.Contains(opts.Subfolder, "..") {
					return fmt.Errorf(
						"invalid folder path %q: must not contain '..'",
						opts.Subfolder,
					)
				}
			}

			result, err := apiClient.ListPrompts(
				cmd.Context(),
				pCtx.projectId,
				ptCfg.RouteSegment,
				opts,
			)
			if err != nil {
				return err
			}

			// Global records are project-wide and unpaginated, so they belong to the
			// root listing only: not inside a folder, and not past the first page.
			// Pages are 0-indexed server-side, so anything <= 0 is that first page.
			if ptCfg.GlobalScope && folder == "" && page <= 0 {
				// Same Limit as the project call on purpose: a server that does not
				// know scope=global ignores it and answers with the project's own
				// page, and only an identical page de-dupes away to nothing. The
				// global scope itself is unpaginated and ignores Limit.
				globalResult, globalErr := apiClient.ListPrompts(
					cmd.Context(),
					pCtx.projectId,
					ptCfg.RouteSegment,
					platform.PromptListOptions{Limit: limit, Scope: platform.ScopeGlobal},
				)
				if globalErr != nil {
					// Hiding the project's own rows because the shared read failed is
					// worse than a listing that is missing the shared ones.
					fmt.Fprintf(
						cmd.ErrOrStderr(),
						"Warning: could not load general %s: %v\n",
						ptCfg.Plural,
						globalErr,
					)
				} else {
					result.Prompts = mergeGlobalRows(result.Prompts, globalResult.Prompts)
				}
			}

			if asJSON {
				return output.PrintStructuredJSON(out, result)
			}
			if asYAML {
				return output.PrintStructuredYAML(out, result)
			}

			return output.PrintPromptList(out, ptCfg.Plural, result.Prompts)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number for pagination")
	cmd.Flags().IntVar(&limit, "limit", 0, "Number of items per page (default: 50)")
	cmd.Flags().StringVar(&folder, "folder", "", "List items inside the given folder path")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output response as JSON")
	cmd.Flags().BoolVar(&asYAML, "yaml", false, "Output response as YAML")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeGetCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		version int
		label   string
		project string
		org     string
		asJSON  bool
		asYAML  bool
	)

	cmd := &cobra.Command{
		Use:     "get <name>",
		Aliases: []string{"describe", "desc"},
		Short:   fmt.Sprintf("Describe a %s in detail", ptCfg.TypeName),
		Long:    ptCfg.GetLong,
		Example: ptCfg.GetExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			result, err := apiClient.GetPrompt(
				cmd.Context(),
				pCtx.projectId,
				ptCfg.RouteSegment,
				name,
				version,
				label,
				"",
			)
			if err != nil {
				if !canFallBackToGlobal(ptCfg, version, label, err) {
					return err
				}
				fallback, fallbackErr := apiClient.GetPrompt(
					cmd.Context(),
					pCtx.projectId,
					ptCfg.RouteSegment,
					name,
					0,
					"",
					platform.ScopeGlobal,
				)
				if fallbackErr != nil {
					// Not there either is the same answer, so it is not worth a warning —
					// and on a server that does not serve the shared scope yet, every
					// miss would print one. Anything else means the check did not happen.
					var fallbackNotFound *platform.NotFoundError
					if !errors.As(fallbackErr, &fallbackNotFound) {
						fmt.Fprintf(
							cmd.ErrOrStderr(),
							"Warning: could not check general %s: %v\n",
							ptCfg.Plural,
							fallbackErr,
						)
					}
					return err
				}
				if servedFromProjectScope(fallback) {
					return err
				}
				fallback.Source = sourceGeneral
				result = fallback
			}

			if asJSON {
				return output.PrintStructuredJSON(out, result)
			}
			if asYAML {
				return output.PrintStructuredYAML(out, result)
			}

			return output.PrintPromptDetail(out, result)
		},
	}

	cmd.Flags().IntVar(&version, "version", 0, "Retrieve a specific version number")
	cmd.Flags().StringVar(&label, "label", "", "Retrieve the version with this label")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output response as JSON")
	cmd.Flags().BoolVar(&asYAML, "yaml", false, "Output response as YAML")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeUpdateCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		file          string
		labels        []string
		tags          []string
		project       string
		org           string
		schemaVersion string
		message       string
	)

	cmd := &cobra.Command{
		Use:     "update <name>",
		Short:   fmt.Sprintf("Update a %s (creates a new version)", ptCfg.TypeName),
		Long:    ptCfg.UpdateLong,
		Example: ptCfg.UpdateExample,
		Args:    cobra.ExactArgs(1),
	}

	var configBuilder ConfigFlagBuilder
	if ptCfg.BindPromptConfigFlags != nil {
		configBuilder = ptCfg.BindPromptConfigFlags(cmd)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", file, err)
		}

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
		if err != nil {
			return err
		}

		payload := platform.CreatePromptBody{
			Name:          name,
			Prompt:        string(content),
			Labels:        labels,
			Tags:          tags,
			SchemaVersion: schemaVersion,
			CommitMessage: message,
		}
		if configBuilder != nil {
			payload.Config = configBuilder()
		}

		fmt.Fprintln(out)
		fmt.Fprintf(out, "Updating %s %q...\n", ptCfg.TypeName, name)

		// CreatePrompt is intentional: the API creates a new version when the name exists.
		result, err := apiClient.CreatePrompt(
			cmd.Context(),
			pCtx.projectId,
			ptCfg.RouteSegment,
			payload,
		)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		return output.PrintPromptDetail(out, result)
	}

	cmd.Flags().
		StringVar(&file, "file", "", "Path to the file containing the updated prompt content")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().
		StringSliceVar(&labels, "labels", nil, "Labels for the new prompt version (comma-separated)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the prompt (comma-separated)")
	cmd.Flags().
		StringVar(&schemaVersion, "schema-version", "", "Schema version to validate against (defaults to latest stable)")
	cmd.Flags().
		StringVarP(&message, "message", "m", "", "Commit message describing the change (stored on the new version)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeDeleteCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		version int
		label   string
		force   bool
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   fmt.Sprintf("Delete a %s", ptCfg.TypeName),
		Long:    ptCfg.DeleteLong,
		Example: ptCfg.DeleteExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			// Deleting all versions is destructive; require confirmation.
			if version == 0 && label == "" && !force {
				fmt.Fprintf(
					out,
					"This will delete %s %q and all its versions. Continue? [y/N] ",
					ptCfg.TypeName,
					name,
				)
				reader := bufio.NewReader(cmd.InOrStdin())
				answer, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					fmt.Fprintln(out, "Aborted.")
					return nil
				}
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			fmt.Fprintln(out)
			fmt.Fprintf(out, "Deleting %s %q...\n", ptCfg.TypeName, name)

			if version > 0 || label != "" {
				err = apiClient.DeletePrompt(
					cmd.Context(),
					pCtx.projectId,
					ptCfg.RouteSegment,
					name,
					version,
					label,
				)
			} else {
				err = apiClient.DeletePromptByName(
					cmd.Context(),
					pCtx.projectId,
					ptCfg.RouteSegment,
					name,
				)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Successfully deleted %s %q.\n", ptCfg.TypeName, name)

			return nil
		},
	}

	cmd.Flags().IntVar(&version, "version", 0, "Delete a specific version only")
	cmd.Flags().StringVar(&label, "label", "", "Delete versions with this label only")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeVersionsCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:     "versions <name>",
		Aliases: []string{"vers"},
		Short:   fmt.Sprintf("List versions of a %s", ptCfg.TypeName),
		Long: fmt.Sprintf(`List all versions of a %s, sorted newest-first.

Each row shows the version number, when it was updated, who updated it, and the
commit message recorded with that version (set with -m on create and update).
A "—" means the value is unavailable: under API-key authentication only version
numbers can be read.`, ptCfg.TypeName),
		Example: fmt.Sprintf(`  iai %s versions my-%s`, ptCfg.Plural, ptCfg.TypeName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			versions, err := apiClient.ListPromptVersions(
				cmd.Context(), pCtx.projectId, ptCfg.RouteSegment, name,
			)
			if err != nil {
				return err
			}
			if len(versions) == 0 {
				return fmt.Errorf("%s %q not found", ptCfg.TypeName, name)
			}

			return output.PrintPromptVersions(out, versions)
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeDiffCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:     "diff <name> <version_a> <version_b>",
		Short:   fmt.Sprintf("Compare two versions of a %s", ptCfg.TypeName),
		Long:    fmt.Sprintf(`Show the differences between two versions of a %s.`, ptCfg.TypeName),
		Example: fmt.Sprintf(`  iai %s diff my-%s 1 3`, ptCfg.Plural, ptCfg.TypeName),
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			versionA, err := inputs.ParseRevisionArg(args[1])
			if err != nil {
				return err
			}
			versionB, err := inputs.ParseRevisionArg(args[2])
			if err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			a, err := apiClient.GetPrompt(
				cmd.Context(), pCtx.projectId, ptCfg.RouteSegment, name, versionA, "", "",
			)
			if err != nil {
				return err
			}

			b, err := apiClient.GetPrompt(
				cmd.Context(), pCtx.projectId, ptCfg.RouteSegment, name, versionB, "", "",
			)
			if err != nil {
				return err
			}

			return output.PrintPromptDiff(out, args[1], a, args[2], b)
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

// sourceGeneral labels a record served under the global scope. The wire calls that
// scope "global"; every user-facing string in this CLI says "general".
const sourceGeneral = "general"

// rowTypeFolder is the row_type the list endpoint uses for a folder entry.
const rowTypeFolder = "folder"

// mergeGlobalRows appends the shared rows the project does not already define,
// marking each one so the listing can say where it came from. A project skill
// wins on a name collision, matching what the Copilot loads at runtime.
//
// TotalCount is deliberately left alone: it is the server's denominator for
// paging the project's own skills, and the shared rows are not part of that
// sequence. Folders cannot shadow anything — a folder and a skill of the same
// name are different things, and the folder row already reads as "name/".
func mergeGlobalRows(project, global []platform.PromptInfo) []platform.PromptInfo {
	owned := make(map[string]bool, len(project))
	for _, p := range project {
		if p.RowType != rowTypeFolder {
			owned[p.Name] = true
		}
	}
	merged := project
	for _, row := range global {
		if owned[row.Name] {
			continue
		}
		row.Source = sourceGeneral
		merged = append(merged, row)
	}
	return merged
}

// servedFromProjectScope reports whether a reply came from the project scope
// even though the global one was asked for — which is what a server that does
// not know the parameter does with it. A global record carries neither a version
// nor an id, so either one means the answer is the project's and must not be
// relabelled as someone else's.
func servedFromProjectScope(detail *platform.PromptDetail) bool {
	return detail.Version > 0 || detail.Id != ""
}

// canFallBackToGlobal reports whether a failed project lookup should be retried
// against the shared scope. Only a 404 qualifies: any other failure is the answer.
// Global records expose one version, so a --version pins the caller to the
// project; "active" is what the help text recommends and is the only version a
// global record has, so it still resolves — any other label cannot.
func canFallBackToGlobal(ptCfg PromptTypeConfig, version int, label string, err error) bool {
	if !ptCfg.GlobalScope || version != 0 || (label != "" && label != "active") {
		return false
	}
	var notFound *platform.NotFoundError
	return errors.As(err, &notFound)
}
