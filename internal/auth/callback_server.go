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
<svg xmlns="http://www.w3.org/2000/svg" width="111" height="16" viewBox="0 0 111 16" fill="none">
<path d="M25.9749 3.00171V12.7421H24.1699V3.00171H25.9749Z" fill="#0D121C"/>
<path d="M35.3493 8.54338V12.7368H33.6384V8.63251C33.6384 7.90892 33.4757 7.35906 33.152 6.98141C32.8283 6.60527 32.3516 6.41644 31.7233 6.41644C31.0536 6.41644 30.5274 6.63699 30.1463 7.07808C29.7636 7.51918 29.5739 8.12645 29.5739 8.89988V12.7368H27.8789V5.12184H29.3538L29.542 6.11885C30.1495 5.39526 31.0074 5.03271 32.1172 5.03271C33.0691 5.03271 33.8472 5.31067 34.4484 5.86506C35.0495 6.42097 35.3509 7.31223 35.3509 8.54187L35.3493 8.54338Z" fill="#0D121C"/>
<path d="M38.636 3.00029V5.12723H40.583V6.53965H38.636V10.511C38.636 10.8086 38.6982 11.0186 38.8242 11.1425C38.9502 11.2664 39.1638 11.3283 39.4684 11.3283H40.7552V12.7407H39.1224C38.3586 12.7407 37.8037 12.5715 37.4577 12.2347C37.1117 11.8978 36.9395 11.3766 36.9395 10.6727V2.99878H38.6344L38.636 3.00029Z" fill="#0D121C"/>
<path d="M43.0451 5.51831C43.6367 5.19655 44.3143 5.03491 45.0781 5.03491C45.8419 5.03491 46.5355 5.18295 47.1271 5.48054C47.7187 5.77813 48.1843 6.19959 48.5239 6.74493C48.8635 7.29026 49.0389 7.92924 49.0501 8.6634C49.0501 8.86129 49.0342 9.06522 49.0023 9.27369H42.943V9.36281C42.9845 10.0275 43.2045 10.5532 43.6032 10.9399C44.0002 11.3266 44.5296 11.52 45.1881 11.52C45.7111 11.52 46.1512 11.4036 46.5068 11.171C46.8624 10.9384 47.0984 10.6091 47.2132 10.1816H48.9082C48.7615 10.955 48.366 11.5895 47.7234 12.0849C47.0793 12.5804 46.2772 12.8282 45.3141 12.8282C44.477 12.8282 43.7467 12.6665 43.1248 12.3448C42.5029 12.023 42.0214 11.5683 41.6802 10.9837C41.3405 10.3991 41.1699 9.71932 41.1699 8.94589C41.1699 8.17245 41.3342 7.47606 41.6642 6.88541C41.9943 6.29627 42.4551 5.84007 43.0451 5.5168V5.51831ZM46.6248 6.7978C46.2326 6.4851 45.7383 6.32951 45.1419 6.32951C44.587 6.32951 44.1086 6.49114 43.7052 6.8129C43.3018 7.13466 43.069 7.56368 43.0068 8.09994H47.3232C47.2499 7.54404 47.0171 7.11049 46.6248 6.79931V6.7978Z" fill="#0D121C"/>
<path d="M54.8017 6.61878H54.0475C53.3459 6.61878 52.8389 6.8348 52.5247 7.26532C52.2106 7.69585 52.0544 8.2442 52.0544 8.90887V12.7307H50.3594V5.11572H51.8662L52.0544 6.26077C52.284 5.90426 52.5822 5.62329 52.9489 5.42087C53.3156 5.21845 53.8068 5.11572 54.4238 5.11572H54.8001V6.61727L54.8017 6.61878Z" fill="#0D121C"/>
<path d="M63.4244 12.7368H61.9335L61.7454 11.5767C61.4632 11.9528 61.1012 12.2565 60.6627 12.4846C60.2242 12.7127 59.7108 12.826 59.124 12.826C58.3905 12.826 57.7367 12.6704 57.1611 12.3577C56.5855 12.045 56.1358 11.5948 55.8105 11.0042C55.4853 10.4135 55.3242 9.72317 55.3242 8.93009C55.3242 8.13702 55.4885 7.47537 55.8185 6.88473C56.1486 6.29559 56.6014 5.83938 57.1771 5.51611C57.7527 5.19435 58.4017 5.03271 59.124 5.03271C59.7203 5.03271 60.2386 5.13997 60.6787 5.35297C61.1187 5.56596 61.4743 5.86053 61.7454 6.23818L61.9495 5.12335H63.4244V12.7383V12.7368ZM61.7454 8.9588C61.7454 8.205 61.5301 7.59018 61.1012 7.11434C60.6723 6.63849 60.1062 6.39982 59.4062 6.39982C58.7062 6.39982 58.1402 6.63548 57.7112 7.10679C57.2823 7.5781 57.067 8.18536 57.067 8.92858C57.067 9.6718 57.2823 10.2942 57.7112 10.7655C58.1402 11.2368 58.7062 11.4725 59.4062 11.4725C60.1062 11.4725 60.6723 11.2398 61.1012 10.773C61.5301 10.3063 61.7454 9.70202 61.7454 8.9588Z" fill="#0D121C"/>
<path d="M68.7464 12.8243C67.9507 12.8243 67.2491 12.6627 66.6432 12.3409C66.0357 12.0191 65.5685 11.5599 65.2384 10.9647C64.9084 10.3695 64.7441 9.68524 64.7441 8.91181C64.7441 8.13838 64.9116 7.4586 65.2464 6.874C65.5813 6.28939 66.0548 5.83469 66.6671 5.51293C67.2794 5.19117 67.989 5.02954 68.7942 5.02954C69.7988 5.02954 70.62 5.2803 71.2594 5.78032C71.8972 6.28033 72.2958 6.96312 72.4521 7.82568H70.7093C70.5945 7.38911 70.3633 7.04167 70.0188 6.78487C69.6728 6.52656 69.2503 6.39815 68.748 6.39815C68.0671 6.39815 67.5218 6.63079 67.1072 7.09757C66.6943 7.56435 66.487 8.16859 66.487 8.91181C66.487 9.65503 66.6943 10.2774 67.1072 10.7487C67.5202 11.22 68.0671 11.4557 68.748 11.4557C69.271 11.4557 69.7063 11.3243 70.0507 11.0614C70.3968 10.7986 70.6216 10.4436 70.7252 9.99794H72.4521C72.2958 10.8801 71.8924 11.572 71.2435 12.072C70.5945 12.572 69.7621 12.8228 68.748 12.8228L68.7464 12.8243Z" fill="#0D121C"/>
<path d="M75.4563 3.00029V5.12723H77.4033V6.53965H75.4563V10.511C75.4563 10.8086 75.5185 11.0186 75.6445 11.1425C75.7705 11.2664 75.9841 11.3283 76.2887 11.3283H77.5755V12.7407H75.9427C75.1789 12.7407 74.624 12.5715 74.278 12.2347C73.932 11.8978 73.7598 11.3766 73.7598 10.6727V2.99878H75.4548L75.4563 3.00029Z" fill="#0D121C"/>
<path d="M80.2653 5.11743V12.7324H78.5703V5.11743H80.2653Z" fill="#0D121C"/>
<path d="M82.9475 5.11743L85.1608 11.0073L87.3437 5.11743H89.1327L86.1653 12.7324H84.0924L81.125 5.11743H82.946H82.9475Z" fill="#0D121C"/>
<path d="M91.1818 5.51831C91.7734 5.19655 92.4511 5.03491 93.2148 5.03491C93.9786 5.03491 94.6722 5.18295 95.2638 5.48054C95.8554 5.77813 96.321 6.19959 96.6606 6.74493C97.0003 7.29026 97.1757 7.92924 97.1868 8.6634C97.1868 8.86129 97.1709 9.06522 97.139 9.27369H91.0798V9.36281C91.1212 10.0275 91.3413 10.5532 91.7399 10.9399C92.1369 11.3266 92.6663 11.52 93.3249 11.52C93.8479 11.52 94.288 11.4036 94.6435 11.171C94.9991 10.9384 95.2351 10.6091 95.3499 10.1816H97.0449C96.8982 10.955 96.5028 11.5895 95.8602 12.0849C95.216 12.5804 94.4139 12.8282 93.4508 12.8282C92.6137 12.8282 91.8834 12.6665 91.2615 12.3448C90.6397 12.023 90.1581 11.5683 89.8169 10.9837C89.4773 10.3991 89.3066 9.71932 89.3066 8.94589C89.3066 8.17245 89.4709 7.47606 89.801 6.88541C90.131 6.29627 90.5918 5.84007 91.1818 5.5168V5.51831ZM94.7615 6.7978C94.3693 6.4851 93.875 6.32951 93.2786 6.32951C92.7237 6.32951 92.2454 6.49114 91.842 6.8129C91.4385 7.13466 91.2057 7.56368 91.1435 8.09994H95.4599C95.3866 7.54404 95.1538 7.11049 94.7615 6.79931V6.7978Z" fill="#0D121C"/>
<path d="M103.803 3.00122H101.652L97.8027 12.7431H99.6715L100.427 10.8398H104.933L105.7 12.7431H107.636L103.803 3.00122ZM100.93 9.38053L102.672 4.84416L104.431 9.38053H100.93Z" fill="#0D121C"/>
<path d="M110.715 3.00122V12.7416H108.91V3.00122H110.715Z" fill="#0D121C"/>
<path d="M0 11.1937V4.79468L2.86229 5.47229C3.16456 5.54418 3.37727 5.80226 3.37727 6.09806V9.89035C3.37727 10.1861 3.16456 10.4442 2.86229 10.5161L0 11.1937Z" fill="#E80F13"/>
<path d="M13.5605 9.89035V6.09806C13.5605 5.80226 13.7733 5.54418 14.0755 5.47229L16.9378 4.79468V11.1937L14.0755 10.5161C13.7733 10.4442 13.5605 10.1861 13.5605 9.89035Z" fill="#E80F13"/>
<path d="M3.34109 0H6.71836V2.66333C6.71836 2.9603 6.97212 3.20071 7.28559 3.20071H9.52965C9.84312 3.20071 10.0969 2.9603 10.0969 2.66333V0H13.4742V1.89615C13.4742 2.19194 13.2614 2.45003 12.9592 2.52191L10.526 3.09818C10.2735 3.15828 10.0969 3.37276 10.0969 3.61906V12.3809C10.0969 12.6272 10.2735 12.8417 10.526 12.9018L12.9592 13.4781C13.2614 13.55 13.4742 13.8081 13.4742 14.1039V16H10.0956V13.3367C10.0956 13.0397 9.84187 12.7993 9.5284 12.7993H7.28435C6.97088 12.7993 6.71712 13.0397 6.71712 13.3367V16H3.33984V14.1039C3.33984 13.8081 3.55256 13.55 3.85483 13.4781L6.28796 12.9018C6.54048 12.8417 6.71712 12.6272 6.71712 12.3809V3.61906C6.71712 3.37276 6.54048 3.15828 6.28796 3.09818L3.85483 2.52191C3.55256 2.45003 3.33984 2.19194 3.33984 1.89615V0H3.34109Z" fill="#E80F13"/>
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
