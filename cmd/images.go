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

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

var (
	imageBuildTag     string
	imageBuildFile    string
	imageBuildContext string
	imageName         string
	imagePushTag      string
	imageOrganization string
)

var imageCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"image"},
	Short:   "Build and manage container images",
	Long:    `Manage container images used by services.`,
}

var imageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List images for an organization",
	Long:  `List container images in the deployment registry for a specific organization.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var orgName string
		if imageOrganization != "" {
			orgName = imageOrganization
		} else {
			selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			orgName = selectedOrg
		}

		// Ensure the user is logged in and load session cookies.
		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		orgId, err := internal.GetOrgId(cmd.Context(), hostname, cfgDirName, sessionFileName, orgName, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment hostname: %w", err)
		}
		u.Path = "/images"

		q := u.Query()
		q.Set("organizationId", orgId)
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		for _, c := range cookies {
			if c != nil {
				req.AddCookie(c)
			}
		}

		client := &http.Client{
			Timeout: defaultHTTPTimeout,
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if msg := internal.ExtractServerMessage(body); msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("failed to list images: server returned %s", resp.Status)
		}

		var result struct {
			Images []struct {
				Name string   `json:"name"`
				Tags []string `json:"tags"`
			} `json:"images"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		if len(result.Images) == 0 {
			fmt.Fprintln(out, "No images found.")
			return nil
		}

		headers := []string{"NAME", "TAGS"}
		rows := make([][]string, len(result.Images))
		for i, img := range result.Images {
			rows[i] = []string{
				img.Name,
				strings.Join(img.Tags, ", "),
			}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var imageBuildCmd = &cobra.Command{
	Use:   "build [image_name]",
	Short: "Build a container image with Docker",
	Long: `Build a container image using the local Docker CLI.

This is a thin wrapper around 'docker build' that requires an explicit tag,
Dockerfile, and build context.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		in := cmd.InOrStdin()

		if len(args) > 0 && imageName == "" {
			imageName = args[0]
		}

		if imageBuildTag == "" {
			return fmt.Errorf("tag is required; please provide --tag")
		}
		if imageName == "" {
			return fmt.Errorf("image name is required; please provide --image-name")
		}
		if imageBuildFile == "" {
			return fmt.Errorf("file is required; please provide --file")
		}
		if imageBuildContext == "" {
			return fmt.Errorf("context is required; please provide --context")
		}

		// Ensure Docker CLI is available.
		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("docker CLI not found in PATH; please install Docker and ensure 'docker' is available: %w", err)
		}

		imageRef := fmt.Sprintf("%s:%s", imageName, imageBuildTag)

		args = []string{
			"build",
			"-t", imageRef,
			"-f", imageBuildFile,
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
	Use:   "push [image_name]",
	Short: "Push an image for an organization",
	Long:  `Create a Docker image tarball and push it to the deployment images endpoint for a specific organization.`,
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		in := cmd.InOrStdin()

		if len(args) > 0 && imageName == "" {
			imageName = args[0]
		}

		if imagePushTag == "" {
			return fmt.Errorf("tag is required; please provide --tag")
		}
		if imageName == "" {
			return fmt.Errorf("image name is required; please provide --image-name")
		}

		// Resolve organization name.
		var orgName string
		if imageOrganization != "" {
			orgName = imageOrganization
		} else {
			selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			orgName = selectedOrg
		}

		// Ensure the user is logged in and load session cookies.
		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		orgId, err := internal.GetOrgId(cmd.Context(), hostname, cfgDirName, sessionFileName, orgName, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		// Ensure Docker CLI is available.
		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("docker CLI not found in PATH; please install Docker and ensure 'docker' is available: %w", err)
		}

		// Local image reference that was built previously.
		imageRef := fmt.Sprintf("%s:%s", imageName, imagePushTag)

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
		u.Path = "/images"

		q := u.Query()
		q.Set("organizationId", orgId)
		q.Set("imageName", imageName)
		q.Set("tag", imagePushTag)
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, u.String(), file)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		for _, c := range cookies {
			if c != nil {
				req.AddCookie(c)
			}
		}

		req.Header.Set("Content-Type", "application/x-tar")

		client := &http.Client{
			Timeout: 5 * time.Minute,
		}

		fmt.Fprint(out, "Uploading image")
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					fmt.Fprint(out, ".")
				}
			}
		}()

		resp, err := client.Do(req)
		close(done)
		fmt.Fprintln(out)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if msg := internal.ExtractServerMessage(body); msg != "" {
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

func init() {
	imageBuildCmd.Flags().StringVarP(&imageBuildTag, "tag", "t", "", "Tag suffix to append to the fixed registry (e.g. 1.2.3)")
	imageBuildCmd.Flags().StringVar(&imageName, "image-name", "", "Image name to append to the fixed registry path")
	imageBuildCmd.Flags().StringVarP(&imageBuildFile, "file", "f", "", "Path to the Dockerfile")
	imageBuildCmd.Flags().StringVar(&imageBuildContext, "context", "", "Build context directory")

	_ = imageBuildCmd.MarkFlagRequired("tag")
	_ = imageBuildCmd.MarkFlagRequired("image-name")
	_ = imageBuildCmd.MarkFlagRequired("file")
	_ = imageBuildCmd.MarkFlagRequired("context")

	imageListCmd.Flags().StringVar(&imageOrganization, "organization", "", "Organization name to list images for")

	imagePushCmd.Flags().StringVarP(&imagePushTag, "tag", "t", "", "Tag for the image in the fixed registry (e.g. 1.2.3)")
	imagePushCmd.Flags().StringVar(&imageName, "image-name", "", "Image name to append to the fixed registry path")
	imagePushCmd.Flags().StringVar(&imageOrganization, "organization", "", "Organization name the image belongs to")
	_ = imagePushCmd.MarkFlagRequired("tag")
	_ = imagePushCmd.MarkFlagRequired("image-name")

	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageBuildCmd)
	imageCmd.AddCommand(imagePushCmd)
	imageCmd.AddCommand(imageListCmd)
}
