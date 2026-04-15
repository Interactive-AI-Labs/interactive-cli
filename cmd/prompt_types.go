package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

type PromptTypeConfig struct {
	TypeName     string   // singular name, e.g. "routine"
	Plural       string   // plural name used as command, e.g. "routines"
	Aliases      []string // command aliases, e.g. ["routine"]
	Short        string   // short description for the parent command
	Long         string   // long description for the parent command
	RouteSegment string   // API URL segment for type-specific routes, e.g. "routines"
	HasSchema    bool     // whether this type supports the schema subcommand
	CreateLong   string   // long description for the create subcommand
	ListLong     string   // long description for the list subcommand
	GetLong      string   // long description for the get subcommand
	UpdateLong   string   // long description for the update subcommand
	DeleteLong   string   // long description for the delete subcommand
}

func registerPromptType(ptCfg PromptTypeConfig) {
	parentCmd := &cobra.Command{
		Use:     ptCfg.Plural,
		Aliases: ptCfg.Aliases,
		Short:   ptCfg.Short,
		Long:    ptCfg.Long,
	}

	createCmd := makeCreateCmd(ptCfg)
	listCmd := makeListCmd(ptCfg)
	getCmd := makeGetCmd(ptCfg)
	updateCmd := makeUpdateCmd(ptCfg)
	deleteCmd := makeDeleteCmd(ptCfg)

	parentCmd.AddCommand(createCmd, listCmd, getCmd, updateCmd, deleteCmd)

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
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			apiClient, err := clients.NewAPIClient(
				hostname,
				defaultHTTPTimeout,
				token,
				apiKey,
				cookies,
			)
			if err != nil {
				return fmt.Errorf("failed to create API client: %w", err)
			}

			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file %q: %w", file, err)
			}

			body := clients.CreatePromptBody{
				Name:   name,
				Prompt: string(content),
				Labels: labels,
				Tags:   tags,
			}

			fmt.Fprintln(out)
			fmt.Fprintf(out, "Creating %s %q...\n", ptCfg.TypeName, name)

			result, err := apiClient.CreatePrompt(
				cmd.Context(),
				pCtx.projectId,
				ptCfg.RouteSegment,
				body,
			)
			if err != nil {
				return err
			}

			fmt.Fprintln(out)
			return output.PrintPromptDetail(out, result)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to the file containing the prompt content")
	cmd.Flags().
		StringSliceVar(&labels, "labels", nil, "Labels for the prompt version (comma-separated)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the prompt (comma-separated)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	_ = cmd.MarkFlagRequired("file")

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

			pCtx, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			apiClient, err := clients.NewAPIClient(
				hostname,
				defaultHTTPTimeout,
				token,
				apiKey,
				cookies,
			)
			if err != nil {
				return fmt.Errorf("failed to create API client: %w", err)
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
		Use:     "get <name>",
		Aliases: []string{"describe", "desc"},
		Short:   fmt.Sprintf("Get details of a %s", ptCfg.TypeName),
		Long:    ptCfg.GetLong,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			apiClient, err := clients.NewAPIClient(
				hostname,
				defaultHTTPTimeout,
				token,
				apiKey,
				cookies,
			)
			if err != nil {
				return fmt.Errorf("failed to create API client: %w", err)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			apiClient, err := clients.NewAPIClient(
				hostname,
				defaultHTTPTimeout,
				token,
				apiKey,
				cookies,
			)
			if err != nil {
				return fmt.Errorf("failed to create API client: %w", err)
			}

			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file %q: %w", file, err)
			}

			body := clients.CreatePromptBody{
				Name:   name,
				Prompt: string(content),
				Labels: labels,
				Tags:   tags,
			}

			fmt.Fprintln(out)
			fmt.Fprintf(out, "Updating %s %q...\n", ptCfg.TypeName, name)

			// CreatePrompt is intentional: the API creates a new version when the
			// prompt name already exists, so create and update use the same endpoint.
			result, err := apiClient.CreatePrompt(
				cmd.Context(),
				pCtx.projectId,
				ptCfg.RouteSegment,
				body,
			)
			if err != nil {
				return err
			}

			fmt.Fprintln(out)
			return output.PrintPromptDetail(out, result)
		},
	}

	cmd.Flags().
		StringVar(&file, "file", "", "Path to the file containing the updated prompt content")
	cmd.Flags().
		StringSliceVar(&labels, "labels", nil, "Labels for the new prompt version (comma-separated)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the prompt (comma-separated)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	_ = cmd.MarkFlagRequired("file")

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

			pCtx, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			apiClient, err := clients.NewAPIClient(
				hostname,
				defaultHTTPTimeout,
				token,
				apiKey,
				cookies,
			)
			if err != nil {
				return fmt.Errorf("failed to create API client: %w", err)
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
