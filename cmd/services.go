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

type CreateServiceRequest struct {
	ServiceName    string    `json:"serviceName"`
	OrganizationId string    `json:"organizationId"`
	Version        string    `json:"version"`
	ServicePort    int       `json:"servicePort"`
	Image          ImageSpec `json:"image"`
	Resources      Resources `json:"resources"`
	Env            []EnvVar  `json:"env,omitempty"`
	Endpoint       bool      `json:"endpoint,omitempty"`
	Hostname       string    `json:"hostname,omitempty"`
	Replicas       int       `json:"replicas"`
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
	serviceHostname        string
	serviceReplicas        int
	serviceReqMemory       string
	serviceReqCPU          string
	serviceLimitMemory     string
	serviceLimitCPU        string

	serviceEndpoint bool
	serviceEnvVars  []string

	serviceReplicaName string
	serviceLogsFollow  bool
)

var servicesCmd = &cobra.Command{
	Use:     "services",
	Aliases: []string{"service"},
	Short:   "Manage services",
	Long:    `Manage deployment of services to InteractiveAI projects.`,
}

var servCCmd = &cobra.Command{
	Use:   "create [service_name]",
	Short: "Create a service in a project",
	Long: `Create a service in a specific project using the deployment service.

All configuration is provided via flags. The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 && serviceName == "" {
			serviceName = args[0]
		}

		// Validate required flags.
		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide --service-name")
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

		hostname := serviceHostname
		if serviceEndpoint && hostname == "" {
			hostname = fmt.Sprintf("%s.%s.dev.interactive.ai", serviceName, projectId)
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
			Env:      env,
			Endpoint: serviceEndpoint,
			Hostname: hostname,
			Replicas: serviceReplicas,
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

		if len(args) > 0 && serviceName == "" {
			serviceName = args[0]
		}

		// Validate required flags.
		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide --service-name")
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

		hostname := serviceHostname
		if serviceEndpoint && hostname == "" {
			hostname = fmt.Sprintf("%s.%s.dev.interactive.ai", serviceName, projectId)
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
			Env:      env,
			Endpoint: serviceEndpoint,
			Hostname: hostname,
			Replicas: serviceReplicas,
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
}

type ServiceReplica struct {
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	Status    string `json:"status"`
	Ready     bool   `json:"ready"`
	StartTime string `json:"startTime,omitempty"`
}

type ListServiceReplicasResponse struct {
	ServiceName string           `json:"serviceName"`
	ProjectId   string           `json:"projectId"`
	Replicas    []ServiceReplica `json:"replicas"`
}

var servListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services in a project",
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

		headers := []string{"NAME", "REVISION", "STATUS", "UPDATED"}
		rows := make([][]string, len(result.Services))
		for i, svc := range result.Services {
			rows[i] = []string{
				svc.Name,
				fmt.Sprintf("%d", svc.Revision),
				svc.Status,
				svc.Updated,
			}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var servReplicasCmd = &cobra.Command{
	Use:   "replicas [service_name]",
	Short: "List replicas for a service",
	Long: `List pods backing a service in a specific project.

The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 && serviceName == "" {
			serviceName = args[0]
		}

		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide --service-name")
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
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/services/%s/replicas", orgId, projectId, serviceName)

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
			return fmt.Errorf("replicas request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := internal.ExtractServerMessage(respBody)
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("replicas request failed with status %s", resp.Status)
		}

		var result ListServiceReplicasResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("failed to decode replicas response: %w", err)
		}

		headers := []string{"NAME", "PHASE", "STATUS", "READY", "STARTED"}
		rows := make([][]string, len(result.Replicas))
		for i, r := range result.Replicas {
			rows[i] = []string{
				r.Name,
				r.Phase,
				r.Status,
				fmt.Sprintf("%t", r.Ready),
				r.StartTime,
			}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var servLogsCmd = &cobra.Command{
	Use:   "logs [replica_name]",
	Short: "Show logs for a specific replica",
	Long: `Show logs for a specific replica (pod) in a project.

The project is selected with --project.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}

		serviceReplicaName = args[0]

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
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/services/replicas/%s/logs", orgId, projectId, serviceReplicaName)

		q := u.Query()
		if serviceLogsFollow {
			q.Set("follow", "true")
		}
		u.RawQuery = q.Encode()

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
			return fmt.Errorf("logs request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			msg := internal.ExtractServerMessage(respBody)
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("logs request failed with status %s", resp.Status)
		}

		_, err = io.Copy(out, resp.Body)
		return err
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

		if len(args) > 0 && serviceName == "" {
			serviceName = args[0]
		}

		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide --service-name")
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
	servCCmd.Flags().StringVar(&serviceProject, "project", "", "Project name to create the service in")
	servCCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")
	servCCmd.Flags().StringVar(&serviceName, "service-name", "", "Name of the service to create")
	servCCmd.Flags().StringVar(&serviceVersion, "version", "", "Version identifier for this service")
	servCCmd.Flags().IntVar(&servicePort, "service-port", 0, "Service port to expose")
	servCCmd.Flags().StringVar(&serviceImageType, "image-type", "", "Image type: internal or external")
	servCCmd.Flags().StringVar(&serviceImageRepository, "image-repository", "", "Container image repository (external images only)")
	servCCmd.Flags().StringVar(&serviceImageName, "image-name", "", "Container image name")
	servCCmd.Flags().StringVar(&serviceImageTag, "image-tag", "", "Container image tag")
	servCCmd.Flags().StringVar(&serviceHostname, "service-hostname", "", "Optional hostname for the service")
	servCCmd.Flags().IntVar(&serviceReplicas, "replicas", 1, "Number of replicas for the service")
	servCCmd.Flags().StringVar(&serviceReqMemory, "requests-memory", "512Mi", "Requested memory (e.g. 512Mi)")
	servCCmd.Flags().StringVar(&serviceReqCPU, "requests-cpu", "250m", "Requested CPU (e.g. 250m)")
	servCCmd.Flags().StringVar(&serviceLimitMemory, "limits-memory", "1Gi", "Memory limit (e.g. 1Gi)")
	servCCmd.Flags().StringVar(&serviceLimitCPU, "limits-cpu", "500m", "CPU limit (e.g. 500m)")
	servCCmd.Flags().StringArrayVar(&serviceEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	servCCmd.Flags().BoolVar(&serviceEndpoint, "endpoint", false, "Expose the service at <service-name>.<project-id>.dev.interactive.ai")

	// Flags for "services update"
	servUCmd.Flags().StringVar(&serviceProject, "project", "", "Project name to update the service in")
	servUCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")
	servUCmd.Flags().StringVar(&serviceName, "service-name", "", "Name of the service to update")
	servUCmd.Flags().StringVar(&serviceVersion, "version", "", "Version identifier for this service")
	servUCmd.Flags().IntVar(&servicePort, "service-port", 0, "Service port to expose")
	servUCmd.Flags().StringVar(&serviceImageType, "image-type", "", "Image type: internal or external")
	servUCmd.Flags().StringVar(&serviceImageRepository, "image-repository", "", "Container image repository (external images only)")
	servUCmd.Flags().StringVar(&serviceImageName, "image-name", "", "Container image name")
	servUCmd.Flags().StringVar(&serviceImageTag, "image-tag", "", "Container image tag")
	servUCmd.Flags().StringVar(&serviceHostname, "service-hostname", "", "Optional hostname for the service")
	servUCmd.Flags().IntVar(&serviceReplicas, "replicas", 1, "Number of replicas for the service")
	servUCmd.Flags().StringVar(&serviceReqMemory, "requests-memory", "512Mi", "Requested memory (e.g. 512Mi)")
	servUCmd.Flags().StringVar(&serviceReqCPU, "requests-cpu", "250m", "Requested CPU (e.g. 250m)")
	servUCmd.Flags().StringVar(&serviceLimitMemory, "limits-memory", "1Gi", "Memory limit (e.g. 1Gi)")
	servUCmd.Flags().StringVar(&serviceLimitCPU, "limits-cpu", "500m", "CPU limit (e.g. 500m)")
	servUCmd.Flags().StringArrayVar(&serviceEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	servUCmd.Flags().BoolVar(&serviceEndpoint, "endpoint", false, "Expose the service at <service-name>.<project-id>.dev.interactive.ai")

	// Flags for "services list"
	servListCmd.Flags().StringVar(&serviceProject, "project", "", "Project name to list services from")
	servListCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")

	// Flags for "services replicas"
	servReplicasCmd.Flags().StringVar(&serviceProject, "project", "", "Project name that owns the service")
	servReplicasCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")
	servReplicasCmd.Flags().StringVar(&serviceName, "service-name", "", "Name of the service to inspect")

	// Flags for "services logs"
	servLogsCmd.Flags().StringVar(&serviceProject, "project", "", "Project name that owns the service")
	servLogsCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")
	servLogsCmd.Flags().BoolVar(&serviceLogsFollow, "follow", false, "Follow log output")

	// Flags for "services delete"
	servDCmd.Flags().StringVar(&serviceProject, "project", "", "Project name to delete the service from")
	servDCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")
	servDCmd.Flags().StringVar(&serviceName, "service-name", "", "Name of the service to delete")

	// Register commands
	rootCmd.AddCommand(servicesCmd)
	rootCmd.AddCommand(servReplicasCmd)
	rootCmd.AddCommand(servLogsCmd)
	servicesCmd.AddCommand(servCCmd)
	servicesCmd.AddCommand(servUCmd)
	servicesCmd.AddCommand(servListCmd)
	servicesCmd.AddCommand(servDCmd)
}
