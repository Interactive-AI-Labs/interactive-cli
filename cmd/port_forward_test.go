package cmd

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// buildPortForwardURL
// ---------------------------------------------------------------------------

func TestBuildPortForwardURL(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		orgId        string
		projectId    string
		resourceType string
		resourceName string
		port         int
		want         string
		wantErr      bool
	}{
		{
			name:         "https becomes wss",
			host:         "https://deployment.example.com",
			orgId:        "org-1",
			projectId:    "proj-2",
			resourceType: "services",
			resourceName: "my-svc",
			port:         0,
			want:         "wss://deployment.example.com/v1/organizations/org-1/projects/proj-2/services/my-svc/port-forward",
		},
		{
			name:         "http becomes ws",
			host:         "http://localhost:8080",
			orgId:        "org-1",
			projectId:    "proj-2",
			resourceType: "agents",
			resourceName: "my-agent",
			port:         0,
			want:         "ws://localhost:8080/v1/organizations/org-1/projects/proj-2/agents/my-agent/port-forward",
		},
		{
			name:         "port query param added when > 0",
			host:         "https://deployment.example.com",
			orgId:        "org-1",
			projectId:    "proj-2",
			resourceType: "databases",
			resourceName: "my-db",
			port:         5432,
			want:         "wss://deployment.example.com/v1/organizations/org-1/projects/proj-2/databases/my-db/port-forward?port=5432",
		},
		{
			name:         "special characters are escaped",
			host:         "https://deployment.example.com",
			orgId:        "org/1",
			projectId:    "proj 2",
			resourceType: "services",
			resourceName: "svc/name",
			port:         0,
			want:         "wss://deployment.example.com/v1/organizations/org%2F1/projects/proj%202/services/svc%2Fname/port-forward",
		},
		{
			name:         "unknown scheme defaults to wss",
			host:         "ftp://deployment.example.com",
			orgId:        "org-1",
			projectId:    "proj-2",
			resourceType: "services",
			resourceName: "my-svc",
			port:         0,
			want:         "wss://deployment.example.com/v1/organizations/org-1/projects/proj-2/services/my-svc/port-forward",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPortForwardURL(
				tt.host, tt.orgId, tt.projectId,
				tt.resourceType, tt.resourceName, tt.port,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("buildPortForwardURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("buildPortForwardURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// authHeaders
// ---------------------------------------------------------------------------

func TestAuthHeaders(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		apiKey     string
		fakeHome   bool
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "bearer token",
			token:      "my-token",
			apiKey:     "",
			wantPrefix: "Bearer my-token",
		},
		{
			name:       "token takes precedence over api key",
			token:      "my-token",
			apiKey:     "my-key",
			wantPrefix: "Bearer ",
		},
		{
			name:       "api key uses basic auth",
			token:      "",
			apiKey:     "my-key",
			wantPrefix: "Basic ",
		},
		{
			name:     "no auth returns error",
			token:    "",
			apiKey:   "",
			fakeHome: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origToken := token
			origApiKey := apiKey
			defer func() { token = origToken; apiKey = origApiKey }()

			token = tt.token
			apiKey = tt.apiKey
			if tt.fakeHome {
				t.Setenv("HOME", t.TempDir())
			}

			h, err := authHeaders()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := h.Get("Authorization")
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("Authorization = %q, want prefix %q", got, tt.wantPrefix)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// handlePortForwardConn — full tunnel test
// ---------------------------------------------------------------------------

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func echoWSHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	for {
		mt, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}
		if err := ws.WriteMessage(mt, msg); err != nil {
			return
		}
	}
}

func TestHandlePortForwardConn_Echo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(echoWSHandler))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx := context.Background()

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handlePortForwardConn(ctx, serverConn, wsURL, http.Header{})
	}()

	payload := "hello tunnel"
	if _, err := clientConn.Write([]byte(payload)); err != nil {
		t.Fatalf("write: %v", err)
	}

	buf := make([]byte, 256)
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := string(buf[:n]); got != payload {
		t.Errorf("echo = %q, want %q", got, payload)
	}

	clientConn.Close()
	wg.Wait()
}

func TestHandlePortForwardConn_MultipleMessages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(echoWSHandler))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx := context.Background()

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handlePortForwardConn(ctx, serverConn, wsURL, http.Header{})
	}()

	for i, msg := range []string{"first", "second", "third"} {
		if _, err := clientConn.Write([]byte(msg)); err != nil {
			t.Fatalf("write[%d]: %v", i, err)
		}

		buf := make([]byte, 256)
		clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := clientConn.Read(buf)
		if err != nil {
			t.Fatalf("read[%d]: %v", i, err)
		}
		if got := string(buf[:n]); got != msg {
			t.Errorf("echo[%d] = %q, want %q", i, got, msg)
		}
	}

	clientConn.Close()
	wg.Wait()
}

func TestHandlePortForwardConn_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(echoWSHandler))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithCancel(context.Background())

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		handlePortForwardConn(ctx, serverConn, wsURL, http.Header{})
		close(done)
	}()

	// Verify the tunnel works before cancellation.
	if _, err := clientConn.Write([]byte("pre-cancel")); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf := make([]byte, 256)
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := string(buf[:n]); got != "pre-cancel" {
		t.Errorf("echo = %q, want %q", got, "pre-cancel")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("handlePortForwardConn did not exit after context cancellation")
	}
}

func TestHandlePortForwardConn_DialFailure(t *testing.T) {
	ctx := context.Background()

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		handlePortForwardConn(ctx, serverConn, "ws://127.0.0.1:1", http.Header{})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handlePortForwardConn did not exit after dial failure")
	}
}

// ---------------------------------------------------------------------------
// Concurrent connections through runPortForward
// ---------------------------------------------------------------------------

func TestRunPortForward_ConcurrentConns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(echoWSHandler))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	// We can't call runPortForward directly (it needs resolveProject), so
	// test the handler concurrently to validate the same code path.
	const numConns = 10
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < numConns; i++ {
		clientConn, serverConn := net.Pipe()
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			handlePortForwardConn(ctx, serverConn, wsURL, http.Header{})
		}(i)

		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer clientConn.Close()

			msg := []byte("ping")
			if _, err := clientConn.Write(msg); err != nil {
				t.Errorf("conn %d write: %v", id, err)
				return
			}

			buf := make([]byte, 256)
			clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Errorf("conn %d read: %v", id, err)
				return
			}
			if got := string(buf[:n]); got != "ping" {
				t.Errorf("conn %d echo = %q, want %q", id, got, "ping")
			}
		}(i)
	}

	wg.Wait()
}
