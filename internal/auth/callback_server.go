package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// CallbackResult holds the authorization code and state received on the localhost callback.
type CallbackResult struct {
	Code  string
	State string
	Err   error
}

// CallbackServer is an ephemeral HTTP server that listens on 127.0.0.1 for
// the OAuth callback redirect from the browser.
type CallbackServer struct {
	listener net.Listener
	server   *http.Server
	result   chan CallbackResult
	once     sync.Once
}

// NewCallbackServer creates a callback server on a random available port on 127.0.0.1.
func NewCallbackServer() (*CallbackServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on 127.0.0.1: %w", err)
	}

	cs := &CallbackServer{
		listener: ln,
		result:   make(chan CallbackResult, 1),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", cs.handleCallback)

	cs.server = &http.Server{Handler: mux}

	return cs, nil
}

// Port returns the port the server is listening on.
func (cs *CallbackServer) Port() int {
	return cs.listener.Addr().(*net.TCPAddr).Port
}

// RedirectURI returns the full redirect URI for the callback.
func (cs *CallbackServer) RedirectURI() string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", cs.Port())
}

// Start begins serving in a goroutine.
func (cs *CallbackServer) Start() {
	go func() {
		if err := cs.server.Serve(cs.listener); err != nil && err != http.ErrServerClosed {
			cs.once.Do(func() {
				cs.result <- CallbackResult{Err: fmt.Errorf("callback server error: %w", err)}
			})
		}
	}()
}

// Wait blocks until a callback is received or the context is cancelled.
func (cs *CallbackServer) Wait(ctx context.Context) CallbackResult {
	select {
	case r := <-cs.result:
		return r
	case <-ctx.Done():
		return CallbackResult{Err: ctx.Err()}
	}
}

// Shutdown gracefully stops the server.
func (cs *CallbackServer) Shutdown(ctx context.Context) {
	_ = cs.server.Shutdown(ctx)
}

func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errParam := r.URL.Query().Get("error")

	cs.once.Do(func() {
		if code == "" {
			msg := "no authorization code in callback"
			if errParam != "" {
				msg = fmt.Sprintf("authorization error: %s", errParam)
			}
			cs.result <- CallbackResult{Err: fmt.Errorf("%s", msg)}
		} else {
			cs.result <- CallbackResult{Code: code, State: state}
		}
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// Determine which page to show based on this request's own query params
	// (no shared mutable state — safe for concurrent requests)
	if code == "" {
		fmt.Fprint(w, errorPageHTML)
	} else {
		fmt.Fprint(w, successPageHTML)
	}
}

const errorPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Login Failed - Interactive CLI</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;
display:flex;align-items:center;justify-content:center;min-height:100vh;
background:#f8fafc;color:#1e293b}
.card{background:#fff;border-radius:12px;box-shadow:0 1px 3px rgba(0,0,0,.1),0 1px 2px rgba(0,0,0,.06);
padding:48px;text-align:center;max-width:420px;width:100%}
.icon{width:64px;height:64px;border-radius:50%;background:#fde8e8;
display:flex;align-items:center;justify-content:center;margin:0 auto 24px}
.icon svg{width:32px;height:32px;color:#E80F13}
h1{font-size:20px;font-weight:600;margin-bottom:8px}
p{font-size:14px;color:#64748b;line-height:1.5}
</style>
</head>
<body>
<div class="card">
<div class="icon">
<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
</svg>
</div>
<h1>Login failed</h1>
<p>Authentication was not completed.<br>Please return to the terminal and try again.</p>
</div>
</body>
</html>`

const successPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Login Successful - Interactive CLI</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;
display:flex;align-items:center;justify-content:center;min-height:100vh;
background:#f8fafc;color:#1e293b}
.card{background:#fff;border-radius:12px;box-shadow:0 1px 3px rgba(0,0,0,.1),0 1px 2px rgba(0,0,0,.06);
padding:48px;text-align:center;max-width:420px;width:100%}
.logo{margin:0 auto 32px;width:160px}
.logo svg{width:100%;height:auto}
.check{width:64px;height:64px;border-radius:50%;background:#fde8e8;
display:flex;align-items:center;justify-content:center;margin:0 auto 24px}
.check svg{width:32px;height:32px;color:#E80F13}
h1{font-size:20px;font-weight:600;margin-bottom:8px}
p{font-size:14px;color:#64748b;line-height:1.5}
</style>
</head>
<body>
<div class="card">
<div class="logo">
<svg viewBox="0 0 254 39" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M59.5916 7.27V30.857H55.4507V7.27H59.5916Z" fill="#1e293b"/>
<path d="M81.096 20.688V30.843H77.17V20.904c0-1.752-.373-3.084-1.116-3.998-.742-.911-1.836-1.366-3.277-1.366-1.536 0-2.744.534-3.618 1.602-1.878 2.136-1.313 3.607-1.313 5.48v8.121H63.957V12.403h3.384l.432 2.414c1.393-1.752 3.361-2.63 5.907-2.63 2.184 0 3.97.673 5.349 2.016 1.379 1.346 2.07 3.504 2.07 6.482z" fill="#1e293b"/>
<path d="M0 27.108V11.612l6.567 1.64a1.55 1.55 0 011.181 1.516v9.183a1.55 1.55 0 01-1.181 1.516L0 27.108z" fill="#E80F13"/>
<path d="M31.107 23.952V14.768c0-.716.488-1.341 1.181-1.515l6.567-1.641v15.496l-6.567-1.641a1.55 1.55 0 01-1.181-1.515z" fill="#E80F13"/>
<path d="M7.664 0h7.748v6.45c0 .719.583 1.301 1.302 1.301h5.148c.72 0 1.301-.582 1.301-1.301V0h7.748v4.592c0 .716-.488 1.341-1.181 1.515l-5.582 1.396a1.55 1.55 0 00-1.181 1.515v21.217c0 .597.405 1.116.984 1.262l5.582 1.395c.694.174 1.181.8 1.181 1.516v4.591h-7.748v-6.449a1.301 1.301 0 00-1.301-1.301h-5.148a1.301 1.301 0 00-1.302 1.301v6.45H7.662v-4.592c0-.716.488-1.341 1.181-1.516l5.582-1.395a1.55 1.55 0 00.985-1.262V8.764a1.55 1.55 0 00-.985-1.261L8.843 6.107A1.55 1.55 0 017.662 4.59V0z" fill="#E80F13"/>
</svg>
</div>
<div class="check">
<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
</svg>
</div>
<h1>Login successful!</h1>
<p>You are now signed in to Interactive CLI.<br>You can close this tab.</p>
</div>
</body>
</html>`
