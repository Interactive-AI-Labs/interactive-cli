package deployment

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDatabaseBackupPath(t *testing.T) {
	for _, tt := range []struct {
		name, databaseName, backupName, want string
	}{
		{
			"plain",
			"my-db",
			"my-db-on-demand-1",
			"/v1/organizations/org%2F1/projects/project%202/databases/my-db/backups/my-db-on-demand-1",
		},
		{
			"escaped",
			"a/b",
			"c d?",
			"/v1/organizations/org%2F1/projects/project%202/databases/a%2Fb/backups/c%20d%3F",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := databaseBackupPath("org/1", "project 2", tt.databaseName, tt.backupName)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
