package inputs

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/google/go-cmp/cmp"
)

func TestBuildDatabaseRequestBody(t *testing.T) {
	tests := []struct {
		name string
		in   DatabaseInput
		want clients.CreateDatabaseBody
	}{
		{
			name: "basic fields",
			in: DatabaseInput{
				Instances:   2,
				CPU:         "1",
				Memory:      "2G",
				StorageSize: "20G",
			},
			want: clients.CreateDatabaseBody{
				Instances: 2,
				Resources: clients.Resources{CPU: "1", Memory: "2G"},
				Storage:   clients.DatabaseStorageConfig{Size: "20G"},
			},
		},
		{
			name: "all fields with backup",
			in: DatabaseInput{
				Instances:       2,
				PostgresVersion: "16",
				CPU:             "2",
				Memory:          "4G",
				StorageSize:     "50G",
				Extensions:      []string{"vector", "pg_trgm"},
				BackupSchedule:  "0 0 2 * * *",
				BackupRetention: "30d",
			},
			want: clients.CreateDatabaseBody{
				Instances:       2,
				PostgresVersion: "16",
				Resources:       clients.Resources{CPU: "2", Memory: "4G"},
				Storage:         clients.DatabaseStorageConfig{Size: "50G"},
				Extensions:      []string{"vector", "pg_trgm"},
				Backup: &clients.DatabaseBackupConfig{
					Schedule:        "0 0 2 * * *",
					RetentionPolicy: "30d",
				},
			},
		},
		{
			name: "no backup when schedule and retention are empty",
			in: DatabaseInput{
				Instances:   1,
				CPU:         "0.5",
				Memory:      "1G",
				StorageSize: "10G",
			},
			want: clients.CreateDatabaseBody{
				Instances: 1,
				Resources: clients.Resources{CPU: "0.5", Memory: "1G"},
				Storage:   clients.DatabaseStorageConfig{Size: "10G"},
			},
		},
		{
			name: "partial backup schedule only",
			in: DatabaseInput{
				Instances:      1,
				CPU:            "1",
				Memory:         "2G",
				StorageSize:    "20G",
				BackupSchedule: "0 0 3 * * *",
			},
			want: clients.CreateDatabaseBody{
				Instances: 1,
				Resources: clients.Resources{CPU: "1", Memory: "2G"},
				Storage:   clients.DatabaseStorageConfig{Size: "20G"},
				Backup: &clients.DatabaseBackupConfig{
					Schedule: "0 0 3 * * *",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildDatabaseRequestBody(tt.in)
			if err != nil {
				t.Fatalf("BuildDatabaseRequestBody() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildRestoreRequestBody(t *testing.T) {
	got, err := BuildRestoreRequestBody(RestoreInput{
		DatabaseInput: DatabaseInput{
			Instances:   2,
			CPU:         "1",
			Memory:      "2G",
			StorageSize: "20G",
		},
		SourceDatabase: "my-source-db",
		TargetTime:     "2026-05-12T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("BuildRestoreRequestBody() error = %v", err)
	}
	if got.SourceDatabase != "my-source-db" {
		t.Errorf("SourceDatabase = %q, want %q", got.SourceDatabase, "my-source-db")
	}
	if got.TargetTime != "2026-05-12T10:00:00Z" {
		t.Errorf("TargetTime = %q, want %q", got.TargetTime, "2026-05-12T10:00:00Z")
	}
	if got.Instances != 2 {
		t.Errorf("Instances = %d, want 2", got.Instances)
	}
}

func TestBuildDatabaseUpdatePatch(t *testing.T) {
	tests := []struct {
		name        string
		in          DatabaseInput
		clearBackup bool
		changed     map[string]bool
		wantKeys    []string
		wantErr     string
	}{
		{
			name:     "single field instances",
			in:       DatabaseInput{Instances: 3},
			changed:  map[string]bool{"instances": true},
			wantKeys: []string{"instances"},
		},
		{
			name:     "resources partial cpu only",
			in:       DatabaseInput{CPU: "2"},
			changed:  map[string]bool{"cpu": true},
			wantKeys: []string{"resources"},
		},
		{
			name:     "resources both cpu and memory",
			in:       DatabaseInput{CPU: "2", Memory: "4G"},
			changed:  map[string]bool{"cpu": true, "memory": true},
			wantKeys: []string{"resources"},
		},
		{
			name:     "storage size",
			in:       DatabaseInput{StorageSize: "50G"},
			changed:  map[string]bool{"storage-size": true},
			wantKeys: []string{"storage"},
		},
		{
			name:     "extensions",
			in:       DatabaseInput{Extensions: []string{"vector", "pg_trgm"}},
			changed:  map[string]bool{"extensions": true},
			wantKeys: []string{"extensions"},
		},
		{
			name:        "clear backup",
			clearBackup: true,
			changed:     map[string]bool{},
			wantKeys:    []string{"backup"},
		},
		{
			name:     "backup schedule only",
			in:       DatabaseInput{BackupSchedule: "0 0 2 * * *"},
			changed:  map[string]bool{"backup-schedule": true},
			wantKeys: []string{"backup"},
		},
		{
			name:        "clear backup conflicts with backup schedule",
			in:          DatabaseInput{BackupSchedule: "0 0 2 * * *"},
			clearBackup: true,
			changed:     map[string]bool{"backup-schedule": true},
			wantErr:     "--clear-backup cannot be combined",
		},
		{
			name:    "no flags produces empty patch",
			changed: map[string]bool{},
		},
		{
			name: "multiple fields",
			in:   DatabaseInput{Instances: 3, CPU: "2", Memory: "4G", StorageSize: "50G"},
			changed: map[string]bool{
				"instances":    true,
				"cpu":          true,
				"memory":       true,
				"storage-size": true,
			},
			wantKeys: []string{"instances", "resources", "storage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changedFn := func(name string) bool { return tt.changed[name] }

			got, err := BuildDatabaseUpdatePatch(tt.in, tt.clearBackup, false, changedFn)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(tt.wantKeys) == 0 {
				if len(got) != 0 {
					t.Errorf("expected empty patch, got %d keys", len(got))
				}
				return
			}

			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("patch missing key %q", key)
				}
			}
		})
	}
}

func TestBuildDatabaseUpdatePatchClearBackupIsNull(t *testing.T) {
	changedFn := func(string) bool { return false }
	got, err := BuildDatabaseUpdatePatch(DatabaseInput{}, true, false, changedFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	raw, ok := got["backup"]
	if !ok {
		t.Fatal("patch missing 'backup' key")
	}
	if string(raw) != "null" {
		t.Errorf("backup value = %q, want %q", string(raw), "null")
	}
}

func TestBuildDatabaseUpdatePatchResourcesPartial(t *testing.T) {
	changedFn := func(name string) bool { return name == "cpu" }
	got, err := BuildDatabaseUpdatePatch(DatabaseInput{CPU: "2"}, false, false, changedFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	raw, ok := got["resources"]
	if !ok {
		t.Fatal("patch missing 'resources' key")
	}
	var res map[string]any
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatalf("failed to unmarshal resources: %v", err)
	}
	if res["cpu"] != "2" {
		t.Errorf("cpu = %v, want %q", res["cpu"], "2")
	}
	if _, hasMemory := res["memory"]; hasMemory {
		t.Error("resources should not contain memory when only cpu changed")
	}
}
