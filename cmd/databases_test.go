package cmd

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDatabaseBackupsCommandRouting(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantPath string
		wantArgs []string
	}{
		{
			name:     "list keeps the database as its argument",
			args:     []string{"databases", "backups", "my-db"},
			wantPath: "iai databases backups",
			wantArgs: []string{"my-db"},
		},
		{
			name:     "delete takes the database and the backup",
			args:     []string{"databases", "backups", "delete", "my-db", "my-db-on-demand-1"},
			wantPath: "iai databases backups delete",
			wantArgs: []string{"my-db", "my-db-on-demand-1"},
		},
		{
			name:     "rm alias",
			args:     []string{"db", "backups", "rm", "my-db", "my-db-on-demand-1"},
			wantPath: "iai databases backups delete",
			wantArgs: []string{"my-db", "my-db-on-demand-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, rest, err := rootCmd.Find(tt.args)
			if err != nil {
				t.Fatalf("Find(%v) error = %v", tt.args, err)
			}
			if cmd.CommandPath() != tt.wantPath || !slices.Equal(rest, tt.wantArgs) {
				t.Fatalf(
					"Find(%v) = %q with args %v, want %q with args %v",
					tt.args,
					cmd.CommandPath(),
					rest,
					tt.wantPath,
					tt.wantArgs,
				)
			}
		})
	}
}

func TestDatabaseBackupDeleteValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "empty database name",
			args: []string{" ", "b-1"},
			want: "database name is required",
		},
		{
			name: "empty backup name",
			args: []string{"my-db", " "},
			want: "backup name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dbBackupDeleteCmd.RunE(dbBackupDeleteCmd, tt.args)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if diff := cmp.Diff(tt.want, err.Error()); diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
