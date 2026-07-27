package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestImageDeleteValidation(t *testing.T) {
	originalTag := imageDeleteTag
	t.Cleanup(func() { imageDeleteTag = originalTag })

	tests := []struct {
		name string
		args []string
		tag  string
		want string
	}{
		{
			name: "empty image name",
			args: []string{" "},
			tag:  "v1",
			want: "image name is required",
		},
		{
			name: "empty tag",
			args: []string{"app"},
			tag:  " ",
			want: "tag is required; please provide --tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageDeleteTag = tt.tag
			err := imageDeleteCmd.RunE(imageDeleteCmd, tt.args)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if diff := cmp.Diff(tt.want, err.Error()); diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWarnExistingImageTag(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		imagesJSON string
		tag        string
		want       string
		wantPrefix string
	}{
		{
			name:       "existing tag warns",
			status:     http.StatusOK,
			imagesJSON: `{"images":[{"name":"app","tags":["0.2.35","0.2.36"]}]}`,
			tag:        "0.2.36",
			want:       "⚠ tag 0.2.36 already exists upstream — pushing replaces it; the previous image is unrecoverable.\n",
		},
		{
			name:       "new tag stays silent",
			status:     http.StatusOK,
			imagesJSON: `{"images":[{"name":"app","tags":["0.2.35"]}]}`,
			tag:        "0.2.36",
			want:       "",
		},
		{
			name:       "same tag on another image stays silent",
			status:     http.StatusOK,
			imagesJSON: `{"images":[{"name":"other","tags":["0.2.36"]}]}`,
			tag:        "0.2.36",
			want:       "",
		},
		{
			name:       "list failure fails open with a note",
			status:     http.StatusInternalServerError,
			imagesJSON: `{}`,
			tag:        "0.2.36",
			wantPrefix: "could not list existing image tags (",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
					fmt.Fprint(w, tt.imagesJSON)
				}),
			)
			t.Cleanup(server.Close)

			origHostname, origToken := deploymentHostname, token
			t.Cleanup(func() { deploymentHostname, token = origHostname, origToken })
			deploymentHostname, token = server.URL, "test-token"

			var buf bytes.Buffer
			warnExistingImageTag(context.Background(), &buf, "o1", "p1", "app", tt.tag, nil)

			got := buf.String()
			if tt.wantPrefix != "" {
				if !strings.HasPrefix(got, tt.wantPrefix) ||
					!strings.HasSuffix(got, "— proceeding without pre-flight check\n") {
					t.Errorf("output = %q, want fail-open note with prefix %q", got, tt.wantPrefix)
				}
				return
			}
			if got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
