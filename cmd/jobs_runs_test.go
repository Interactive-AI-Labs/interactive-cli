package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestJobRunCommandArgs(t *testing.T) {
	tests := []struct {
		name      string
		command   *cobra.Command
		args      []string
		wantError string
	}{
		{"describe by ID", jobRunDescribeCmd, []string{"run-id"}, ""},
		{"stop by ID", jobRunStopCmd, []string{"run-id"}, ""},
		{"delete by ID", jobRunDeleteCmd, []string{"run-id"}, ""},
		{"logs by ID", jobRunLogsCmd, []string{"run-id"}, ""},
		{"list by job", jobRunsListCmd, []string{"report"}, ""},
		{"logs by job", jobLogsCmd, []string{"report"}, ""},
		{
			"describe rejects redundant job name",
			jobRunDescribeCmd,
			[]string{"report", "run-id"},
			"accepts 1 arg(s), received 2",
		},
		{
			"stop rejects redundant job name",
			jobRunStopCmd,
			[]string{"report", "run-id"},
			"accepts 1 arg(s), received 2",
		},
		{
			"delete rejects redundant job name",
			jobRunDeleteCmd,
			[]string{"report", "run-id"},
			"accepts 1 arg(s), received 2",
		},
		{
			"logs rejects redundant job name",
			jobRunLogsCmd,
			[]string{"report", "run-id"},
			"accepts 1 arg(s), received 2",
		},
		{"missing run ID", jobRunDescribeCmd, nil, "accepts 1 arg(s), received 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.command.Args(tt.command, tt.args)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != tt.wantError {
				t.Fatalf("error = %q, want %q", message, tt.wantError)
			}
		})
	}
}
