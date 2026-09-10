package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return newTestClientMode(t, false, handler)
}

func newTestClientMode(t *testing.T, follow bool, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient(server.URL, func(req *http.Request) error {
		req.Header.Set("Authorization", "Bearer test-bearer")
		return nil
	}, follow, 5*time.Second)
}

func waitOpts(onProgress func(*Run)) WaitOptions {
	return WaitOptions{
		Interval:    time.Millisecond,
		Timeout:     time.Second,
		NotFoundCap: 30 * time.Millisecond,
		OnProgress:  onProgress,
	}
}

func TestStartReplay(t *testing.T) {
	var gotBody, gotAuth string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/replays" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(
			w,
			`{"run_id":"abc","skipped":[{"id":"old","reason":"not active"}],"status":"accepted"}`,
		)
	})

	resp, err := client.StartReplay(context.Background(), StartRequest{
		Dataset: "replay-chat", Scenarios: []string{"a", "b"}, Repeat: 2, Concurrency: 8,
	})
	if err != nil {
		t.Fatalf("StartReplay() error = %v", err)
	}
	if resp.RunID != "abc" || len(resp.Skipped) != 1 || resp.Skipped[0].ID != "old" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if gotAuth != "Bearer test-bearer" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	want := `{"dataset":"replay-chat","scenarios":["a","b"],"repeat":2,"concurrency":8}`
	if gotBody != want {
		t.Errorf("body = %s, want %s", gotBody, want)
	}
}

func TestStartReplayErrors(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantDetail  string
		wantSkipped int
	}{
		{
			name:        "flat detail",
			status:      400,
			body:        `{"detail":"dataset 'x' holds no replayable scenario","skipped":[{"id":"p","reason":"invalid"}]}`,
			wantDetail:  "dataset 'x' holds no replayable scenario",
			wantSkipped: 1,
		},
		{
			name:       "422 list",
			status:     422,
			body:       `{"detail":[{"loc":["body","repeat"],"msg":"Input should be less than or equal to 20","type":"x"}]}`,
			wantDetail: "repeat: Input should be less than or equal to 20",
		},
		{
			name:       "plain text",
			status:     401,
			body:       "Unauthorized",
			wantDetail: "Unauthorized",
		},
		{
			name:       "non-JSON html",
			status:     403,
			body:       "<html>blocked</html>",
			wantDetail: "<html>blocked</html>",
		},
		{
			name:       "empty body",
			status:     502,
			body:       "",
			wantDetail: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			})
			_, err := client.StartReplay(context.Background(), StartRequest{Dataset: "x"})
			var ae *Error
			if !errors.As(err, &ae) {
				t.Fatalf("error = %v, want *Error", err)
			}
			if ae.Status != tt.status || ae.Detail != tt.wantDetail ||
				len(ae.Skipped) != tt.wantSkipped {
				t.Errorf("got %+v", ae)
			}
		})
	}
}

func TestGetReplayNormalisesBatch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(
			w,
			`{"run_id":"b1","scenario":"account-lock","status":"passed","repeat":2,"passed":2,`+
				`"iterations":[{"run_id":"b1-1","status":"passed"},{"run_id":"b1-2","status":"passed"}]}`,
		)
	})
	run, err := client.GetReplay(context.Background(), "b1")
	if err != nil {
		t.Fatalf("GetReplay() error = %v", err)
	}
	if len(run.Batches) != 1 || run.Batches[0].Scenario != "account-lock" ||
		len(run.Batches[0].Iterations) != 2 || run.Iterations != nil {
		t.Errorf("batch not normalised: %+v", run)
	}
	if !strings.Contains(string(run.Raw), `"run_id":"b1"`) {
		t.Errorf("Raw not preserved: %s", run.Raw)
	}
}

func TestWait(t *testing.T) {
	t.Run("running then passed", func(t *testing.T) {
		var calls atomic.Int32
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/replays/r1" {
				t.Errorf("path = %s", r.URL.Path)
			}
			if calls.Add(1) < 3 {
				fmt.Fprint(
					w,
					`{"run_id":"r1","status":"running","batches":[{"status":"passed"},{"status":"running"}]}`,
				)
				return
			}
			fmt.Fprint(
				w,
				`{"run_id":"r1","status":"passed","batches":[{"status":"passed"},{"status":"passed"}]}`,
			)
		})
		var progress int
		run, err := client.Wait(context.Background(), "r1", waitOpts(func(*Run) { progress++ }))
		if err != nil {
			t.Fatalf("Wait() error = %v", err)
		}
		if run.Status != StatusPassed || progress != 3 {
			t.Errorf("status = %s, progress calls = %d", run.Status, progress)
		}
	})

	t.Run("404 then 200", func(t *testing.T) {
		var calls atomic.Int32
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) < 3 {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"detail":"No such replay run: r1"}`)
				return
			}
			fmt.Fprint(w, `{"run_id":"r1","status":"failed"}`)
		})
		opts := waitOpts(nil)
		var retries []error
		opts.OnRetry = func(err error) { retries = append(retries, err) }
		run, err := client.Wait(context.Background(), "r1", opts)
		if err != nil {
			t.Fatalf("Wait() error = %v", err)
		}
		if run.Status != StatusFailed {
			t.Errorf("status = %s", run.Status)
		}
		if len(retries) != 1 || !strings.Contains(retries[0].Error(), "No such replay run") {
			t.Errorf("OnRetry calls = %v, want exactly one for the streak", retries)
		}
	})

	t.Run("404 past the cap is a lost run", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		_, err := client.Wait(context.Background(), "r1", waitOpts(nil))
		if err == nil || !strings.Contains(err.Error(), "not found for 30ms") {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("5xx then 200", func(t *testing.T) {
		var calls atomic.Int32
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) == 1 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			fmt.Fprint(w, `{"run_id":"r1","status":"passed"}`)
		})
		if _, err := client.Wait(context.Background(), "r1", waitOpts(nil)); err != nil {
			t.Fatalf("Wait() error = %v", err)
		}
	})

	t.Run("401 aborts at once", func(t *testing.T) {
		var calls atomic.Int32
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized")
		})
		_, err := client.Wait(context.Background(), "r1", waitOpts(nil))
		var ae *Error
		if !errors.As(err, &ae) || ae.Status != 401 || calls.Load() != 1 {
			t.Errorf("error = %v, calls = %d", err, calls.Load())
		}
	})

	t.Run("timeout", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"run_id":"r1","status":"running"}`)
		})
		opts := waitOpts(nil)
		opts.Timeout = 20 * time.Millisecond
		_, err := client.Wait(context.Background(), "r1", opts)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("error = %v, want DeadlineExceeded", err)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"run_id":"r1","status":"running"}`)
		})
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()
		_, err := client.Wait(ctx, "r1", waitOpts(nil))
		if !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want Canceled", err)
		}
	})
}

// ndjson writes each line as one stream event and flushes, as the platform does.
func ndjson(w http.ResponseWriter, lines ...string) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	f, _ := w.(http.Flusher)
	for _, l := range lines {
		fmt.Fprintln(w, l)
		if f != nil {
			f.Flush()
		}
	}
}

func TestWaitFollow(t *testing.T) {
	tests := []struct {
		name         string
		handler      func(calls *atomic.Int32, w http.ResponseWriter, r *http.Request)
		wantStatus   string
		wantErr      string
		wantProgress int
		wantRequests int32
	}{
		{
			name: "one stream carries the run to its verdict",
			handler: func(_ *atomic.Int32, w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("follow") != "true" {
					t.Errorf("follow query missing: %s", r.URL.RawQuery)
				}
				ndjson(
					w,
					`{"run":{"run_id":"r1","status":"running","batches":[{"status":"running"}]}}`,
					`{"run":{"run_id":"r1","status":"running","batches":[{"status":"passed"}]}}`,
					`{"run":{"run_id":"r1","status":"passed","batches":[{"status":"passed"}]}}`,
				)
			},
			wantStatus: StatusPassed, wantProgress: 3, wantRequests: 1,
		},
		{
			name: "the platform's timeout line ends the follow, no reconnection",
			handler: func(_ *atomic.Int32, w http.ResponseWriter, r *http.Request) {
				ndjson(w, `{"run":{"run_id":"r1","status":"running"}}`, `{"timeout":true}`)
			},
			wantErr: ErrStreamEnded.Error(), wantProgress: 1, wantRequests: 1,
		},
		{
			name: "the platform's error line ends the follow with its own wording",
			handler: func(_ *atomic.Int32, w http.ResponseWriter, r *http.Request) {
				ndjson(w, `{"error":"Agent could not be reached for 5m0s"}`)
			},
			wantErr: "Agent could not be reached for 5m0s", wantRequests: 1,
		},
		{
			name: "a stream that drops before the verdict is not retried",
			handler: func(_ *atomic.Int32, w http.ResponseWriter, r *http.Request) {
				ndjson(w, `{"run":{"run_id":"r1","status":"running"}}`)
			},
			wantErr: ErrStreamEnded.Error(), wantProgress: 1, wantRequests: 1,
		},
		{
			name: "a refusal before the stream surfaces the platform's message",
			handler: func(_ *atomic.Int32, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"code":403,"message":"Not allowed to replay this agent"}`)
			},
			wantErr: "Not allowed to replay this agent", wantRequests: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			var requests atomic.Int32
			client := newTestClientMode(t, true, func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				tt.handler(&calls, w, r)
			})
			var progress int
			opts := waitOpts(func(*Run) { progress++ })
			opts.OnRetry = func(error) { t.Error("a followed run must not report retries") }

			run, err := client.Wait(context.Background(), "r1", opts)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("Wait() error = %v", err)
			} else if run.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", run.Status, tt.wantStatus)
			}
			if progress != tt.wantProgress {
				t.Errorf("progress calls = %d, want %d", progress, tt.wantProgress)
			}
			if tt.wantRequests != 0 && requests.Load() != tt.wantRequests {
				t.Errorf("requests = %d, want %d", requests.Load(), tt.wantRequests)
			}
		})
	}
}
