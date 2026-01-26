package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/spf13/cobra"
)

var (
	imageBuildTag      string
	imageBuildFile     string
	imageBuildContext  string
	imageBuildPlatform string
	imagePushTag       string
	imageOrganization  string
	imageProject       string
)

var imageCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"image"},
	Short:   "Build and manage container images",
	Long:    `Manage container images used by services.`,
}

var imageListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List images for a project",
	Long:    `List container images in the deployment registry for a specific project.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cfg := &files.StackConfig{}
		if cfgFilePath != "" {
			var err error
			cfg, err = files.LoadStackConfig(cfgFilePath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
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

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, imageOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, imageProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		images, err := deployClient.ListImages(cmd.Context(), orgId, projectId)
		if err != nil {
			return err
		}

		if len(images) == 0 {
			fmt.Fprintln(out, "No images found.")
			return nil
		}

		headers := []string{"NAME", "TAGS"}
		rows := make([][]string, len(images))
		for i, img := range images {
			rows[i] = []string{
				img.Name,
				strings.Join(img.Tags, ", "),
			}
		}

		if err := output.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var imageBuildCmd = &cobra.Command{
	Use:     "build [image_name]",
	Aliases: []string{"b"},
	Short:   "Build a container image with Docker",
	Long: `Build a container image using the local Docker CLI.

This is a thin wrapper around 'docker build' that requires an explicit tag,
Dockerfile, and build context.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		in := cmd.InOrStdin()

		imageName := strings.TrimSpace(args[0])

		if imageBuildTag == "" {
			return fmt.Errorf("tag is required; please provide --tag")
		}
		if imageBuildFile == "" {
			return fmt.Errorf("file is required; please provide --file")
		}
		if imageBuildContext == "" {
			return fmt.Errorf("context is required; please provide --context")
		}

		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("docker CLI not found in PATH; please install Docker and ensure 'docker' is available: %w", err)
		}

		imageRef := fmt.Sprintf("%s:%s", imageName, imageBuildTag)

		args = []string{
			"build",
			"-t", imageRef,
			"-f", imageBuildFile,
			"--platform", imageBuildPlatform,
			imageBuildContext,
		}

		cmdExec := exec.CommandContext(cmd.Context(), "docker", args...)
		cmdExec.Stdout = out
		cmdExec.Stderr = out
		cmdExec.Stdin = in

		if err := cmdExec.Run(); err != nil {
			return fmt.Errorf("docker build failed: %w", err)
		}

		return nil
	},
}

var imagePushCmd = &cobra.Command{
	Use:     "push [image_name]",
	Aliases: []string{"p"},
	Short:   "Push an image for a project",
	Long:    `Create a Docker image tarball and push it to the deployment images endpoint for a specific project.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		in := cmd.InOrStdin()

		imageName := strings.TrimSpace(args[0])

		if imagePushTag == "" {
			return fmt.Errorf("tag is required; please provide --tag")
		}

		cfg := &files.StackConfig{}
		if cfgFilePath != "" {
			var err error
			cfg, err = files.LoadStackConfig(cfgFilePath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, imageOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, imageProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("docker CLI not found in PATH; please install Docker and ensure 'docker' is available: %w", err)
		}

		imageRef := fmt.Sprintf("%s:%s", imageName, imagePushTag)

		if err := validateImageArchitecture(imageRef); err != nil {
			return err
		}

		tmpFile, err := os.CreateTemp("", "image-*.tar")
		if err != nil {
			return fmt.Errorf("failed to create temporary file for image tarball: %w", err)
		}
		defer func() {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFile.Name())
		}()

		dockerArgs := []string{
			"save",
			"-o", tmpFile.Name(),
			imageRef,
		}

		cmdExec := exec.CommandContext(cmd.Context(), "docker", dockerArgs...)
		cmdExec.Stdout = out
		cmdExec.Stderr = out
		cmdExec.Stdin = in

		if err := cmdExec.Run(); err != nil {
			return fmt.Errorf("docker save failed: %w", err)
		}

		file, err := os.Open(tmpFile.Name())
		if err != nil {
			return fmt.Errorf("failed to open image tarball: %w", err)
		}
		defer file.Close()

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment hostname: %w", err)
		}
		u.Path = fmt.Sprintf("/v1/organizations/%s/projects/%s/images", orgId, projectId)

		q := u.Query()
		q.Set("imageName", imageName)
		q.Set("tag", imagePushTag)
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, u.String(), file)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		if err := clients.ApplyAuth(req, apiKey, cookies); err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-tar")

		httpClient := &http.Client{
			Timeout: 5 * time.Minute,
		}

		fmt.Fprintln(out)
		fmt.Fprint(out, "Uploading image")
		done := output.PrintLoadingDots(out)

		resp, err := httpClient.Do(req)
		close(done)
		fmt.Fprintln(out)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if msg := clients.ExtractServerMessage(body); msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("failed to push image: server returned %s", resp.Status)
		}

		var result struct {
			ImageRef string `json:"ImageRef"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to decode message: %s", err.Error())
		}

		fmt.Fprintf(out, "Pushed image %s\n", result.ImageRef)

		return nil
	},
}

func validateImageArchitecture(imageRef string) error {
	inspectArgs := []string{"inspect", "--format", "{{.Architecture}}", imageRef}
	cmdExec := exec.Command("docker", inspectArgs...)

	output, err := cmdExec.Output()
	if err != nil {
		return fmt.Errorf("failed to inspect image architecture: %w", err)
	}

	arch := strings.TrimSpace(string(output))
	if arch != "amd64" && arch != "x86_64" {
		return fmt.Errorf("unsupported architecture %q detected in image; only amd64 images are supported on this platform", arch)
	}

	return nil
}

func init() {
	imageBuildCmd.Flags().StringVarP(&imageBuildTag, "tag", "t", "", "Tag suffix to append to the fixed registry (e.g. 1.2.3)")
	imageBuildCmd.Flags().StringVarP(&imageBuildFile, "file", "f", "Dockerfile", "Path to the Dockerfile (default: ./Dockerfile)")
	imageBuildCmd.Flags().StringVarP(&imageBuildContext, "context", "c", ".", "Build context directory (default: current directory)")
	imageBuildCmd.Flags().StringVar(&imageBuildPlatform, "platform", "linux/amd64", "Target platform for the build (currently only linux/amd64 is supported)")

	_ = imageBuildCmd.MarkFlagRequired("tag")

	imageListCmd.Flags().StringVarP(&imageOrganization, "organization", "o", "", "Organization name that owns the project")
	imageListCmd.Flags().StringVarP(&imageProject, "project", "p", "", "Project name to list images for")

	imagePushCmd.Flags().StringVarP(&imagePushTag, "tag", "t", "", "Tag for the image in the fixed registry (e.g. 1.2.3)")
	imagePushCmd.Flags().StringVarP(&imageOrganization, "organization", "o", "", "Organization name that owns the project")
	imagePushCmd.Flags().StringVarP(&imageProject, "project", "p", "", "Project name the image belongs to")
	_ = imagePushCmd.MarkFlagRequired("tag")

	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageBuildCmd)
	imageCmd.AddCommand(imagePushCmd)
	imageCmd.AddCommand(imageListCmd)
}
