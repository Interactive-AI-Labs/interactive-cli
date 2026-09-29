package deployment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

type JobCron struct {
	Schedule string `json:"schedule"           yaml:"schedule"`
	Timezone string `json:"timezone,omitempty" yaml:"timezone,omitempty"`
}

type JobRetention struct {
	TTL            *int32 `json:"ttl,omitempty"            yaml:"ttl,omitempty"`
	SuccessfulRuns *int32 `json:"successfulRuns,omitempty" yaml:"successfulRuns,omitempty"`
	FailedRuns     *int32 `json:"failedRuns,omitempty"     yaml:"failedRuns,omitempty"`
}

type CreateJobBody struct {
	Type       string        `json:"type"`
	Image      *ImageSpec    `json:"image,omitempty"`
	Command    []string      `json:"command,omitempty"`
	Args       []string      `json:"args,omitempty"`
	Script     string        `json:"script,omitempty"`
	Pyproject  string        `json:"pyproject,omitempty"`
	Resources  Resources     `json:"resources"`
	Env        []EnvVar      `json:"env,omitempty"`
	SecretRefs []SecretRef   `json:"secretRefs,omitempty"`
	StackId    string        `json:"stackId,omitempty"`
	Cron       *JobCron      `json:"cron,omitempty"`
	Timeout    *int64        `json:"timeout,omitempty"`
	Retries    *int32        `json:"retries,omitempty"`
	Retention  *JobRetention `json:"retention,omitempty"`
}

type JobOutput struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectId"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
	Updated   string `json:"updated,omitempty"`
}

type DescribeJobResponse struct {
	JobOutput
	CreateJobBody
	Message string `json:"message"`
}

type JobRun struct {
	RunId    string `json:"runId"`
	JobName  string `json:"jobName"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Created  string `json:"created"`
	Started  string `json:"started,omitempty"`
	Finished string `json:"finished,omitempty"`
}

func jobsPath(orgId, projectId, jobName string) string {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/jobs",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
	)
	if jobName != "" {
		path += "/" + url.PathEscape(jobName)
	}
	return path
}

func jobRunPath(orgId, projectId, runId string) string {
	return fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/job-runs/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(runId),
	)
}

func (c *DeploymentClient) CreateJob(
	ctx context.Context,
	orgId, projectId, jobName string,
	body CreateJobBody,
) (string, error) {
	data, err := c.sendJSONRequest(ctx, http.MethodPost, jobsPath(orgId, projectId, jobName), body)
	if err != nil {
		return "", err
	}
	return clients.ExtractServerMessage(data), nil
}

// ListJobs lists every job in the project when stackId is empty.
func (c *DeploymentClient) ListJobs(
	ctx context.Context,
	orgId, projectId, stackId string,
) ([]JobOutput, error) {
	path := jobsPath(orgId, projectId, "")
	if stackId != "" {
		path += "?" + url.Values{"stackId": {stackId}}.Encode()
	}
	var response struct {
		Jobs []JobOutput `json:"jobs"`
	}
	err := c.sendJSONInto(
		ctx,
		http.MethodGet,
		path,
		nil,
		"list jobs",
		&response,
	)
	return response.Jobs, err
}

func (c *DeploymentClient) DescribeJob(
	ctx context.Context,
	orgId, projectId, jobName string,
) (*DescribeJobResponse, error) {
	var response DescribeJobResponse
	if err := c.sendJSONInto(
		ctx,
		http.MethodGet,
		jobsPath(orgId, projectId, jobName),
		nil,
		"describe job",
		&response,
	); err != nil {
		return nil, err
	}
	return &response, nil
}

// PutJob replaces the whole definition; omitted optional fields return to their defaults.
func (c *DeploymentClient) PutJob(
	ctx context.Context,
	orgId, projectId, jobName string,
	body CreateJobBody,
) (string, error) {
	data, err := c.sendJSONRequest(ctx, http.MethodPut, jobsPath(orgId, projectId, jobName), body)
	if err != nil {
		return "", err
	}
	return clients.ExtractServerMessage(data), nil
}

// PatchJob changes only the supplied definition fields.
func (c *DeploymentClient) PatchJob(
	ctx context.Context,
	orgId, projectId, jobName string,
	patch UpdatePatch,
) (string, error) {
	data, err := c.sendJSONRequest(
		ctx,
		http.MethodPatch,
		jobsPath(orgId, projectId, jobName),
		patch,
	)
	if err != nil {
		return "", err
	}
	return clients.ExtractServerMessage(data), nil
}

func (c *DeploymentClient) DeleteJob(
	ctx context.Context,
	orgId, projectId, jobName string,
) (string, error) {
	data, err := c.sendJSONRequest(ctx, http.MethodDelete, jobsPath(orgId, projectId, jobName), nil)
	if err != nil {
		return "", err
	}
	return clients.ExtractServerMessage(data), nil
}

// JobAction activates or deactivates a job.
func (c *DeploymentClient) JobAction(
	ctx context.Context,
	orgId, projectId, jobName, action string,
) (string, error) {
	data, err := c.sendJSONRequest(
		ctx,
		http.MethodPost,
		jobsPath(orgId, projectId, jobName)+"/"+url.PathEscape(action),
		nil,
	)
	if err != nil {
		return "", err
	}
	return clients.ExtractServerMessage(data), nil
}

func (c *DeploymentClient) StartJobRun(
	ctx context.Context,
	orgId, projectId, jobName string,
) (*JobRun, error) {
	var response JobRun
	if err := c.sendJSONInto(
		ctx,
		http.MethodPost,
		jobsPath(orgId, projectId, jobName)+"/run",
		nil,
		"run job",
		&response,
	); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *DeploymentClient) ListJobRuns(
	ctx context.Context,
	orgId, projectId, jobName string,
) ([]JobRun, error) {
	var response struct {
		Runs []JobRun `json:"runs"`
	}
	err := c.sendJSONInto(
		ctx,
		http.MethodGet,
		jobsPath(orgId, projectId, jobName)+"/runs",
		nil,
		"list job runs",
		&response,
	)
	return response.Runs, err
}

func (c *DeploymentClient) DescribeJobRun(
	ctx context.Context,
	orgId, projectId, runId string,
) (*JobRun, error) {
	var response JobRun
	if err := c.sendJSONInto(
		ctx,
		http.MethodGet,
		jobRunPath(orgId, projectId, runId),
		nil,
		"describe job run",
		&response,
	); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *DeploymentClient) DeleteJobRun(
	ctx context.Context,
	orgId, projectId, runId string,
) (string, error) {
	data, err := c.sendJSONRequest(
		ctx,
		http.MethodDelete,
		jobRunPath(orgId, projectId, runId),
		nil,
	)
	if err != nil {
		return "", err
	}
	return clients.ExtractServerMessage(data), nil
}

func (c *DeploymentClient) GetJobLogs(
	ctx context.Context,
	orgId, projectId, jobName string,
	opts LogsOptions,
) (*LogsResponse, error) {
	return c.fetchLogs(ctx, jobsPath(orgId, projectId, jobName)+"/logs", opts)
}

func (c *DeploymentClient) GetJobRunLogs(
	ctx context.Context,
	orgId, projectId, runId string,
	opts LogsOptions,
) (*LogsResponse, error) {
	return c.fetchLogs(ctx, jobRunPath(orgId, projectId, runId)+"/logs", opts)
}

func (c *DeploymentClient) StopJobRun(
	ctx context.Context,
	orgId, projectId, runId string,
) (*JobRun, error) {
	var response JobRun
	if err := c.sendJSONInto(
		ctx,
		http.MethodPost,
		jobRunPath(orgId, projectId, runId)+"/stop",
		nil,
		"stop job run",
		&response,
	); err != nil {
		return nil, err
	}
	return &response, nil
}
