package platform

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestSessionWrites(t *testing.T) {
	const sent = "sent"
	refused := func(action string) string {
		return "this command requires API key authentication; a session login or token is not supported for " +
			action + ". Create a project API key in the web app and pass it with --api-key or INTERACTIVE_API_KEY, " +
			"or do this in the web app"
	}
	tests := []struct {
		name string
		call func(context.Context, *APIClient) error
		want string
	}{
		{"create comment", func(ctx context.Context, c *APIClient) error {
			_, _, err := c.CreateComment(ctx, "o", "p", CommentCreateBody{})
			return err
		}, sent},
		{"create queue", func(ctx context.Context, c *APIClient) error {
			_, _, err := c.CreateAnnotationQueue(ctx, "o", "p", AnnotationQueueCreateBody{})
			return err
		}, sent},
		{"assign queue", func(ctx context.Context, c *APIClient) error {
			_, err := c.AssignQueue(ctx, "o", "p", "q", "u")
			return err
		}, refused("assigning queues")},
		{"unassign queue", func(ctx context.Context, c *APIClient) error {
			_, err := c.UnassignQueue(ctx, "o", "p", "q", "u")
			return err
		}, refused("unassigning queues")},
		{"create queue item", func(ctx context.Context, c *APIClient) error {
			_, _, err := c.CreateQueueItem(ctx, "o", "p", "q", QueueItemCreateBody{})
			return err
		}, sent},
		{"delete trace", func(ctx context.Context, c *APIClient) error {
			_, err := c.DeleteTrace(ctx, "o", "p", "t")
			return err
		}, sent},
		{"delete traces", func(ctx context.Context, c *APIClient) error {
			_, err := c.DeleteTraces(ctx, "o", "p", BulkTraceDeleteBody{})
			return err
		}, sent},
		{"create score config", func(ctx context.Context, c *APIClient) error {
			_, _, err := c.CreateScoreConfig(ctx, "o", "p", ScoreConfigCreateBody{})
			return err
		}, sent},
		{"update score config", func(ctx context.Context, c *APIClient) error {
			_, _, err := c.UpdateScoreConfig(ctx, "o", "p", "sc", ScoreConfigUpdateBody{})
			return err
		}, sent},
		{"create score", func(ctx context.Context, c *APIClient) error {
			_, _, err := c.CreateScore(ctx, "o", "p", ScoreCreateBody{})
			return err
		}, refused("creating scores")},
		{"delete score", func(ctx context.Context, c *APIClient) error {
			_, err := c.DeleteScore(ctx, "o", "p", "s")
			return err
		}, refused("deleting scores")},
		{"create run item", func(ctx context.Context, c *APIClient) error {
			_, _, err := c.CreateDatasetRunItem(ctx, "o", "p", DatasetRunItemCreateBody{})
			return err
		}, refused("creating dataset run items")},
	}

	// A cancelled request proves the write got past the client without reaching a server.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := newEvalTestClient(t, "http://platform.invalid")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(ctx, client)
			got := fmt.Sprint(err)
			if errors.Is(err, context.Canceled) {
				got = sent
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
