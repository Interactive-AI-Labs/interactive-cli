package inputs

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

type DatabaseInput struct {
	Instances       int
	PostgresVersion string
	CPU             string
	Memory          string
	StorageSize     string
	Extensions      []string
	BackupSchedule  string
	BackupRetention string
}

type RestoreInput struct {
	DatabaseInput
	SourceDatabase string
	TargetTime     string
}

func BuildDatabaseRequestBody(in DatabaseInput) (clients.CreateDatabaseBody, error) {
	body := clients.CreateDatabaseBody{
		Instances:       in.Instances,
		PostgresVersion: in.PostgresVersion,
		Resources: clients.Resources{
			CPU:    in.CPU,
			Memory: in.Memory,
		},
		Storage: clients.DatabaseStorageConfig{
			Size: in.StorageSize,
		},
		Extensions: in.Extensions,
	}

	if in.BackupSchedule != "" || in.BackupRetention != "" {
		body.Backup = &clients.DatabaseBackupConfig{
			Schedule:        in.BackupSchedule,
			RetentionPolicy: in.BackupRetention,
		}
	}

	return body, nil
}

func BuildRestoreRequestBody(in RestoreInput) (clients.RestoreDatabaseBody, error) {
	base, err := BuildDatabaseRequestBody(in.DatabaseInput)
	if err != nil {
		return clients.RestoreDatabaseBody{}, err
	}

	return clients.RestoreDatabaseBody{
		CreateDatabaseBody: base,
		SourceDatabase:     in.SourceDatabase,
		TargetTime:         in.TargetTime,
	}, nil
}

var DatabaseUpdateFlags = struct {
	Instances       string
	PostgresVersion string
	CPU             string
	Memory          string
	StorageSize     string
	Extensions      string
	BackupSchedule  string
	BackupRetention string
}{
	Instances:       "instances",
	PostgresVersion: "postgres-version",
	CPU:             "cpu",
	Memory:          "memory",
	StorageSize:     "storage-size",
	Extensions:      "extensions",
	BackupSchedule:  "backup-schedule",
	BackupRetention: "backup-retention",
}

func BuildDatabaseUpdatePatch(
	in DatabaseInput,
	clearBackup bool,
	changed func(string) bool,
) (clients.UpdatePatch, error) {
	f := DatabaseUpdateFlags
	patch := clients.UpdatePatch{}

	backupFlags := []string{f.BackupSchedule, f.BackupRetention}
	if clearBackup && anyChanged(changed, backupFlags...) {
		return nil, fmt.Errorf(
			"--clear-backup cannot be combined with --backup-schedule or --backup-retention",
		)
	}

	if changed(f.Instances) {
		if err := setJSON(patch, "instances", in.Instances); err != nil {
			return nil, err
		}
	}

	if changed(f.PostgresVersion) {
		if err := setJSON(patch, "postgresVersion", in.PostgresVersion); err != nil {
			return nil, err
		}
	}

	if anyChanged(changed, f.CPU, f.Memory) {
		res := map[string]any{}
		if changed(f.CPU) {
			res["cpu"] = in.CPU
		}
		if changed(f.Memory) {
			res["memory"] = in.Memory
		}
		if err := setJSON(patch, "resources", res); err != nil {
			return nil, err
		}
	}

	if changed(f.StorageSize) {
		if err := setJSON(patch, "storage", map[string]any{"size": in.StorageSize}); err != nil {
			return nil, err
		}
	}

	if changed(f.Extensions) {
		if err := setJSON(patch, "extensions", in.Extensions); err != nil {
			return nil, err
		}
	}

	switch {
	case clearBackup:
		patch["backup"] = []byte("null")
	case anyChanged(changed, backupFlags...):
		backup := map[string]any{}
		if changed(f.BackupSchedule) {
			backup["schedule"] = in.BackupSchedule
		}
		if changed(f.BackupRetention) {
			backup["retentionPolicy"] = in.BackupRetention
		}
		if err := setJSON(patch, "backup", backup); err != nil {
			return nil, err
		}
	}

	return patch, nil
}
