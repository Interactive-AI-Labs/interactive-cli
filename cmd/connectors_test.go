package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/spf13/cobra"
)

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

func TestConnectorEmptyArgGuards(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		args []string
		want string
	}{
		{"describe", connectorGetCmd, []string{"  "}, "connector id is required"},
		{"verify", connectorVerifyCmd, []string{"  "}, "connector id is required"},
		{"delete", connectorDeleteCmd, []string{"  "}, "connector id is required"},
		{"create", connectorCreateCmd, []string{"  "}, "connector name is required"},
		{
			"run-tool empty id",
			connectorRunToolCmd,
			[]string{"  ", "search"},
			"connector id is required",
		},
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

// Regression guard: run-tool previously always exited 0, so a failed call could be silently chained with '&&'.
func TestEmitToolResult(t *testing.T) {
	t.Run("ok status prints result and returns nil", func(t *testing.T) {
		var out bytes.Buffer
		res := &clients.McpToolCallData{Status: "ok", Result: json.RawMessage(`{"hits":3}`)}
		if err := emitToolResult(&out, res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "hits") {
			t.Fatalf("result not printed: %q", out.String())
		}
	})

	t.Run("error status returns error and prints nothing", func(t *testing.T) {
		var out bytes.Buffer
		res := &clients.McpToolCallData{
			Status:       "error",
			ErrorClass:   "tool_error",
			ErrorMessage: "boom",
		}
		err := emitToolResult(&out, res)
		if err == nil {
			t.Fatal("expected a non-nil error for a failed tool call")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("error missing detail: %v", err)
		}
		if out.Len() != 0 {
			t.Fatalf("nothing should be printed on failure, got: %q", out.String())
		}
	})
}

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
