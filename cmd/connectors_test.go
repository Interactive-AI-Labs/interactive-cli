package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Confirmation must honor input without a trailing newline and treat empty input
// as a decline, not an error.
func TestConfirmDeletion(t *testing.T) {
	tests := []struct {
		name   string
		stdin  string
		wantOK bool
	}{
		{"yes with newline", "y\n", true},
		{"yes without newline (EOF mid-line)", "y", true},
		{"uppercase yes", "Y\n", true},
		{"yes with surrounding space", "  y  \n", true},
		{"no", "n\n", false},
		{"empty stdin (bare EOF)", "", false},
		{"anything else declines", "yes please\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			ok, err := confirmDeletion(strings.NewReader(tt.stdin), &out, "some-id")
			if err != nil {
				t.Fatalf("stdin=%q unexpected error: %v", tt.stdin, err)
			}
			if ok != tt.wantOK {
				t.Fatalf("stdin=%q ok=%v wantOK=%v", tt.stdin, ok, tt.wantOK)
			}
			if !strings.Contains(out.String(), "some-id") {
				t.Fatalf("prompt did not mention the connector id: %q", out.String())
			}
		})
	}
}

// Each command must reject an empty or whitespace-only positional arg before
// doing any network work.
func TestConnectorEmptyArgGuards(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		args []string
		want string
	}{
		{"describe", connectorDescribeCmd, []string{"  "}, "connector id is required"},
		{"verify", connectorVerifyCmd, []string{"  "}, "connector id is required"},
		{"delete", connectorDeleteCmd, []string{"  "}, "connector id is required"},
		{"create", connectorCreateCmd, []string{"  "}, "connector name is required"},
		{"run-tool empty id", connectorRunToolCmd, []string{"  ", "search"}, "connector id is required"},
		{"run-tool empty tool", connectorRunToolCmd, []string{"c1", "  "}, "tool name is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.RunE(tc.cmd, tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
		})
	}
}

// A custom connector (no --catalog-id) must require --endpoint-url before any
// network work.
func TestConnectorCreateRequiresEndpointOrCatalog(t *testing.T) {
	t.Cleanup(func() {
		connectorCatalogID = ""
		connectorEndpointURL = ""
	})
	connectorCatalogID = ""
	connectorEndpointURL = ""

	err := connectorCreateCmd.RunE(connectorCreateCmd, []string{"my-connector"})
	if err == nil || !strings.Contains(err.Error(), "--endpoint-url is required") {
		t.Fatalf("got %v, want error requiring --endpoint-url", err)
	}
}
