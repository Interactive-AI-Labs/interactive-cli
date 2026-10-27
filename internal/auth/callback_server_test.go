package auth

import (
	"net/url"
	"strconv"
	"testing"
)

func TestRedirectURI(t *testing.T) {
	cs, err := NewCallbackServer()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, ln := range cs.listeners {
			ln.Close()
		}
	})

	got, err := url.Parse(cs.RedirectURI())
	if err != nil {
		t.Fatalf("RedirectURI() = %q is not a URL: %v", cs.RedirectURI(), err)
	}

	tests := []struct {
		part string
		got  string
		want string
	}{
		{part: "scheme", got: got.Scheme, want: "http"},
		{part: "hostname", got: got.Hostname(), want: "localhost"},
		{part: "port", got: got.Port(), want: strconv.Itoa(cs.Port())},
		{part: "path", got: got.Path, want: "/callback"},
	}
	for _, tt := range tests {
		t.Run(tt.part, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("RedirectURI() %s = %q, want %q", tt.part, tt.got, tt.want)
			}
		})
	}
}
