package cmd

import (
	"context"
	"fmt"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	serviceProject         string
	serviceOrganization    string
	serviceName            string
	servicePort            int
	serviceImageType       string
	serviceImageRepository string
	serviceImageName       string
	serviceImageTag        string
	serviceReplicas        int
	serviceReqMemory       string
	serviceReqCPU          string
	serviceLimitMemory     string
	serviceLimitCPU        string

	serviceEndpoint   bool
	serviceEnvVars    []string
	serviceSecretRefs []string
)

var servicesCmd = &cobra.Command{
	Use:     "services",
	Aliases: []string{"service"},
	Short:   "Manage services",
	Long:    `Manage deployment of services to InteractiveAI projects.`,
}

var servCCmd = &cobra.Command{
	Use:     "create [service_name]",
	Aliases: []string{"new"},
	Short:   "Create a service in a project",
	Long: `Create a service in a specific project using the deployment service.

All configuration is provided via flags. The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 {
			serviceName = args[0]
		}

		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide it as a positional argument")
		}

		if servicePort <= 0 {
			return fmt.Errorf("service port must be greater than zero; please provide --service-port")
		}
		if serviceImageName == "" {
			return fmt.Errorf("image name is required; please provide --image-name")
		}
		if serviceImageTag == "" {
			return fmt.Errorf("image tag is required; please provide --image-tag")
		}
		if serviceImageType == "" {
			return fmt.Errorf("image type is required; please provide --image-type")
		}
		if serviceImageType != "internal" && serviceImageType != "external" {
			return fmt.Errorf("image type must be either 'internal' or 'external'; please provide --image-type")
		}
		if serviceImageType == "external" && serviceImageRepository == "" {
			return fmt.Errorf("image repository is required for external images; please provide --image-repository")
		}

		var cfg *files.StackConfig
		if cfgFilePath != "" {
			loadedCfg, err := files.LoadStackConfig(cfgFilePath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
			cfg = loadedCfg
		} else {
			cfg = &files.StackConfig{}
		}

		// Ensure the user is logged in and load session cookies.
		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, serviceOrganization, selectedOrg)
		if err != nil {
			return err
		}

		projectName, err := files.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		// Build env vars from repeated --env flags (NAME=VALUE).
		var env []clients.EnvVar
		for _, e := range serviceEnvVars {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				return fmt.Errorf("invalid --env value %q; expected NAME=VALUE", e)
			}
			env = append(env, clients.EnvVar{
				Name:  strings.TrimSpace(parts[0]),
				Value: parts[1],
			})
		}

		// Build secret references from repeated --secret flags (secret names).
		var secretRefs []clients.SecretRef
		for _, name := range serviceSecretRefs {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				return fmt.Errorf("invalid --secret value %q; name must not be empty", name)
			}
			secretRefs = append(secretRefs, clients.SecretRef{SecretName: trimmed})
		}

		reqBody := clients.CreateServiceBody{
			ServicePort: servicePort,
			Image: clients.ImageSpec{
				Type:       serviceImageType,
				Repository: serviceImageRepository,
				Name:       serviceImageName,
				Tag:        serviceImageTag,
			},
			Resources: clients.Resources{
				Requests: clients.ResourceRequirements{
					Memory: serviceReqMemory,
					CPU:    serviceReqCPU,
				},
				Limits: clients.ResourceRequirements{
					Memory: serviceLimitMemory,
					CPU:    serviceLimitCPU,
				},
			},
			Env:        env,
			SecretRefs: secretRefs,
			Endpoint:   serviceEndpoint,
			Replicas:   serviceReplicas,
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service creation request...")

		serverMessage, err := deployClient.CreateService(cmd.Context(), orgId, projectId, serviceName, reqBody)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var servUCmd = &cobra.Command{
	Use:   "update [service_name]",
	Short: "Update a service in a project",
	Long: `Update a service in a specific project using the deployment service.

All configuration is provided via flags. The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 {
			serviceName = args[0]
		}

		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide [service_name] as the first positional argument")
		}

		if servicePort <= 0 {
			return fmt.Errorf("service port must be greater than zero; please provide --service-port")
		}
		if serviceImageName == "" {
			return fmt.Errorf("image name is required; please provide --image-name")
		}
		if serviceImageTag == "" {
			return fmt.Errorf("image tag is required; please provide --image-tag")
		}
		if serviceImageType == "" {
			return fmt.Errorf("image type is required; please provide --image-type")
		}
		if serviceImageType != "internal" && serviceImageType != "external" {
			return fmt.Errorf("image type must be either 'internal' or 'external'; please provide --image-type")
		}
		if serviceImageType == "external" && serviceImageRepository == "" {
			return fmt.Errorf("image repository is required for external images; please provide --image-repository")
		}

		var cfg *files.StackConfig
		if cfgFilePath != "" {
			loadedCfg, err := files.LoadStackConfig(cfgFilePath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
			cfg = loadedCfg
		} else {
			cfg = &files.StackConfig{}
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, serviceOrganization, selectedOrg)
		if err != nil {
			return err
		}

		projectName, err := files.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		// Build env vars from repeated --env flags (NAME=VALUE).
		var env []clients.EnvVar
		for _, e := range serviceEnvVars {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				return fmt.Errorf("invalid --env value %q; expected NAME=VALUE", e)
			}
			env = append(env, clients.EnvVar{
				Name:  strings.TrimSpace(parts[0]),
				Value: parts[1],
			})
		}

		// Build secret references from repeated --secret flags (secret names).
		var secretRefs []clients.SecretRef
		for _, name := range serviceSecretRefs {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				return fmt.Errorf("invalid --secret value %q; name must not be empty", name)
			}
			secretRefs = append(secretRefs, clients.SecretRef{SecretName: trimmed})
		}

		reqBody := clients.CreateServiceBody{
			ServicePort: servicePort,
			Image: clients.ImageSpec{
				Type:       serviceImageType,
				Repository: serviceImageRepository,
				Name:       serviceImageName,
				Tag:        serviceImageTag,
			},
			Resources: clients.Resources{
				Requests: clients.ResourceRequirements{
					Memory: serviceReqMemory,
					CPU:    serviceReqCPU,
				},
				Limits: clients.ResourceRequirements{
					Memory: serviceLimitMemory,
					CPU:    serviceLimitCPU,
				},
			},
			Env:        env,
			SecretRefs: secretRefs,
			Endpoint:   serviceEndpoint,
			Replicas:   serviceReplicas,
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service update request...")

		serverMessage, err := deployClient.UpdateService(cmd.Context(), orgId, projectId, serviceName, reqBody)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

type ServiceReplica struct {
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	Status    string `json:"status"`
	Ready     bool   `json:"ready"`
	StartTime string `json:"startTime,omitempty"`
	CPU       string `json:"cpu,omitempty"`
	Memory    string `json:"memory,omitempty"`
}

type ListServiceReplicasResponse struct {
	ServiceName string           `json:"serviceName"`
	ProjectId   string           `json:"projectId"`
	Replicas    []ServiceReplica `json:"replicas"`
}

var servListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List services in a project",
	Long: `List services in a specific project using the deployment service.

The project is selected with --project.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var cfg *files.StackConfig
		if cfgFilePath != "" {
			loadedCfg, err := files.LoadStackConfig(cfgFilePath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
			cfg = loadedCfg
		} else {
			cfg = &files.StackConfig{}
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, serviceOrganization, selectedOrg)
		if err != nil {
			return err
		}

		projectName, err := files.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		services, err := deployClient.ListServices(cmd.Context(), orgId, projectId, "")
		if err != nil {
			return err
		}

		headers := []string{"NAME", "REVISION", "STATUS", "ENDPOINT", "UPDATED"}
		rows := make([][]string, len(services))
		for i, svc := range services {
			rows[i] = []string{
				svc.Name,
				fmt.Sprintf("%d", svc.Revision),
				svc.Status,
				svc.Endpoint,
				svc.Updated,
			}
		}

		if err := output.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var servDCmd = &cobra.Command{
	Use:   "delete [service_name]",
	Short: "Delete a service from a project",
	Long: `Delete a service from a specific project using the deployment service.

The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var serviceName string
		if len(args) > 0 {
			serviceName = args[0]
		}

		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide the service name as an argument")
		}

		var cfg *files.StackConfig
		if cfgFilePath != "" {
			loadedCfg, err := files.LoadStackConfig(cfgFilePath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
			cfg = loadedCfg
		} else {
			cfg = &files.StackConfig{}
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, serviceOrganization, selectedOrg)
		if err != nil {
			return err
		}

		projectName, err := files.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service deletion request...")

		serverMessage, err := deployClient.DeleteService(cmd.Context(), orgId, projectId, serviceName)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var servRestartCmd = &cobra.Command{
	Use:   "restart [service_name]",
	Short: "Restart a service in a project",
	Long: `Restart a service in a specific project using the deployment service.

The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var serviceName string
		if len(args) > 0 {
			serviceName = args[0]
		}

		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide the service name as an argument")
		}

		cfg := &files.StackConfig{}
		var err error
		if cfgFilePath != "" {
			cfg, err = files.LoadStackConfig(cfgFilePath)
		}
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, serviceOrganization, selectedOrg)
		if err != nil {
			return err
		}

		projectName, err := files.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service restart request...")

		serverMessage, err := deployClient.RestartService(cmd.Context(), orgId, projectId, serviceName)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var (
	syncProject      string
	syncOrganization string
)

var servicesSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync services in a project from a stack config file",
	Long: `Sync services in a specific project from a stack configuration file.

The sync command will:
- Create services that exist in the config but not in the project
- Update services that exist in both the config and the project
- Delete services that exist in the project but not in the config (for the specified stack)

The project is selected with --project and the config file with --cfg-file.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if cfgFilePath == "" {
			return fmt.Errorf("config file is required; please provide --cfg-file")
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load stack config: %w", err)
		}

		if len(cfg.Services) == 0 {
			return fmt.Errorf("services are required in config file for sync command")
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, syncOrganization, selectedOrg)
		if err != nil {
			return err
		}

		projectName, err := files.ResolveProject(cfg.Project, syncProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprint(out, "Syncing services")
		done := output.PrintLoadingDots(out)

		result, err := SyncServices(cmd.Context(), apiClient, deployClient, orgName, projectName, cfg)
		fmt.Fprintln(out)
		close(done)
		if err != nil {
			return err
		}

		if len(result.Created) > 0 {
			fmt.Fprintf(out, "Created services: %s\n", strings.Join(result.Created, ", "))
		}
		if len(result.Updated) > 0 {
			fmt.Fprintf(out, "Updated services: %s\n", strings.Join(result.Updated, ", "))
		}
		if len(result.Deleted) > 0 {
			fmt.Fprintf(out, "Deleted services: %s\n", strings.Join(result.Deleted, ", "))
		}
		if len(result.Created) == 0 && len(result.Updated) == 0 && len(result.Deleted) == 0 {
			fmt.Fprintln(out, "No changes required; services already match config.")
		}

		return nil
	},
}

type SyncResult struct {
	Created []string
	Updated []string
	Deleted []string
}

func SyncServices(
	ctx context.Context,
	apiClient *clients.APIClient,
	deployClient *clients.DeploymentClient,
	orgName,
	projectName string,
	cfg *files.StackConfig,
) (*SyncResult, error) {
	orgId, projectId, err := apiClient.GetProjectId(ctx, orgName, projectName)
	if err != nil {
		return nil, err
	}

	existing, err := deployClient.ListServices(ctx, orgId, projectId, cfg.StackId)
	if err != nil {
		return nil, err
	}

	existingByName := make(map[string]clients.ServiceOutput)
	for _, svc := range existing {
		existingByName[svc.Name] = svc
	}

	result := &SyncResult{
		Created: []string{},
		Updated: []string{},
		Deleted: []string{},
	}

	for name, svcCfg := range cfg.Services {
		req := svcCfg.ToCreateRequest(cfg.StackId)

		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateService(ctx, orgId, projectId, name, req)
			if err != nil {
				return nil, err
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := deployClient.UpdateService(ctx, orgId, projectId, name, req)
			if err != nil {
				return nil, err
			}
			result.Updated = append(result.Updated, name)
		}
	}

	for name := range existingByName {
		if _, desired := cfg.Services[name]; !desired {
			_, err := deployClient.DeleteService(ctx, orgId, projectId, name)
			if err != nil {
				return nil, err
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}

func init() {

	// Flags for "services create"
	servCCmd.Flags().StringVarP(&serviceProject, "project", "p", "", "Project name to create the service in")
	servCCmd.Flags().StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")
	servCCmd.Flags().IntVar(&servicePort, "port", 0, "Service port to expose")
	servCCmd.Flags().StringVar(&serviceImageType, "image-type", "", "Image type: internal or external")
	servCCmd.Flags().StringVar(&serviceImageRepository, "image-repository", "", "Container image repository (external images only)")
	servCCmd.Flags().StringVar(&serviceImageName, "image-name", "", "Container image name")
	servCCmd.Flags().StringVar(&serviceImageTag, "image-tag", "", "Container image tag")

	servCCmd.Flags().IntVar(&serviceReplicas, "replicas", 1, "Number of replicas for the service")
	servCCmd.Flags().StringVar(&serviceReqMemory, "requests-memory", "512Mi", "Requested memory (e.g. 512Mi)")
	servCCmd.Flags().StringVar(&serviceReqCPU, "requests-cpu", "250m", "Requested CPU (e.g. 250m)")
	servCCmd.Flags().StringVar(&serviceLimitMemory, "limits-memory", "1Gi", "Memory limit (e.g. 1Gi)")
	servCCmd.Flags().StringVar(&serviceLimitCPU, "limits-cpu", "500m", "CPU limit (e.g. 500m)")
	servCCmd.Flags().StringArrayVar(&serviceEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	servCCmd.Flags().StringArrayVar(&serviceSecretRefs, "secret", nil, "Secrets to be loaded as env vars; can be repeated")
	servCCmd.Flags().BoolVar(&serviceEndpoint, "endpoint", false, "Expose the service at <service-name>-<project-hash>.dev.interactive.ai")

	// Flags for "services update"
	servUCmd.Flags().StringVarP(&serviceProject, "project", "p", "", "Project name to update the service in")
	servUCmd.Flags().StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")
	servUCmd.Flags().IntVar(&servicePort, "port", 0, "Service port to expose")
	servUCmd.Flags().StringVar(&serviceImageType, "image-type", "", "Image type: internal or external")
	servUCmd.Flags().StringVar(&serviceImageRepository, "image-repository", "", "Container image repository (external images only)")
	servUCmd.Flags().StringVar(&serviceImageName, "image-name", "", "Container image name")
	servUCmd.Flags().StringVar(&serviceImageTag, "image-tag", "", "Container image tag")

	servUCmd.Flags().IntVar(&serviceReplicas, "replicas", 1, "Number of replicas for the service")
	servUCmd.Flags().StringVar(&serviceReqMemory, "requests-memory", "512Mi", "Requested memory (e.g. 512Mi)")
	servUCmd.Flags().StringVar(&serviceReqCPU, "requests-cpu", "250m", "Requested CPU (e.g. 250m)")
	servUCmd.Flags().StringVar(&serviceLimitMemory, "limits-memory", "1Gi", "Memory limit (e.g. 1Gi)")
	servUCmd.Flags().StringVar(&serviceLimitCPU, "limits-cpu", "500m", "CPU limit (e.g. 500m)")
	servUCmd.Flags().StringArrayVar(&serviceEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	servUCmd.Flags().StringArrayVar(&serviceSecretRefs, "secret", nil, "Secrets to be loaded as env vars; can be repeated")
	servUCmd.Flags().BoolVar(&serviceEndpoint, "endpoint", false, "Expose the service at <service-name>-<project-hash>.dev.interactive.ai")

	// Flags for "services list"
	servListCmd.Flags().StringVarP(&serviceProject, "project", "p", "", "Project name to list services from")
	servListCmd.Flags().StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "services delete"
	servDCmd.Flags().StringVarP(&serviceProject, "project", "p", "", "Project name to delete the service from")
	servDCmd.Flags().StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "services restart"
	servRestartCmd.Flags().StringVarP(&serviceProject, "project", "p", "", "Project name to restart the service in")
	servRestartCmd.Flags().StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "services sync"
	servicesSyncCmd.Flags().StringVarP(&syncProject, "project", "p", "", "Project name to sync services in")
	servicesSyncCmd.Flags().StringVarP(&syncOrganization, "organization", "o", "", "Organization name that owns the project")

	// Register commands
	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(servCCmd)
	servicesCmd.AddCommand(servUCmd)
	servicesCmd.AddCommand(servListCmd)
	servicesCmd.AddCommand(servDCmd)
	servicesCmd.AddCommand(servRestartCmd)
	servicesCmd.AddCommand(servicesSyncCmd)
}
