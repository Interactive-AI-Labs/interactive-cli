package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/gorilla/websocket"
)

type portForwardOpts struct {
	resourceType string // "services", "agents", "databases"
	resourceName string
	remotePort   int
	localPort    int
	org          string
	project      string
}

func runPortForward(cmdCtx context.Context, opts portForwardOpts) error {
	pCtx, _, _, err := resolveProject(cmdCtx, opts.org, opts.project)
	if err != nil {
		return err
	}

	wsURL, err := buildPortForwardURL(
		deploymentHostname,
		pCtx.orgId,
		pCtx.projectId,
		opts.resourceType,
		opts.resourceName,
		opts.remotePort,
	)
	if err != nil {
		return err
	}

	headers, err := authHeaders()
	if err != nil {
		return err
	}

	// Port 0 makes the OS assign a random available port.
	localAddr := fmt.Sprintf("127.0.0.1:%d", opts.localPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}
	defer listener.Close()

	ctx, stop := signal.NotifyContext(cmdCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	displayPort := opts.localPort
	if displayPort == 0 {
		// OS assigned the port; read it back from the listener.
		_, p, _ := net.SplitHostPort(listener.Addr().String())
		fmt.Sscanf(p, "%d", &displayPort)
	}
	displayAddr := fmt.Sprintf("localhost:%d", displayPort)

	if opts.remotePort > 0 {
		fmt.Fprintf(os.Stderr, "Forwarding %s → %s/%s (port %d)\n",
			displayAddr, opts.resourceType, opts.resourceName, opts.remotePort)
	} else {
		fmt.Fprintf(os.Stderr, "Forwarding %s → %s/%s\n",
			displayAddr, opts.resourceType, opts.resourceName)
	}
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n")

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	// Accept connections in a loop; each gets its own WS tunnel.
	var wg sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("accept error: %v", err)
			continue
		}
		wg.Go(func() {
			handlePortForwardConn(ctx, conn, wsURL, headers)
		})
	}

	wg.Wait()
	return nil
}

func handlePortForwardConn(
	ctx context.Context,
	tcpConn net.Conn,
	wsURL string,
	headers http.Header,
) {
	defer tcpConn.Close()

	dialer := websocket.Dialer{}
	wsConn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil && resp != nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("websocket dial failed (%s): %s", resp.Status, string(body))
		return
	}
	if err != nil {
		log.Printf("websocket dial failed: %v", err)
		return
	}
	defer wsConn.Close()

	// wsMu serializes WebSocket writes: gorilla/websocket does not
	// support concurrent writers, so WriteMessage and WriteControl
	// must not overlap.
	var wsMu sync.Mutex
	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			wsMu.Lock()
			deadline := time.Now().Add(time.Second)
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			wsConn.WriteControl(websocket.CloseMessage, msg, deadline)
			wsMu.Unlock()
			wsConn.Close()
			tcpConn.Close()
		})
	}

	connDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdown()
		case <-connDone:
		}
	}()

	done := make(chan struct{})

	// TCP → WS
	go func() {
		defer close(done)
		buf := make([]byte, 32*1024)
		for {
			n, readErr := tcpConn.Read(buf)
			if n > 0 {
				wsMu.Lock()
				writeErr := wsConn.WriteMessage(
					websocket.BinaryMessage,
					buf[:n],
				)
				wsMu.Unlock()
				if writeErr != nil {
					return
				}
			}
			if readErr != nil {
				shutdown()
				return
			}
		}
	}()

	// WS → TCP
	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			break
		}
		if _, err := tcpConn.Write(msg); err != nil {
			break
		}
	}

	tcpConn.Close()
	<-done
	close(connDone)
}

func buildPortForwardURL(
	host, orgId, projectId, resourceType, resourceName string,
	port int,
) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("failed to parse deployment hostname: %w", err)
	}

	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else {
		u.Scheme = "wss"
	}

	u.RawPath = fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/%s/%s/port-forward",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		resourceType,
		url.PathEscape(resourceName),
	)
	unescaped, err := url.PathUnescape(u.RawPath)
	if err != nil {
		return "", fmt.Errorf("failed to unescape path: %w", err)
	}
	u.Path = unescaped

	if port > 0 {
		q := u.Query()
		q.Set("port", fmt.Sprintf("%d", port))
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

func authHeaders() (http.Header, error) {
	cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	req := &http.Request{Header: http.Header{}}
	if err := clients.ApplyAuth(req, token, apiKey, cookies); err != nil {
		return nil, err
	}

	return req.Header, nil
}
