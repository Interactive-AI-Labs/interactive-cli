package cmd

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	secretsProject      string
	secretsOrganization string
	secretName          string
	secretDataKVs       []string
	secretEnvFile       string
	secretReplaceFlag   bool
	secretRemoveKeys    []string
	secretsListJSON     bool
	secretsListYAML     bool
	secretsGetJSON      bool
	secretsGetYAML      bool
)

var secretsCmd = &cobra.Command{
	Use:     "secrets",
	Aliases: []string{"secret"},
	Short:   "Encrypted key-value pairs for services and agents",
	GroupID: groupInfra,
	Long:    `Manage secrets in InteractiveAI projects.`,
}

var secretsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List secrets in a project",
	Long:    `List secrets in a specific project.`,
	Example: `  iai secrets list
  iai secrets list -p my-project -o my-org
  iai secrets list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			secretsOrganization,
			secretsProject,
		)
		if err != nil {
			return err
		}

		secrets, err := deployClient.ListSecrets(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}

		if secretsListJSON {
			return output.PrintStructuredJSON(out, secrets)
		}
		if secretsListYAML {
			return output.PrintStructuredYAML(out, secrets)
		}

		return output.PrintSecretList(out, secrets)
	},
}

var secretsCreateCmd = &cobra.Command{
	Use:   "create [secret_name]",
	Short: "Create a secret in a project",
	Long: `Create a secret in a specific project using the deployment service.

Secret data can be provided via:
  --data KEY=VALUE         (can be repeated)
  --from-env-file FILE     (KEY=VALUE pairs, one per line)

When both are provided, --data values take precedence.`,
	Example: `  iai secrets create my-secret -d API_KEY=abc123
  iai secrets create my-secret -d API_KEY=abc123 -d DB_PASS=secret
  iai secrets create my-secret --from-env-file .env
  iai secrets create my-secret --from-env-file .env -d API_KEY=override -p my-project`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 && strings.TrimSpace(secretName) == "" {
			secretName = args[0]
		}

		if strings.TrimSpace(secretName) == "" {
			return fmt.Errorf(
				"secret name is required; please provide --secret-name or positional argument",
			)
		}
		if len(secretDataKVs) == 0 && strings.TrimSpace(secretEnvFile) == "" {
			return fmt.Errorf("at least one --data KEY=VALUE pair or --from-env-file is required")
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			secretsOrganization,
			secretsProject,
		)
		if err != nil {
			return err
		}

		data, err := mergeSecretData(secretDataKVs, secretEnvFile)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting secret creation request...")

		serverMessage, err := deployClient.CreateSecret(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			secretName,
			data,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var secretsUpdateCmd = &cobra.Command{
	Use:   "update <secret_name>",
	Short: "Update keys in a secret",
	Long: `Update one or more keys in an existing secret.

By default, only the specified keys are updated (merge/upsert). Existing keys
not included in the update are preserved.

With --replace, ALL secret data is replaced. Any keys not included in the new
data will be permanently deleted.

With --remove, the specified keys are deleted from the secret. Cannot be
combined with --data, --from-env-file, or --replace.

Secret data can be provided via:
  --data KEY=VALUE         (can be repeated)
  --from-env-file FILE     (KEY=VALUE pairs, one per line)

When both are provided, --data values take precedence.`,
	Example: `  # Update a single key (other keys preserved)
  iai secrets update my-secret -d API_KEY=new-value

  # Update multiple keys (other keys preserved)
  iai secrets update my-secret -d API_KEY=val1 -d DB_PASS=val2

  # Replace all keys (keys not provided will be deleted)
  iai secrets update my-secret -d API_KEY=val1 --replace

  # Remove specific keys from a secret
  iai secrets update my-secret --remove API_KEY

  # Remove multiple keys
  iai secrets update my-secret --remove KEY1 --remove KEY2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretName := strings.TrimSpace(args[0])
		if secretName == "" {
			return fmt.Errorf("secret name is required")
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			secretsOrganization,
			secretsProject,
		)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)

		if len(secretRemoveKeys) > 0 {
			fmt.Fprintln(out, "Submitting secret key remove request...")

			sort.Strings(secretRemoveKeys)

			var removedKeys []string
			for _, keyName := range secretRemoveKeys {
				keyName = strings.TrimSpace(keyName)
				_, err := deployClient.DeleteSecretKey(
					cmd.Context(),
					pCtx.orgId,
					pCtx.projectId,
					secretName,
					keyName,
				)
				if err != nil && len(removedKeys) > 0 {
					fmt.Fprintf(
						out,
						"Partial failure: removed keys %s before error\n",
						strings.Join(removedKeys, ", "),
					)
				}
				if err != nil {
					return fmt.Errorf("failed to remove key %q: %w", keyName, err)
				}
				removedKeys = append(removedKeys, keyName)
			}
			fmt.Fprintf(out, "Success: removed keys %s\n", strings.Join(removedKeys, ", "))
			return nil
		}

		data, err := mergeSecretData(secretDataKVs, secretEnvFile)
		if err != nil {
			return err
		}

		if secretReplaceFlag {
			fmt.Fprintln(out, "Submitting secret replace request...")

			serverMessage, err := deployClient.ReplaceSecret(
				cmd.Context(),
				pCtx.orgId,
				pCtx.projectId,
				secretName,
				data,
			)
			if err != nil {
				return err
			}

			if serverMessage != "" {
				fmt.Fprintln(out, serverMessage)
			}
			return nil
		}

		fmt.Fprintln(out, "Submitting secret update request...")

		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var updatedKeys []string
		for _, keyName := range keys {
			serverMessage, err := deployClient.UpdateSecretKey(
				cmd.Context(),
				pCtx.orgId,
				pCtx.projectId,
				secretName,
				keyName,
				data[keyName],
			)
			if err != nil && len(updatedKeys) > 0 {
				fmt.Fprintf(
					out,
					"Partial failure: updated keys %s before error\n",
					strings.Join(updatedKeys, ", "),
				)
			}
			if err != nil {
				return fmt.Errorf("failed to update key %q: %w", keyName, err)
			}
			updatedKeys = append(updatedKeys, keyName)

			if serverMessage != "" {
				fmt.Fprintln(out, serverMessage)
			}
		}

		return nil
	},
}

var secretsDeleteCmd = &cobra.Command{
	Use:     "delete <secret_name>",
	Aliases: []string{"rm"},
	Short:   "Delete a secret in a project",
	Long:    `Delete a secret in a specific project using the deployment service.`,
	Example: `  iai secrets delete my-secret
  iai secrets delete my-secret -p my-project -o my-org`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretToDelete := strings.TrimSpace(args[0])
		if secretToDelete == "" {
			return fmt.Errorf("secret name is required")
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			secretsOrganization,
			secretsProject,
		)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting secret delete request...")

		serverMessage, err := deployClient.DeleteSecret(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			secretToDelete,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var secretsGetCmd = &cobra.Command{
	Use:   "get <secret_name>",
	Short: "Get a secret in a project",
	Long:  `Get a secret in a specific project using the deployment service.`,
	Example: `  iai secrets get my-secret
  iai secrets get my-secret -p my-project
  iai secrets get my-secret --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretName := strings.TrimSpace(args[0])
		if secretName == "" {
			return fmt.Errorf("secret name is required")
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			secretsOrganization,
			secretsProject,
		)
		if err != nil {
			return err
		}

		secret, err := deployClient.GetSecret(cmd.Context(), pCtx.orgId, pCtx.projectId, secretName)
		if err != nil {
			return fmt.Errorf("failed to get secret %q: %w", secretName, err)
		}

		if secretsGetJSON {
			return output.PrintStructuredJSON(out, secret)
		}
		if secretsGetYAML {
			return output.PrintStructuredYAML(out, secret)
		}

		return output.PrintSecretData(out, secret.Data)
	},
}

// parseKeyValuePairs parses KEY=VALUE strings and validates each key and value.
func parseKeyValuePairs(pairs []string) (map[string]string, error) {
	data := make(map[string]string, len(pairs))

	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --data value %q; expected KEY=VALUE", p)
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("invalid --data value %q; key must not be empty", p)
		}
		if err := inputs.ValidateSecretKey(key); err != nil {
			return nil, err
		}

		value := parts[1]
		if err := inputs.ValidateSecretValue(key, value); err != nil {
			return nil, err
		}
		data[key] = value
	}

	return data, nil
}

// mergeSecretData merges secret data from --data flags and/or an env file.
// When both sources are provided, --data values take precedence over env file values.
func mergeSecretData(pairs []string, envFilePath string) (map[string]string, error) {
	data := make(map[string]string)

	if strings.TrimSpace(envFilePath) != "" {
		envData, err := files.ParseEnvFile(envFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
		maps.Copy(data, envData)
	}

	if len(pairs) > 0 {
		pairData, err := parseKeyValuePairs(pairs)
		if err != nil {
			return nil, err
		}
		maps.Copy(data, pairData)
	}

	return data, nil
}

func init() {
	// secrets list
	secretsListCmd.Flags().
		StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsListCmd.Flags().
		StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsListCmd.Flags().
		BoolVar(&secretsListJSON, "json", false, "Output raw API response as JSON")
	secretsListCmd.Flags().
		BoolVar(&secretsListYAML, "yaml", false, "Output raw API response as YAML")
	secretsListCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// secrets create
	secretsCreateCmd.Flags().
		StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsCreateCmd.Flags().
		StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsCreateCmd.Flags().StringVarP(&secretName, "secret-name", "s", "", "Name of the secret")
	secretsCreateCmd.Flags().
		StringArrayVarP(&secretDataKVs, "data", "d", nil, "Secret data in KEY=VALUE form (repeatable)")
	secretsCreateCmd.Flags().
		StringVar(&secretEnvFile, "from-env-file", "", "Path to env file with KEY=VALUE pairs (one per line)")

	// secrets update
	secretsUpdateCmd.Flags().
		StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsUpdateCmd.Flags().
		StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsUpdateCmd.Flags().
		StringArrayVarP(&secretDataKVs, "data", "d", nil, "Secret data in KEY=VALUE form (repeatable)")
	secretsUpdateCmd.Flags().
		StringVar(&secretEnvFile, "from-env-file", "", "Path to env file with KEY=VALUE pairs (one per line)")
	secretsUpdateCmd.Flags().
		BoolVar(&secretReplaceFlag, "replace", false, "Replace all secret data (keys not provided will be deleted)")
	secretsUpdateCmd.Flags().
		StringArrayVar(&secretRemoveKeys, "remove", nil, "Key name to remove from the secret (repeatable)")
	secretsUpdateCmd.MarkFlagsOneRequired("data", "from-env-file", "remove")
	secretsUpdateCmd.MarkFlagsMutuallyExclusive("remove", "data")
	secretsUpdateCmd.MarkFlagsMutuallyExclusive("remove", "from-env-file")
	secretsUpdateCmd.MarkFlagsMutuallyExclusive("remove", "replace")

	// secrets delete
	secretsDeleteCmd.Flags().
		StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsDeleteCmd.Flags().
		StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	// secrets get
	secretsGetCmd.Flags().
		StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsGetCmd.Flags().
		StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsGetCmd.Flags().BoolVar(&secretsGetJSON, "json", false, "Output raw API response as JSON")
	secretsGetCmd.Flags().BoolVar(&secretsGetYAML, "yaml", false, "Output raw API response as YAML")
	secretsGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Wire up the command hierarchy
	secretsCmd.AddCommand(
		secretsListCmd,
		secretsCreateCmd,
		secretsUpdateCmd,
		secretsDeleteCmd,
		secretsGetCmd,
	)
	rootCmd.AddCommand(secretsCmd)
}
