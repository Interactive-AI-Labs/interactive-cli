package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

// validPromptTypes lists the allowed values for the --type flag.
var validPromptTypes = []string{"text", "chat"}

func init() {
	parentCmd := &cobra.Command{
		Use:     "prompts",
		Aliases: []string{"prompt"},
		Short:   "Versioned prompts for agents, evaluators, and guardrails",
		GroupID: groupContext,
		Long: `Manage general-purpose text and chat prompts in InteractiveAI projects.

Unlike typed commands (routines, policies, glossaries, variables, macros),
prompts managed here have no enforced schema or structure. They support two
types: "text" (default) and "chat".`,
	}

	parentCmd.AddCommand(
		makeGenericCreateCmd(),
		makeGenericListCmd(),
		makeGenericGetCmd(),
		makeGenericUpdateCmd(),
		makeGenericDeleteCmd(),
		makeGenericVersionsCmd(),
		makeGenericDiffCmd(),
	)

	rootCmd.AddCommand(parentCmd)
}

func makeGenericCreateCmd() *cobra.Command {
	var (
		file       string
		content    string
		promptType string
		labels     []string
		tags       []string
		project    string
		org        string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a prompt",
		Long: `Create a new text or chat prompt in an InteractiveAI project.

Content is provided via --file (path to a file) or --content (inline string).
Exactly one of --file or --content must be specified.

The --type flag selects the prompt type: "text" (default) or "chat".

The server automatically assigns the "latest" label to new versions. To make a
version retrievable via the default 'get' (which resolves "production"), assign
the "production" label with --labels production.`,
		Example: `  iai prompts create greeting --content "Hello, how can I help you?"
  iai prompts create greeting --file greeting.txt
  iai prompts create greeting --file greeting.txt --type chat
  iai prompts create greeting --content "Hi!" --labels production
  iai prompts create greeting --file greeting.txt --tags support`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			promptContent, err := resolveContent(file, content)
			if err != nil {
				return err
			}

			if err := validatePromptType(promptType); err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			body := clients.CreatePromptBody{
				Name:       name,
				Prompt:     promptContent,
				Labels:     labels,
				Tags:       tags,
				PromptType: promptType,
			}

			fmt.Fprintln(out)
			fmt.Fprintf(out, "Creating prompt %q...\n", name)

			result, err := apiClient.CreatePrompt(
				cmd.Context(),
				pCtx.projectId,
				"", // empty route segment → generic /prompts endpoint
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
	cmd.Flags().StringVar(&content, "content", "", "Inline prompt content string")
	cmd.Flags().StringVar(
		&promptType, "type", "text",
		`Prompt type: "text" (default) or "chat"`,
	)
	cmd.Flags().
		StringSliceVar(&labels, "labels", nil, "Labels for the prompt version (comma-separated)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the prompt (comma-separated)")
	cmd.Flags().
		StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().
		StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	cmd.MarkFlagsMutuallyExclusive("file", "content")

	return cmd
}

func makeGenericListCmd() *cobra.Command {
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
		Short:   "List prompts in a project",
		Long: `List text and chat prompts in a specific project.

Returns all general-purpose prompts with their name, labels, tags, and last
update time. Typed prompts (routines, policies, etc.) are excluded.
Folders are shown with a trailing "/" and can be browsed into with --folder.`,
		Example: `  iai prompts list
  iai prompts list --folder my-folder
  iai prompts list --folder my-folder/sub-folder
  iai prompts list --page 2 --limit 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			opts := clients.PromptListOptions{
				Page:   page,
				Limit:  limit,
				Folder: "prompts",
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
				"", // empty route segment → generic /prompts endpoint
				opts,
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

			return output.PrintPromptList(out, result.Prompts)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number for pagination")
	cmd.Flags().IntVar(&limit, "limit", 0, "Number of items per page (default: 50)")
	cmd.Flags().StringVar(&folder, "folder", "", "List items inside the given folder path")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output response as JSON")
	cmd.Flags().BoolVar(&asYAML, "yaml", false, "Output response as YAML")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")
	cmd.Flags().
		StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().
		StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeGenericGetCmd() *cobra.Command {
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
		Short:   "Describe a prompt in detail",
		Long: `Show detailed information about a specific prompt, including its full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.`,
		Example: `  iai prompts get greeting
  iai prompts get greeting --version 3
  iai prompts get greeting --label staging`,
		Args: cobra.ExactArgs(1),
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
				"", // empty route segment → generic /prompts endpoint
				name,
				version,
				label,
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

			return output.PrintPromptDetail(out, result)
		},
	}

	cmd.Flags().IntVar(&version, "version", 0, "Retrieve a specific version number")
	cmd.Flags().StringVar(
		&label, "label", "",
		`Retrieve the version with this label (default: server resolves "production")`,
	)
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output response as JSON")
	cmd.Flags().BoolVar(&asYAML, "yaml", false, "Output response as YAML")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")
	cmd.Flags().
		StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().
		StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeGenericUpdateCmd() *cobra.Command {
	var (
		file    string
		content string
		labels  []string
		tags    []string
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a prompt (creates a new version)",
		Long: `Update a prompt by creating a new version with updated content.

This creates a new version of the prompt using the content from the provided
file or inline string. Previous versions are preserved and can still be accessed
by version number.

Exactly one of --file or --content must be specified.`,
		Example: `  iai prompts update greeting --content "Hello! How may I assist you today?"
  iai prompts update greeting --file greeting.txt
  iai prompts update greeting --file greeting.txt --labels production,staging`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			promptContent, err := resolveContent(file, content)
			if err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			body := clients.CreatePromptBody{
				Name:   name,
				Prompt: promptContent,
				Labels: labels,
				Tags:   tags,
			}

			fmt.Fprintln(out)
			fmt.Fprintf(out, "Updating prompt %q...\n", name)

			// CreatePrompt is intentional: the API creates a new version when the
			// prompt name already exists, so create and update use the same endpoint.
			result, err := apiClient.CreatePrompt(
				cmd.Context(),
				pCtx.projectId,
				"", // empty route segment → generic /prompts endpoint
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
	cmd.Flags().StringVar(&content, "content", "", "Inline updated prompt content string")
	cmd.Flags().StringSliceVar(
		&labels, "labels", nil, "Labels for the new prompt version (comma-separated)",
	)
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the prompt (comma-separated)")
	cmd.Flags().
		StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().
		StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	cmd.MarkFlagsMutuallyExclusive("file", "content")

	return cmd
}

func makeGenericDeleteCmd() *cobra.Command {
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
		Short:   "Delete a prompt",
		Long: `Delete a prompt and all its versions, or delete specific versions.

Without flags, deletes the prompt and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.`,
		Example: `  iai prompts delete greeting
  iai prompts delete greeting -f
  iai prompts delete greeting --version 3
  iai prompts delete greeting --label staging`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			// Deleting all versions is destructive; require confirmation.
			if version == 0 && label == "" && !force {
				fmt.Fprintf(
					out,
					"This will delete prompt %q and all its versions. Continue? [y/N] ",
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
			fmt.Fprintf(out, "Deleting prompt %q...\n", name)

			if version > 0 || label != "" {
				err = apiClient.DeletePrompt(
					cmd.Context(),
					pCtx.projectId,
					"", // empty route segment → generic /prompts endpoint
					name,
					version,
					label,
				)
			} else {
				err = apiClient.DeletePromptByName(
					cmd.Context(),
					pCtx.projectId,
					"", // empty route segment → generic /prompts endpoint
					name,
				)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Successfully deleted prompt %q.\n", name)

			return nil
		},
	}

	cmd.Flags().IntVar(&version, "version", 0, "Delete a specific version only")
	cmd.Flags().StringVar(&label, "label", "", "Delete versions with this label only")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().
		StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().
		StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

// resolveContent returns the prompt content from either the --file flag or the
// --content flag. Exactly one must be provided.
func resolveContent(file, content string) (string, error) {
	if file == "" && content == "" {
		return "", fmt.Errorf("either --file or --content must be provided")
	}
	if file != "" && content != "" {
		return "", fmt.Errorf("--file and --content are mutually exclusive")
	}
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("failed to read file %q: %w", file, err)
		}
		return string(data), nil
	}
	return content, nil
}

// validatePromptType checks that the given type is one of the allowed values.
func validatePromptType(promptType string) error {
	for _, valid := range validPromptTypes {
		if promptType == valid {
			return nil
		}
	}
	return fmt.Errorf(
		"invalid prompt type %q: must be one of %s",
		promptType,
		strings.Join(validPromptTypes, ", "),
	)
}

func makeGenericVersionsCmd() *cobra.Command {
	var (
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:     "versions <name>",
		Aliases: []string{"vers"},
		Short:   "List versions of a prompt",
		Long:    `List all versions of a prompt, sorted newest-first.`,
		Example: `  iai prompts versions greeting`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			opts := clients.PromptListOptions{Limit: 1000, Folder: "prompts"}
			result, err := apiClient.ListPrompts(
				cmd.Context(), pCtx.projectId, "", opts,
			)
			if err != nil {
				return err
			}

			var versions []int
			for _, p := range result.Prompts {
				if p.Name == name {
					versions = p.Versions
					break
				}
			}
			if versions == nil {
				return fmt.Errorf("prompt %q not found", name)
			}

			return output.PrintPromptVersions(out, versions)
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the prompts")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")

	return cmd
}

func makeGenericDiffCmd() *cobra.Command {
	var (
		project string
		org     string
	)

	cmd := &cobra.Command{
		Use:     "diff <name> <version_a> <version_b>",
		Short:   "Compare two versions of a prompt",
		Long:    `Show the differences between two versions of a prompt.`,
		Example: `  iai prompts diff greeting 1 3`,
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
				cmd.Context(), pCtx.projectId, "", name, versionA, "",
			)
			if err != nil {
				return err
			}

			b, err := apiClient.GetPrompt(
				cmd.Context(), pCtx.projectId, "", name, versionB, "",
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
