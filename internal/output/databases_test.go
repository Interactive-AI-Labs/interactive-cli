package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/google/go-cmp/cmp"
)

func TestPrintDatabaseList(t *testing.T) {
	tests := []struct {
		name      string
		databases []clients.DatabaseOutput
		want      string
	}{
		{
			name:      "empty list",
			databases: []clients.DatabaseOutput{},
			want:      "No databases found.\n",
		},
		{
			name:      "nil list",
			databases: nil,
			want:      "No databases found.\n",
		},
		{
			name: "single database",
			databases: []clients.DatabaseOutput{
				{
					Name:     "my-db",
					Revision: 1,
					Status:   "Healthy",
					Updated:  "",
				},
			},
			want: "NAME    REVISION   STATUS    UPDATED\n" +
				"my-db   1          Healthy   \n",
		},
		{
			name: "multiple databases",
			databases: []clients.DatabaseOutput{
				{
					Name:     "db-alpha",
					Revision: 3,
					Status:   "Healthy",
				},
				{
					Name:     "db-beta",
					Revision: 1,
					Status:   "Provisioning",
				},
			},
			want: "NAME       REVISION   STATUS         UPDATED\n" +
				"db-alpha   3          Healthy        \n" +
				"db-beta    1          Provisioning   \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintDatabaseList(&buf, tt.databases)
			if err != nil {
				t.Fatalf("PrintDatabaseList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintDatabaseDescribe(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")

	tests := []struct {
		name string
		db   *clients.DescribeDatabaseResponse
		want string
	}{
		{
			name: "full describe output",
			db: &clients.DescribeDatabaseResponse{
				Name:             "my-db",
				Revision:         2,
				Status:           "Healthy",
				Message:          "All instances running",
				Updated:          "2025-01-15T10:30:00Z",
				PostgresVersion:  "17",
				Replicas:         2,
				HighAvailability: true,
				Resources:        clients.Resources{CPU: "1", Memory: "2G"},
				Storage:          clients.DatabaseStorageConfig{Size: "20G"},
				Extensions:       []string{"vector", "pg_trgm"},
				Backup: clients.DatabaseBackupStatus{
					Enabled: true,
					DatabaseBackupConfig: clients.DatabaseBackupConfig{
						Schedule:        "0 0 2 * * *",
						RetentionPolicy: "30d",
					},
				},
				StackId:           "my-stack",
				CredentialsSecret: "my-db-app",
			},
			want: "Name:                 my-db\n" +
				"Stack Id:             my-stack\n" +
				"Revision:             2\n" +
				"Status:               Healthy\n" +
				"Message:              All instances running\n" +
				"Updated:              2025-01-15 11:30:00 CET\n" +
				"PostgreSQL Version:   17\n" +
				"Replicas:             2\n" +
				"High Availability:    Yes\n" +
				"Resources:\n" +
				"  CPU:      1\n" +
				"  Memory:   2G\n" +
				"Storage:\n" +
				"  Size:       20G\n" +
				"Extensions:   vector, pg_trgm\n" +
				"\n" +
				"Backup:\n" +
				"  Schedule:           0 0 2 * * *\n" +
				"  Retention:          30d\n" +
				"Credentials Secret:   my-db-app\n",
		},
		{
			name: "HA independent of replica count",
			db: &clients.DescribeDatabaseResponse{
				Name:              "flexible-db",
				Revision:          1,
				Status:            "Healthy",
				PostgresVersion:   "17",
				Replicas:          2,
				HighAvailability:  false,
				Resources:         clients.Resources{CPU: "0.5", Memory: "1G"},
				Storage:           clients.DatabaseStorageConfig{Size: "10G"},
				CredentialsSecret: "flexible-db-app",
			},
			want: "Name:                 flexible-db\n" +
				"Revision:             1\n" +
				"Status:               Healthy\n" +
				"PostgreSQL Version:   17\n" +
				"Replicas:             2\n" +
				"High Availability:    No\n" +
				"Resources:\n" +
				"  CPU:      0.5\n" +
				"  Memory:   1G\n" +
				"Storage:\n" +
				"  Size:   10G\n" +
				"\n" +
				"Backup:               not configured\n" +
				"Credentials Secret:   flexible-db-app\n",
		},
		{
			name: "without backup configuration",
			db: &clients.DescribeDatabaseResponse{
				Name:              "basic-db",
				Revision:          1,
				Status:            "Provisioning",
				PostgresVersion:   "17",
				Replicas:          1,
				Resources:         clients.Resources{CPU: "0.5", Memory: "1G"},
				Storage:           clients.DatabaseStorageConfig{Size: "10G"},
				CredentialsSecret: "basic-db-app",
			},
			want: "Name:                 basic-db\n" +
				"Revision:             1\n" +
				"Status:               Provisioning\n" +
				"PostgreSQL Version:   17\n" +
				"Replicas:             1\n" +
				"High Availability:    No\n" +
				"Resources:\n" +
				"  CPU:      0.5\n" +
				"  Memory:   1G\n" +
				"Storage:\n" +
				"  Size:   10G\n" +
				"\n" +
				"Backup:               not configured\n" +
				"Credentials Secret:   basic-db-app\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintDatabaseDescribe(&buf, tt.db)
			if err != nil {
				t.Fatalf("PrintDatabaseDescribe() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
				t.Errorf("output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPrintDatabaseBackups(t *testing.T) {
	tests := []struct {
		name    string
		backups []clients.BackupOutput
		want    string
	}{
		{
			name:    "empty list",
			backups: []clients.BackupOutput{},
			want:    "No backups found.\n",
		},
		{
			name:    "nil list",
			backups: nil,
			want:    "No backups found.\n",
		},
		{
			name: "single backup",
			backups: []clients.BackupOutput{
				{
					Name:  "my-db-on-demand-1234",
					Phase: "completed",
				},
			},
			want: "NAME                   PHASE       STARTED   STOPPED   ERROR\n" +
				"my-db-on-demand-1234   completed                       \n",
		},
		{
			name: "backup with error",
			backups: []clients.BackupOutput{
				{
					Name:  "my-db-on-demand-5678",
					Phase: "failed",
					Error: "backup failed: timeout",
				},
			},
			want: "NAME                   PHASE    STARTED   STOPPED   ERROR\n" +
				"my-db-on-demand-5678   failed                       backup failed: timeout\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintDatabaseBackups(&buf, tt.backups)
			if err != nil {
				t.Fatalf("PrintDatabaseBackups() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
