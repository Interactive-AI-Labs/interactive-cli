package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

type ResourceRequirements struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

type Resources struct {
	Requests ResourceRequirements `json:"requests"`
	Limits   ResourceRequirements `json:"limits"`
}

type ImageSpec struct {
	Type       string `json:"type"`
	Repository string `json:"repository,omitempty"`
	Name       string `json:"name"`
	Tag        string `json:"tag"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SecretRef struct {
	SecretName string `json:"secretName"`
}

type CreateServiceRequest struct {
	ServiceName    string      `json:"serviceName"`
	OrganizationId string      `json:"organizationId"`
	Version        string      `json:"version"`
	ServicePort    int         `json:"servicePort"`
	Image          ImageSpec   `json:"image"`
	Resources      Resources   `json:"resources"`
	Env            []EnvVar    `json:"env,omitempty"`
	SecretRefs     []SecretRef `json:"secretRefs,omitempty"`
	Endpoint       bool        `json:"endpoint,omitempty"`
	Hostname       string      `json:"hostname,omitempty"`
	Replicas       int         `json:"replicas"`
}

type DeleteServiceRequest struct {
	ServiceName string `json:"serviceName"`
}

var (
	serviceProject         string
	serviceOrganization    string
	serviceName            string
	serviceVersion         string
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

		// Validate required flags.
		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
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

		// Ensure the user is logged in and load session cookies.
		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if serviceOrganization == "" {
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select &lt;name&gt;'", rootCmd.Use)
			}
			serviceOrganization = selectedOrg
		}

		orgId, projectId, err := internal.GetProjectId(cmd.Context(), hostname, cfgDirName, sessionFileName, serviceOrganization, serviceProject, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", serviceProject, err)
		}

		// Build env vars from repeated --env flags (NAME=VALUE).
		var env []EnvVar
		for _, e := range serviceEnvVars {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				return fmt.Errorf("invalid --env value %q; expected NAME=VALUE", e)
			}
			env = append(env, EnvVar{
				Name:  strings.TrimSpace(parts[0]),
				Value: parts[1],
			})
		}

		// Build secret references from repeated --secret flags (secret names).
		var secretRefs []SecretRef
		for _, name := range serviceSecretRefs {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				return fmt.Errorf("invalid --secret value %q; name must not be empty", name)
			}
			secretRefs = append(secretRefs, SecretRef{SecretName: trimmed})
		}

		reqBody := CreateServiceRequest{
			ServiceName:    serviceName,
			OrganizationId: orgId,
			Version:        serviceVersion,
			ServicePort:    servicePort,
			Image: ImageSpec{
				Type:       serviceImageType,
				Repository: serviceImageRepository,
				Name:       serviceImageName,
				Tag:        serviceImageTag,
			},
			Resources: Resources{
				Requests: ResourceRequirements{
					Memory: serviceReqMemory,
					CPU:    serviceReqCPU,
				},
				Limits: ResourceRequirements{
					Memory: serviceLimitMemory,
					CPU:    serviceLimitCPU,
				},
			},
			Env:        env,
			SecretRefs: secretRefs,
			Endpoint:   serviceEndpoint,
			Replicas:   serviceReplicas,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/services", orgId, projectId)

		req, err := internal.NewJSONRequestWithCookies(cmd.Context(), http.MethodPost, u.String(), bodyBytes, cookies)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		client := &http.Client{
			Timeout: defaultHTTPTimeout,
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service creation request...")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("service creation request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		serverMessage := internal.ExtractServerMessage(respBody)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if serverMessage != "" {
				return fmt.Errorf("%s", serverMessage)
			}
			return fmt.Errorf("service creation failed with status %s", resp.Status)
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

		// Validate required flags.
		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
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

		// Ensure the user is logged in and load session cookies.
		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if serviceOrganization == "" {
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select &lt;name&gt;'", rootCmd.Use)
			}
			serviceOrganization = selectedOrg
		}

		orgId, projectId, err := internal.GetProjectId(cmd.Context(), hostname, cfgDirName, sessionFileName, serviceOrganization, serviceProject, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", serviceProject, err)
		}

		// Build env vars from repeated --env flags (NAME=VALUE).
		var env []EnvVar
		for _, e := range serviceEnvVars {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				return fmt.Errorf("invalid --env value %q; expected NAME=VALUE", e)
			}
			env = append(env, EnvVar{
				Name:  strings.TrimSpace(parts[0]),
				Value: parts[1],
			})
		}

		// Build secret references from repeated --secret flags (secret names).
		var secretRefs []SecretRef
		for _, name := range serviceSecretRefs {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				return fmt.Errorf("invalid --secret value %q; name must not be empty", name)
			}
			secretRefs = append(secretRefs, SecretRef{SecretName: trimmed})
		}

		reqBody := CreateServiceRequest{
			ServiceName:    serviceName,
			OrganizationId: orgId,
			Version:        serviceVersion,
			ServicePort:    servicePort,
			Image: ImageSpec{
				Type:       serviceImageType,
				Repository: serviceImageRepository,
				Name:       serviceImageName,
				Tag:        serviceImageTag,
			},
			Resources: Resources{
				Requests: ResourceRequirements{
					Memory: serviceReqMemory,
					CPU:    serviceReqCPU,
				},
				Limits: ResourceRequirements{
					Memory: serviceLimitMemory,
					CPU:    serviceLimitCPU,
				},
			},
			Env:        env,
			SecretRefs: secretRefs,
			Endpoint:   serviceEndpoint,
			Replicas:   serviceReplicas,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/services", orgId, projectId)

		req, err := internal.NewJSONRequestWithCookies(cmd.Context(), http.MethodPut, u.String(), bodyBytes, cookies)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		client := &http.Client{
			Timeout: defaultHTTPTimeout,
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service update request...")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("service update request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		serverMessage := internal.ExtractServerMessage(respBody)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if serverMessage != "" {
				return fmt.Errorf("%s", serverMessage)
			}
			return fmt.Errorf("service update failed with status %s", resp.Status)
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

type ListServicesResponse struct {
	Services []ServiceOutput `json:"services"`
}

type ServiceOutput struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectId"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
	Updated   string `json:"updated,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
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

		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if serviceOrganization == "" {
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select &lt;name&gt;'", rootCmd.Use)
			}
			serviceOrganization = selectedOrg
		}

		orgId, projectId, err := internal.GetProjectId(cmd.Context(), hostname, cfgDirName, sessionFileName, serviceOrganization, serviceProject, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", serviceProject, err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/services", orgId, projectId)

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		client := &http.Client{
			Timeout: defaultHTTPTimeout,
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("service list request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := internal.ExtractServerMessage(respBody)
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("service listing failed with status %s", resp.Status)
		}

		var result ListServicesResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("failed to decode services response: %w", err)
		}

		headers := []string{"NAME", "REVISION", "STATUS", "ENDPOINT", "UPDATED"}
		rows := make([][]string, len(result.Services))
		for i, svc := range result.Services {
			rows[i] = []string{
				svc.Name,
				fmt.Sprintf("%d", svc.Revision),
				svc.Status,
				svc.Endpoint,
				svc.Updated,
			}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
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

		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide the service name as an argument")
		}

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if serviceOrganization == "" {
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select &lt;name&gt;'", rootCmd.Use)
			}
			serviceOrganization = selectedOrg
		}

		orgId, projectId, err := internal.GetProjectId(cmd.Context(), hostname, cfgDirName, sessionFileName, serviceOrganization, serviceProject, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", serviceProject, err)
		}

		reqBody := DeleteServiceRequest{
			ServiceName: serviceName,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/services", orgId, projectId)

		req, err := internal.NewJSONRequestWithCookies(cmd.Context(), http.MethodDelete, u.String(), bodyBytes, cookies)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		client := &http.Client{
			Timeout: defaultHTTPTimeout,
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting service deletion request...")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("service deletion request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		serverMessage := internal.ExtractServerMessage(respBody)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if serverMessage != "" {
				return fmt.Errorf("%s", serverMessage)
			}
			return fmt.Errorf("service deletion failed with status %s", resp.Status)
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

func init() {

	// Flags for "services create"
	servCCmd.Flags().StringVarP(&serviceProject, "project", "p", "", "Project name to create the service in")
	servCCmd.Flags().StringVarP(&serviceOrganization, "organization", "o", "", "Organization name that owns the project")
	servCCmd.Flags().StringVar(&serviceVersion, "version", "", "Version identifier for this service")
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
	servUCmd.Flags().StringVar(&serviceVersion, "version", "", "Version identifier for this service")
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

	// Register commands
	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(servCCmd)
	servicesCmd.AddCommand(servUCmd)
	servicesCmd.AddCommand(servListCmd)
	servicesCmd.AddCommand(servDCmd)
}
