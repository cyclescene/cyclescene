package admin

import (
	"context"
	"fmt"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
)

// RunClient provides methods to interact with Cloud Run Jobs
type RunClient struct {
	jobsClient       *run.JobsClient
	executionsClient *run.ExecutionsClient
	projectID        string
	region           string
}

// NewRunClient creates a new Cloud Run Jobs client
func NewRunClient(ctx context.Context, projectID, region string) (*RunClient, error) {
	jobsClient, err := run.NewJobsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run Jobs client: %w", err)
	}

	executionsClient, err := run.NewExecutionsClient(ctx)
	if err != nil {
		jobsClient.Close()
		return nil, fmt.Errorf("failed to create Cloud Run Executions client: %w", err)
	}

	return &RunClient{
		jobsClient:       jobsClient,
		executionsClient: executionsClient,
		projectID:        projectID,
		region:           region,
	}, nil
}

// TriggerStravaSync triggers the strava-sync Cloud Run Job
// Returns the execution name if successful
func (c *RunClient) TriggerStravaSync(ctx context.Context) (string, error) {
	jobName := fmt.Sprintf("projects/%s/locations/%s/jobs/strava-sync", c.projectID, c.region)

	req := &runpb.RunJobRequest{
		Name: jobName,
	}

	op, err := c.jobsClient.RunJob(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to trigger job: %w", err)
	}

	// Wait for the operation to complete (job started, not finished)
	execution, err := op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for job execution: %w", err)
	}

	return execution.Name, nil
}

// GetJobExecution gets the status of a specific job execution
func (c *RunClient) GetJobExecution(ctx context.Context, executionName string) (*runpb.Execution, error) {
	req := &runpb.GetExecutionRequest{
		Name: executionName,
	}

	execution, err := c.executionsClient.GetExecution(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return execution, nil
}

// ListRecentExecutions lists recent executions of the strava-sync job
func (c *RunClient) ListRecentExecutions(ctx context.Context, limit int) ([]*runpb.Execution, error) {
	jobName := fmt.Sprintf("projects/%s/locations/%s/jobs/strava-sync", c.projectID, c.region)

	req := &runpb.ListExecutionsRequest{
		Parent:   jobName,
		PageSize: int32(limit),
	}

	var executions []*runpb.Execution
	it := c.executionsClient.ListExecutions(ctx, req)
	for {
		execution, err := it.Next()
		if err != nil {
			break // End of iteration
		}
		executions = append(executions, execution)
		if len(executions) >= limit {
			break
		}
	}

	return executions, nil
}

// Close closes the underlying client connections
func (c *RunClient) Close() error {
	if err := c.executionsClient.Close(); err != nil {
		c.jobsClient.Close()
		return err
	}
	return c.jobsClient.Close()
}
