package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

type CreateServiceRequest struct {
	ServiceName    string    `json:"serviceName"`
	ProjectId      string    `json:"projectId"`
	OrganizationId string    `json:"organizationId"`
	Version        string    `json:"version"`
	ServicePort    int       `json:"servicePort"`
	Image          ImageSpec `json:"image"`
	Resources      Resources `json:"resources"`
	Hostname       string    `json:"hostname,omitempty"`
	Replicas       int       `json:"replicas"`
}

type DeleteServiceRequest struct {
	ServiceName string `json:"serviceName"`
	ProjectId   string `json:"projectId"`
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
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage services",
	Long:  `Manage deployment of services to InteractiveAI projects.`,
}

var servCCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a service in a project",
	Long: `Create a service in a specific project using the deployment service.

All configuration is provided via flags. The project is selected with --project.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		// Validate required flags.
		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide --service-name")
		}

		if serviceVersion == "" {
			return fmt.Errorf("version is required; please provide --version")
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

		reqBody := CreateServiceRequest{
			ServiceName:    serviceName,
			ProjectId:      projectId,
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
			Hostname: serviceHostname,
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
		u.Path = "/services"

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
	Use:   "update",
	Short: "Update a service in a project",
	Long: `Update a service in a specific project using the deployment service.

All configuration is provided via flags. The project is selected with --project.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		// Validate required flags.
		if serviceProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required; please provide --service-name")
		}

		if serviceVersion == "" {
			return fmt.Errorf("version is required; please provide --version")
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

		reqBody := CreateServiceRequest{
			ServiceName:    serviceName,
			ProjectId:      projectId,
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
			Hostname: serviceHostname,
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
		u.Path = "/services"

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

		_, projectId, err := internal.GetProjectId(cmd.Context(), hostname, cfgDirName, sessionFileName, serviceOrganization, serviceProject, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", serviceProject, err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = "/services"

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, u.String()+"?project-id="+projectId, nil)
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

var servDCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a service from a project",
	Long: `Delete a service from a specific project using the deployment service.

The project is selected with --project.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

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

		_, projectId, err := internal.GetProjectId(cmd.Context(), hostname, cfgDirName, sessionFileName, serviceOrganization, serviceProject, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", serviceProject, err)
		}

		reqBody := DeleteServiceRequest{
			ServiceName: serviceName,
			ProjectId:   projectId,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = "/services"

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

	// Flags for "services list"
	servListCmd.Flags().StringVar(&serviceProject, "project", "", "Project name to list services from")
	servListCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")

	// Flags for "services delete"
	servDCmd.Flags().StringVar(&serviceProject, "project", "", "Project name to delete the service from")
	servDCmd.Flags().StringVar(&serviceOrganization, "organization", "", "Organization name that owns the project")
	servDCmd.Flags().StringVar(&serviceName, "service-name", "", "Name of the service to delete")

	// Register commands
	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(servCCmd)
	servicesCmd.AddCommand(servUCmd)
	servicesCmd.AddCommand(servListCmd)
	servicesCmd.AddCommand(servDCmd)
}
