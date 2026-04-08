package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/sync"
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
	serviceMemory          string
	serviceCPU             string

	serviceAutoscalingEnabled bool
	serviceAutoscalingMin     int
	serviceAutoscalingMax     int
	serviceAutoscalingCPU     int
	serviceAutoscalingMemory  int

	serviceEndpoint   bool
	serviceEnvVars    []string
	serviceSecretRefs []string

	serviceHealthcheckPath         string
	serviceHealthcheckInitialDelay int
)

func isHealthcheckEnabled() bool {
	return serviceHealthcheckPath != "" || serviceHealthcheckInitialDelay != 0
}

var (
	serviceScheduleUptime   string
	serviceScheduleDowntime string
	serviceScheduleTimezone string
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
`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 {
			serviceName = args[0]
		}

		input := inputs.ServiceInput{
			Name:            serviceName,
			Port:            servicePort,
			ImageType:       serviceImageType,
			ImageRepository: serviceImageRepository,
			ImageName:       serviceImageName,
			ImageTag:        serviceImageTag,
			Memory:          serviceMemory,
			CPU:             serviceCPU,
			Replicas:        serviceReplicas,
		}

		if serviceAutoscalingEnabled {
			input.Autoscaling = &inputs.AutoscalingInput{
				Enabled:       true,
				MinReplicas:   serviceAutoscalingMin,
				MaxReplicas:   serviceAutoscalingMax,
				CPUPercentage: serviceAutoscalingCPU,
				MemoryPercent: serviceAutoscalingMemory,
			}
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, serviceOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		if err := inputs.ValidateServiceEnvVars(serviceEnvVars); err != nil {
			return err
		}
		var env []clients.EnvVar
		for _, e := range serviceEnvVars {
			parts := strings.SplitN(e, "=", 2)
			env = append(env, clients.EnvVar{
				Name:  strings.TrimSpace(parts[0]),
				Value: parts[1],
			})
		}

		if err := inputs.ValidateServiceSecretRefs(serviceSecretRefs); err != nil {
			return err
		}
		var secretRefs []clients.SecretRef
		for _, name := range serviceSecretRefs {
			secretRefs = append(secretRefs, clients.SecretRef{
				SecretName: strings.TrimSpace(name),
			})
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
				Memory: serviceMemory,
				CPU:    serviceCPU,
			},
			Env:        env,
			SecretRefs: secretRefs,
			Endpoint:   serviceEndpoint,
		}

		if serviceAutoscalingEnabled {
			reqBody.Autoscaling = &clients.Autoscaling{
				Enabled:          true,
				MinReplicas:      serviceAutoscalingMin,
				MaxReplicas:      serviceAutoscalingMax,
				CPUPercentage:    serviceAutoscalingCPU,
				MemoryPercentage: serviceAutoscalingMemory,
			}
		} else if serviceReplicas > 0 {
			reqBody.Replicas = serviceReplicas
		}

		if isHealthcheckEnabled() {
			reqBody.Healthcheck = &clients.Healthcheck{
				Path:                serviceHealthcheckPath,
				InitialDelaySeconds: serviceHealthcheckInitialDelay,
			}
		}

		if serviceScheduleUptime != "" || serviceScheduleDowntime != "" ||
			serviceScheduleTimezone != "" {
			reqBody.Schedule = &clients.Schedule{
				Uptime:   serviceScheduleUptime,
				Downtime: serviceScheduleDowntime,
				Timezone: serviceScheduleTimezone,
			}
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service creation request...")

		serverMessage, err := deployClient.CreateService(
			cmd.Context(),
			orgId,
			projectId,
			serviceName,
			reqBody,
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

var servUCmd = &cobra.Command{
	Use:   "update [service_name]",
	Short: "Update a service in a project",
	Long: `Update a service in a specific project using the deployment service.
`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 {
			serviceName = args[0]
		}

		input := inputs.ServiceInput{
			Name:            serviceName,
			Port:            servicePort,
			ImageType:       serviceImageType,
			ImageRepository: serviceImageRepository,
			ImageName:       serviceImageName,
			ImageTag:        serviceImageTag,
			Memory:          serviceMemory,
			CPU:             serviceCPU,
			Replicas:        serviceReplicas,
		}

		if serviceAutoscalingEnabled {
			input.Autoscaling = &inputs.AutoscalingInput{
				Enabled:       true,
				MinReplicas:   serviceAutoscalingMin,
				MaxReplicas:   serviceAutoscalingMax,
				CPUPercentage: serviceAutoscalingCPU,
				MemoryPercent: serviceAutoscalingMemory,
			}
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, serviceOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		if err := inputs.ValidateServiceEnvVars(serviceEnvVars); err != nil {
			return err
		}
		var env []clients.EnvVar
		for _, e := range serviceEnvVars {
			parts := strings.SplitN(e, "=", 2)
			env = append(env, clients.EnvVar{
				Name:  strings.TrimSpace(parts[0]),
				Value: parts[1],
			})
		}

		if err := inputs.ValidateServiceSecretRefs(serviceSecretRefs); err != nil {
			return err
		}
		var secretRefs []clients.SecretRef
		for _, name := range serviceSecretRefs {
			secretRefs = append(secretRefs, clients.SecretRef{
				SecretName: strings.TrimSpace(name),
			})
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
				Memory: serviceMemory,
				CPU:    serviceCPU,
			},
			Env:        env,
			SecretRefs: secretRefs,
			Endpoint:   serviceEndpoint,
		}

		if serviceAutoscalingEnabled {
			reqBody.Autoscaling = &clients.Autoscaling{
				Enabled:          true,
				MinReplicas:      serviceAutoscalingMin,
				MaxReplicas:      serviceAutoscalingMax,
				CPUPercentage:    serviceAutoscalingCPU,
				MemoryPercentage: serviceAutoscalingMemory,
			}
		} else if serviceReplicas > 0 {
			reqBody.Replicas = serviceReplicas
		}

		if isHealthcheckEnabled() {
			reqBody.Healthcheck = &clients.Healthcheck{
				Path:                serviceHealthcheckPath,
				InitialDelaySeconds: serviceHealthcheckInitialDelay,
			}
		}

		if serviceScheduleUptime != "" || serviceScheduleDowntime != "" ||
			serviceScheduleTimezone != "" {
			reqBody.Schedule = &clients.Schedule{
				Uptime:   serviceScheduleUptime,
				Downtime: serviceScheduleDowntime,
				Timezone: serviceScheduleTimezone,
			}
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service update request...")

		serverMessage, err := deployClient.UpdateService(
			cmd.Context(),
			orgId,
			projectId,
			serviceName,
			reqBody,
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
	Long:    `List services in a specific project using the deployment service.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, serviceOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, serviceProject)
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

		return output.PrintServiceList(out, services)
	},
}

var servDCmd = &cobra.Command{
	Use:   "delete [service_name]",
	Short: "Delete a service from a project",
	Long:  `Delete a service from a specific project using the deployment service.`,
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var serviceName string
		if len(args) > 0 {
			serviceName = args[0]
		}

		if serviceName == "" {
			return fmt.Errorf(
				"service name is required; please provide the service name as an argument",
			)
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, serviceOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service deletion request...")

		serverMessage, err := deployClient.DeleteService(
			cmd.Context(),
			orgId,
			projectId,
			serviceName,
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

var servRestartCmd = &cobra.Command{
	Use:   "restart <service_name>",
	Short: "Restart a service in a project",
	Long:  `Restart a service in a specific project using the deployment service.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		serviceName := args[0]

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, serviceOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service restart request...")

		serverMessage, err := deployClient.RestartService(
			cmd.Context(),
			orgId,
			projectId,
			serviceName,
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

var (
	servLogsFollow    bool
	servLogsSince     string
	servLogsStartTime string
	servLogsEndTime   string
)

var servLogsCmd = &cobra.Command{
	Use:   "logs <service_name>",
	Short: "Show logs for a service",
	Long: `Show logs for all replicas of a service in a project.

Returns up to 5000 log entries in chronological order.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		serviceName := strings.TrimSpace(args[0])
		if serviceName == "" {
			return fmt.Errorf("service name is required")
		}

		ctx := cmd.Context()
		if servLogsFollow {
			var stop func()
			ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		timeout := 1 * time.Minute
		if servLogsFollow {
			timeout = 0
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			timeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, serviceOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, serviceProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		opts := clients.LogsOptions{
			Follow:    servLogsFollow,
			Since:     servLogsSince,
			StartTime: servLogsStartTime,
			EndTime:   servLogsEndTime,
		}

		logsResp, err := deployClient.GetServiceLogs(ctx, orgId, projectId, serviceName, opts)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		meta := output.LogsMeta{Since: logsResp.Since, Truncated: logsResp.Truncated}
		err = output.PrintLogStream(out, logsResp.Body, true, meta)
		if servLogsFollow && ctx.Err() != nil {
			return nil
		}
		return err
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

The project is selected with --project or via 'iai projects select', and the config file with --cfg-file.`,
	Args:       cobra.NoArgs,
	Deprecated: "use 'iai stack sync' instead",
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

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, syncOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, syncProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprint(out, "Syncing services")
		done := output.PrintLoadingDots(out)

		svcBodies := make(map[string]clients.CreateServiceBody)
		for name, svcCfg := range cfg.Services {
			svcBodies[name] = svcCfg.ToCreateRequest(cfg.StackId)
		}

		result, err := sync.Services(
			cmd.Context(),
			deployClient,
			orgId,
			projectId,
			cfg.StackId,
			svcBodies,
		)
		close(done)
		fmt.Fprintln(out)

		return sync.PrintResult(out, "services", result, err)
	},
}

func init() {
	// Flags for "services create"
	servCCmd.Flags().
		StringVarP(&serviceProject, "project", "p", "", "Project name to create the service in")
	servCCmd.Flags().
		StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")
	servCCmd.Flags().IntVar(&servicePort, "port", 0, "Service port to expose")
	servCCmd.Flags().
		StringVar(&serviceImageType, "image-type", "", "Image type: 'internal' (project's private registry), 'external' (any public registry), or 'platform' (Interactive AI registries)")
	servCCmd.Flags().
		StringVar(&serviceImageRepository, "image-repository", "", "Container image repository (external images only)")
	servCCmd.Flags().StringVar(&serviceImageName, "image-name", "", "Container image name")
	servCCmd.Flags().StringVar(&serviceImageTag, "image-tag", "", "Container image tag")

	servCCmd.Flags().
		IntVar(&serviceReplicas, "replicas", 0, "Number of replicas for the service (mutually exclusive with autoscaling)")
	servCCmd.Flags().
		StringVar(&serviceMemory, "memory", "", "Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G)")
	servCCmd.Flags().
		StringVar(&serviceCPU, "cpu", "", "CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)")
	_ = servCCmd.MarkFlagRequired("memory")
	_ = servCCmd.MarkFlagRequired("cpu")

	servCCmd.Flags().
		BoolVar(&serviceAutoscalingEnabled, "autoscaling-enabled", false, "Enable autoscaling (mutually exclusive with replicas)")
	servCCmd.Flags().
		IntVar(&serviceAutoscalingMin, "autoscaling-min-replicas", 0, "Minimum number of replicas when autoscaling is enabled")
	servCCmd.Flags().
		IntVar(&serviceAutoscalingMax, "autoscaling-max-replicas", 0, "Maximum number of replicas when autoscaling is enabled")
	servCCmd.Flags().
		IntVar(&serviceAutoscalingCPU, "autoscaling-cpu-percentage", 0, "CPU percentage threshold for autoscaling")
	servCCmd.Flags().
		IntVar(&serviceAutoscalingMemory, "autoscaling-memory-percentage", 0, "Memory percentage threshold for autoscaling")

	servCCmd.Flags().
		StringArrayVar(&serviceEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	servCCmd.Flags().
		StringArrayVar(&serviceSecretRefs, "secret", nil, "Secrets to be loaded as env vars; can be repeated")
	servCCmd.Flags().
		BoolVar(&serviceEndpoint, "endpoint", false, "Expose the service at <service-name>-<project-hash>.interactive.ai")

	servCCmd.Flags().
		StringVar(&serviceHealthcheckPath, "healthcheck-path", "", "HTTP path for healthcheck endpoint (e.g. /health)")
	servCCmd.Flags().
		IntVar(&serviceHealthcheckInitialDelay, "healthcheck-initial-delay", 0, "Initial delay in seconds before starting healthchecks")

	servCCmd.Flags().
		StringVar(&serviceScheduleUptime, "schedule-uptime", "", "When the service should be running (mutually exclusive with --schedule-downtime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Weekdays: Mon, Tue, Wed, Thu, Fri, Sat, Sun (case-insensitive). Times in 24h format; start: 00:00-23:59, end: 00:00-24:00 (24:00 = end of day). Example: 'Mon-Fri 07:30-20:30' or 'Mon-Fri 08:00-18:00, Sat 10:00-14:00'")
	servCCmd.Flags().
		StringVar(&serviceScheduleDowntime, "schedule-downtime", "", "When the service should be scaled down (mutually exclusive with --schedule-uptime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Weekdays: Mon, Tue, Wed, Thu, Fri, Sat, Sun (case-insensitive). Times in 24h format; start: 00:00-23:59, end: 00:00-24:00 (24:00 = end of day). Example: 'Sat-Sun 00:00-24:00'")
	servCCmd.Flags().
		StringVar(&serviceScheduleTimezone, "schedule-timezone", "", "IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); required with --schedule-uptime or --schedule-downtime")

	// Flags for "services update"
	servUCmd.Flags().
		StringVarP(&serviceProject, "project", "p", "", "Project name to update the service in")
	servUCmd.Flags().
		StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")
	servUCmd.Flags().IntVar(&servicePort, "port", 0, "Service port to expose")
	servUCmd.Flags().
		StringVar(&serviceImageType, "image-type", "", "Image type: 'external' (Docker Hub, ghcr.io) or 'internal' (InteractiveAI private registry)")
	servUCmd.Flags().
		StringVar(&serviceImageRepository, "image-repository", "", "Container image repository (external images only)")
	servUCmd.Flags().StringVar(&serviceImageName, "image-name", "", "Container image name")
	servUCmd.Flags().StringVar(&serviceImageTag, "image-tag", "", "Container image tag")

	servUCmd.Flags().
		IntVar(&serviceReplicas, "replicas", 0, "Number of replicas for the service (mutually exclusive with autoscaling)")
	servUCmd.Flags().
		StringVar(&serviceMemory, "memory", "", "Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G)")
	servUCmd.Flags().
		StringVar(&serviceCPU, "cpu", "", "CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)")
	_ = servUCmd.MarkFlagRequired("memory")
	_ = servUCmd.MarkFlagRequired("cpu")

	servUCmd.Flags().
		BoolVar(&serviceAutoscalingEnabled, "autoscaling-enabled", false, "Enable autoscaling (mutually exclusive with replicas)")
	servUCmd.Flags().
		IntVar(&serviceAutoscalingMin, "autoscaling-min-replicas", 0, "Minimum number of replicas when autoscaling is enabled")
	servUCmd.Flags().
		IntVar(&serviceAutoscalingMax, "autoscaling-max-replicas", 0, "Maximum number of replicas when autoscaling is enabled")
	servUCmd.Flags().
		IntVar(&serviceAutoscalingCPU, "autoscaling-cpu-percentage", 0, "CPU percentage threshold for autoscaling")
	servUCmd.Flags().
		IntVar(&serviceAutoscalingMemory, "autoscaling-memory-percentage", 0, "Memory percentage threshold for autoscaling")

	servUCmd.Flags().
		StringArrayVar(&serviceEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	servUCmd.Flags().
		StringArrayVar(&serviceSecretRefs, "secret", nil, "Secrets to be loaded as env vars; can be repeated")
	servUCmd.Flags().
		BoolVar(&serviceEndpoint, "endpoint", false, "Expose the service at <service-name>-<project-hash>.interactive.ai")

	servUCmd.Flags().
		StringVar(&serviceHealthcheckPath, "healthcheck-path", "", "HTTP path for healthcheck endpoint (e.g. /health)")
	servUCmd.Flags().
		IntVar(&serviceHealthcheckInitialDelay, "healthcheck-initial-delay", 0, "Initial delay in seconds before starting healthchecks")

	servUCmd.Flags().
		StringVar(&serviceScheduleUptime, "schedule-uptime", "", "When the service should be running (mutually exclusive with --schedule-downtime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Weekdays: Mon, Tue, Wed, Thu, Fri, Sat, Sun (case-insensitive). Times in 24h format; start: 00:00-23:59, end: 00:00-24:00 (24:00 = end of day). Example: 'Mon-Fri 07:30-20:30' or 'Mon-Fri 08:00-18:00, Sat 10:00-14:00'")
	servUCmd.Flags().
		StringVar(&serviceScheduleDowntime, "schedule-downtime", "", "When the service should be scaled down (mutually exclusive with --schedule-uptime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Weekdays: Mon, Tue, Wed, Thu, Fri, Sat, Sun (case-insensitive). Times in 24h format; start: 00:00-23:59, end: 00:00-24:00 (24:00 = end of day). Example: 'Sat-Sun 00:00-24:00'")
	servUCmd.Flags().
		StringVar(&serviceScheduleTimezone, "schedule-timezone", "", "IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); required with --schedule-uptime or --schedule-downtime")

	// Flags for "services list"
	servListCmd.Flags().
		StringVarP(&serviceProject, "project", "p", "", "Project name to list services from")
	servListCmd.Flags().
		StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "services delete"
	servDCmd.Flags().
		StringVarP(&serviceProject, "project", "p", "", "Project name to delete the service from")
	servDCmd.Flags().
		StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "services restart"
	servRestartCmd.Flags().
		StringVarP(&serviceProject, "project", "p", "", "Project name to restart the service in")
	servRestartCmd.Flags().
		StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")

	// Flags for "services logs"
	servLogsCmd.Flags().
		StringVarP(&serviceProject, "project", "p", "", "Project name that owns the service")
	servLogsCmd.Flags().
		StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")
	servLogsCmd.Flags().
		BoolVarP(&servLogsFollow, "follow", "f", false, "Stream new log entries as they arrive; mutually exclusive with --end-time")
	servLogsCmd.Flags().
		StringVar(&servLogsSince, "since", "", "Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time")
	servLogsCmd.Flags().
		StringVar(&servLogsStartTime, "start-time", "", "Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window")
	servLogsCmd.Flags().
		StringVar(&servLogsEndTime, "end-time", "", "Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow")

	// Flags for "services sync"
	servicesSyncCmd.Flags().
		StringVarP(&syncProject, "project", "p", "", "Project name to sync services in")
	servicesSyncCmd.Flags().
		StringVarP(&syncOrganization, "organization", "o", "", "Organization name that owns the project")

	// Register commands
	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(servCCmd)
	servicesCmd.AddCommand(servUCmd)
	servicesCmd.AddCommand(servListCmd)
	servicesCmd.AddCommand(servDCmd)
	servicesCmd.AddCommand(servRestartCmd)
	servicesCmd.AddCommand(servLogsCmd)
	servicesCmd.AddCommand(servicesSyncCmd)
}
