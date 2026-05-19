package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

// ConfigFlagBuilder assembles the payload's "config" field from flag values.
// Returns nil if no config flags were set.
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
	// BindPromptConfigFlags registers type-specific flags on create/update
	// and returns a builder for the payload's config field.
	BindPromptConfigFlags func(cmd *cobra.Command) ConfigFlagBuilder
	CreateLong            string // long description for the create subcommand
	ListLong              string // long description for the list subcommand
	GetLong               string // long description for the describe subcommand
	UpdateLong            string // long description for the update subcommand
	DeleteLong            string // long description for the delete subcommand
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
	labelsCmd := makeLabelsCmd(ptCfg)

	parentCmd.AddCommand(
		createCmd,
		listCmd,
		getCmd,
		updateCmd,
		deleteCmd,
		versionsCmd,
		diffCmd,
		labelsCmd,
	)

	if ptCfg.HasSchema {
		schemaCmd := makeSchemaCmd(ptCfg)
		parentCmd.AddCommand(schemaCmd)
	}

	rootCmd.AddCommand(parentCmd)
}

func makeSchemaCmd(ptCfg PromptTypeConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: fmt.Sprintf("Display the JSON Schema for %s", ptCfg.Plural),
		Long: fmt.Sprintf(`Fetch and display the current JSON Schema for %s from the backend API.

This is a public endpoint and does not require authentication.`, ptCfg.Plural),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			result, err := clients.GetPromptSchema(
				cmd.Context(), hostname, defaultHTTPTimeout, ptCfg.TypeName,
			)
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Schema version: %s\n\n", result.SchemaVersion)

			var indented bytes.Buffer
			if err := json.Indent(&indented, result.Schema, "", "  "); err != nil {
				return fmt.Errorf("failed to format schema: %w", err)
			}
			fmt.Fprintln(out, indented.String())

			return nil
		},
	}
}

func makeCreateCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		file    string
		labels  []string
		tags    []string
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: fmt.Sprintf("Create a %s", ptCfg.TypeName),
		Long:  ptCfg.CreateLong,
		Args:  cobra.ExactArgs(1),
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

		payload := clients.CreatePromptBody{
			Name:   name,
			Prompt: string(content),
			Labels: labels,
			Tags:   tags,
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
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   fmt.Sprintf("List %s in a project", ptCfg.Plural),
		Long:    ptCfg.ListLong,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			opts := clients.PromptListOptions{
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

			return output.PrintPromptList(out, result.Prompts)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number for pagination")
	cmd.Flags().IntVar(&limit, "limit", 0, "Number of items per page (default: 50)")
	cmd.Flags().StringVar(&folder, "folder", "", "List items inside the given folder path")
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
	)

	cmd := &cobra.Command{
		Use:     "describe <name>",
		Aliases: []string{"desc", "get"},
		Short:   fmt.Sprintf("Describe a %s in detail", ptCfg.TypeName),
		Long:    ptCfg.GetLong,
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
			)
			if err != nil {
				return err
			}

			return output.PrintPromptDetail(out, result)
		},
	}

	cmd.Flags().IntVar(&version, "version", 0, "Retrieve a specific version number")
	cmd.Flags().
		StringVar(&label, "label", "", "Retrieve the version with this label (default: server resolves 'production')")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeUpdateCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		file    string
		labels  []string
		tags    []string
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: fmt.Sprintf("Update a %s (creates a new version)", ptCfg.TypeName),
		Long:  ptCfg.UpdateLong,
		Args:  cobra.ExactArgs(1),
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

		payload := clients.CreatePromptBody{
			Name:   name,
			Prompt: string(content),
			Labels: labels,
			Tags:   tags,
		}
		if configBuilder != nil {
			payload.Config = configBuilder()
		}

		fmt.Fprintln(out)
		fmt.Fprintf(out, "Updating %s %q...\n", ptCfg.TypeName, name)

		// CreatePrompt is intentional: the API creates a new version when the
		// prompt name already exists, so create and update use the same endpoint.
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

Examples:
  iai %s versions my-%s`, ptCfg.TypeName, ptCfg.Plural, ptCfg.TypeName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			versions, err := findPromptVersions(
				cmd.Context(), apiClient, pCtx.projectId, ptCfg.RouteSegment, name, ptCfg.TypeName,
			)
			if err != nil {
				return err
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
		Use:   "diff <name> <version_a> <version_b>",
		Short: fmt.Sprintf("Compare two versions of a %s", ptCfg.TypeName),
		Long: fmt.Sprintf(`Show the differences between two versions of a %s.

Examples:
  iai %s diff my-%s 1 3`, ptCfg.TypeName, ptCfg.Plural, ptCfg.TypeName),
		Args: cobra.ExactArgs(3),
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
				cmd.Context(), pCtx.projectId, ptCfg.RouteSegment, name, versionA, "",
			)
			if err != nil {
				return err
			}

			b, err := apiClient.GetPrompt(
				cmd.Context(), pCtx.projectId, ptCfg.RouteSegment, name, versionB, "",
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

func makeLabelsCmd(ptCfg PromptTypeConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "labels",
		Aliases: []string{"label"},
		Short:   fmt.Sprintf("Manage labels on %s versions", ptCfg.Plural),
		Long: fmt.Sprintf(`Manage labels on existing %s versions.

This command group reassigns labels to existing versions without creating a
new version. Labels are unique per prompt: assigning a label to one version
removes it from any other version that currently has it.`, ptCfg.Plural),
	}

	cmd.AddCommand(makeLabelsSetCmd(ptCfg))

	return cmd
}

func makeLabelsSetCmd(ptCfg PromptTypeConfig) *cobra.Command {
	var (
		version int
		labels  []string
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: fmt.Sprintf("Set labels on a %s version", ptCfg.TypeName),
		Long: fmt.Sprintf(
			`Set labels on an existing %s version, identified by name and version number.

Labels are unique per prompt: assigning a label to one version removes it
from any other version that currently has it. No new version is created.

Examples:
  iai %s labels set my-%s --version 3 --labels production
  iai %s labels set my-%s --version 1 --labels staging,canary`,
			ptCfg.TypeName,
			ptCfg.Plural,
			ptCfg.TypeName,
			ptCfg.Plural,
			ptCfg.TypeName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			fmt.Fprintln(out)
			fmt.Fprintf(
				out,
				"Setting labels on %s %q version %d...\n",
				ptCfg.TypeName,
				name,
				version,
			)

			result, err := apiClient.SetPromptLabels(
				cmd.Context(),
				pCtx.projectId,
				ptCfg.RouteSegment,
				name,
				version,
				labels,
			)
			if err != nil {
				return err
			}

			fmt.Fprintln(out)
			return output.PrintPromptDetail(out, result)
		},
	}

	cmd.Flags().IntVar(&version, "version", 0, "Version number to assign labels to")
	_ = cmd.MarkFlagRequired("version")
	cmd.Flags().
		StringSliceVar(&labels, "labels", nil, "Labels to assign (comma-separated)")
	_ = cmd.MarkFlagRequired("labels")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func findPromptVersions(
	ctx context.Context,
	apiClient *clients.APIClient,
	projectId, routeSegment, name, typeName string,
) ([]int, error) {
	opts := clients.PromptListOptions{Limit: 1000}
	result, err := apiClient.ListPrompts(ctx, projectId, routeSegment, opts)
	if err != nil {
		return nil, err
	}

	for _, p := range result.Prompts {
		if p.Name == name {
			return p.Versions, nil
		}
	}

	return nil, fmt.Errorf("%s %q not found", typeName, name)
}
