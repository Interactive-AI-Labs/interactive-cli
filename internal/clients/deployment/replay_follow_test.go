package deployment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Only an end the next follow can resume is ErrReplayStreamEnded.
func TestFollowReplayStreamEnds(t *testing.T) {
	running := `{"run":{"status":"running"}}` + "\n"
	tests := []struct {
		name          string
		body          string
		wantResumable bool
	}{
		{"the operator's 10-minute cap", running + `{"timeout":true}` + "\n", true},
		{"a stream closed without a verdict", running, true},
		{
			"a status line past the buffer",
			`{"run":"` + strings.Repeat("x", 9<<20) + `"}` + "\n",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					fmt.Fprint(w, tt.body)
				}),
			)
			t.Cleanup(server.Close)
			client, err := NewDeploymentClient(server.URL, 5*time.Second, "test-token", "", nil)
			if err != nil {
				t.Fatalf("NewDeploymentClient() error = %v", err)
			}
			_, err = client.FollowReplay(context.Background(), "org", "proj", "agent", "run-1", nil)
			if err == nil || errors.Is(err, ErrReplayStreamEnded) != tt.wantResumable {
				t.Fatalf("err = %v, want resumable=%v", err, tt.wantResumable)
			}
		})
	}
}
