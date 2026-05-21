package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	dbProject      string
	dbOrganization string

	dbInstances       int
	dbPostgresVersion string
	dbCPU             string
	dbMemory          string
	dbStorageSize     string
	dbExtensions      []string
	dbBackupSchedule  string
	dbBackupRetention string

	dbClearBackup  bool
	dbClearStackId bool

	dbStackId string

	dbSourceDatabase string
	dbTargetTime     string

	dbLogsFollow    bool
	dbLogsSince     string
	dbLogsStartTime string
	dbLogsEndTime   string
	dbLogsRaw       bool
	dbLogsFields    []string
	dbLogsAllFields bool
)

var databasesCmd = &cobra.Command{
	Use:     "databases",
	Aliases: []string{"database", "db"},
	Short:   "PostgreSQL instances with extension support, including pgvector",
	GroupID: groupInfra,
	Long: `Manage PostgreSQL databases in InteractiveAI projects.

Databases are managed PostgreSQL instances that can also be used as vector
stores. The "vector" extension (pgvector) is installed by default, enabling
vector similarity search for AI/ML workloads such as RAG and embeddings.

Each database automatically creates a secret named <database_name>-app with
connection credentials (host, port, username, password, URI).`,
}

var dbListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List databases in a project",
	Long:    `List databases in a project.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		databases, err := deployClient.ListDatabases(cmd.Context(), pCtx.orgId, pCtx.projectId, "")
		if err != nil {
			return err
		}

		return output.PrintDatabaseList(out, databases)
	},
}

var dbDescribeCmd = &cobra.Command{
	Use:     "describe <database_name>",
	Aliases: []string{"desc"},
	Short:   "Describe a database in detail",
	Long: `Show detailed information about a database including configuration, runtime
status, and connection credentials.

Examples:
  iai databases describe my-db`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		db, err := deployClient.DescribeDatabase(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
		)
		if err != nil {
			return err
		}

		return output.PrintDatabaseDescribe(out, db)
	},
}

var dbCreateCmd = &cobra.Command{
	Use:   "create <database_name>",
	Short: "Create a database in a project",
	Long: `Create a managed PostgreSQL database in a project.

The "vector" extension is installed by default. To add other extensions, use
--extensions. Values above 1 for --instances enable high availability.

Changing the PostgreSQL major version after creation causes cluster downtime
during the upgrade.

Examples:
  iai databases create my-db --instances 2 --cpu 1 --memory 2G --storage-size 20G
  iai databases create my-db --instances 1 --cpu 0.5 --memory 1G --storage-size 20G --extensions vector --extensions pg_trgm
  iai databases create my-db --instances 2 --cpu 1 --memory 2G --storage-size 50G --backup-schedule "0 0 2 * * *" --backup-retention 30d`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		reqBody, err := inputs.BuildDatabaseRequestBody(inputs.DatabaseInput{
			Instances:       dbInstances,
			PostgresVersion: dbPostgresVersion,
			CPU:             dbCPU,
			Memory:          dbMemory,
			StorageSize:     dbStorageSize,
			Extensions:      dbExtensions,
			BackupSchedule:  dbBackupSchedule,
			BackupRetention: dbBackupRetention,
			StackId:         dbStackId,
		})
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting database creation request...")

		serverMessage, err := deployClient.CreateDatabase(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
			reqBody,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var dbUpdateCmd = &cobra.Command{
	Use:   "update <database_name>",
	Short: "Update a database in a project",
	Long: `Partial update of a database. Only the flags you pass are applied; everything
else keeps its current value.

Storage can only be increased. Use --clear-backup to disable backups entirely.
Changing the PostgreSQL major version triggers an automatic upgrade with cluster
downtime.

Use --clear-stack-id to remove the database from its stack.

Examples:
  iai databases update my-db --instances 3
  iai databases update my-db --cpu 2 --memory 4G
  iai databases update my-db --storage-size 50G
  iai databases update my-db --backup-schedule "0 0 3 * * *" --backup-retention 60d
  iai databases update my-db --clear-backup
  iai databases update my-db --stack-id my-stack
  iai databases update my-db --clear-stack-id`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		patch, err := inputs.BuildDatabaseUpdatePatch(inputs.DatabaseInput{
			Instances:       dbInstances,
			PostgresVersion: dbPostgresVersion,
			CPU:             dbCPU,
			Memory:          dbMemory,
			StorageSize:     dbStorageSize,
			Extensions:      dbExtensions,
			BackupSchedule:  dbBackupSchedule,
			BackupRetention: dbBackupRetention,
			StackId:         dbStackId,
		}, dbClearBackup, dbClearStackId, cmd.Flags().Changed)
		if err != nil {
			return err
		}
		if len(patch) == 0 {
			return fmt.Errorf("no fields to update; pass at least one flag")
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting database update request...")

		serverMessage, err := deployClient.PatchDatabase(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
			patch,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var dbDeleteCmd = &cobra.Command{
	Use:     "delete <database_name>",
	Aliases: []string{"rm"},
	Short:   "Delete a database from a project",
	Long:    `Delete a database and all associated resources from a project.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting database deletion request...")

		serverMessage, err := deployClient.DeleteDatabase(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var dbLogsCmd = &cobra.Command{
	Use:   "logs <database_name>",
	Short: "Show logs for a database",
	Long: `Show logs for a database in a project.

Returns up to 5000 log entries in chronological order. Default lookback is 1h.

Structured (JSON) logs are automatically formatted: the level and message are
extracted and displayed. PostgreSQL-style logs wrapped in a "record" envelope
are unwrapped transparently.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		databaseName := strings.TrimSpace(args[0])
		if databaseName == "" {
			return fmt.Errorf("database name is required")
		}

		ctx := cmd.Context()
		if dbLogsFollow {
			var stop func()
			ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
		}

		timeout := 1 * time.Minute
		if dbLogsFollow {
			timeout = 0
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(), dbOrganization, dbProject,
			resolveOpts{deployTimeout: timeout},
		)
		if err != nil {
			return err
		}

		opts := clients.LogsOptions{
			Follow:    dbLogsFollow,
			Since:     dbLogsSince,
			StartTime: dbLogsStartTime,
			EndTime:   dbLogsEndTime,
		}

		logsResp, err := deployClient.GetDatabaseLogs(
			ctx,
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
			opts,
		)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		meta := output.LogsMeta{Since: logsResp.Since, Truncated: logsResp.Truncated}
		fmtOpts := output.LogFormatOptions{
			Raw:       dbLogsRaw,
			Fields:    dbLogsFields,
			AllFields: dbLogsAllFields,
		}
		err = output.PrintLogStream(out, logsResp.Body, true, meta, fmtOpts)
		if dbLogsFollow && ctx.Err() != nil {
			return nil
		}
		return err
	},
}

var dbLogFieldsSince string

var dbLogFieldsCmd = &cobra.Command{
	Use:   "log-fields <database_name>",
	Short: "List available fields in structured logs",
	Long: `Scan recent logs and list the extra fields present in structured (JSON) log entries.

Use the reported field names with 'iai databases logs --fields' to include them in output.

Examples:
  iai databases log-fields my-db
  iai databases log-fields my-db --since 1h`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		since := dbLogFieldsSince

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		opts := clients.LogsOptions{Since: since}
		logsResp, err := deployClient.GetDatabaseLogs(
			cmd.Context(), pCtx.orgId, pCtx.projectId, databaseName, opts,
		)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		fields := output.DiscoverLogFields(logsResp.Body)
		return output.PrintLogFields(out, fields)
	},
}

var dbBackupsCmd = &cobra.Command{
	Use:   "backups <database_name>",
	Short: "List backups for a database",
	Long:  `List backups for a database, sorted by most recent first.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		backups, err := deployClient.ListDatabaseBackups(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
		)
		if err != nil {
			return err
		}

		return output.PrintDatabaseBackups(out, backups)
	},
}

var dbBackupCmd = &cobra.Command{
	Use:   "backup <database_name>",
	Short: "Trigger an on-demand backup",
	Long: `Trigger an on-demand backup for a database. The database must have backups
enabled.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Triggering database backup...")

		backup, err := deployClient.TriggerDatabaseBackup(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
		)
		if err != nil {
			return err
		}

		return output.PrintDatabaseBackup(out, backup)
	},
}

var dbRestoreCmd = &cobra.Command{
	Use:   "restore <database_name>",
	Short: "Restore a new database from a backup",
	Long: `Create a new database by restoring from an existing database's backup. The
source database must have backups enabled.

Optionally specify --target-time for point-in-time recovery (RFC3339 format).
If omitted, the latest backup is restored.

Examples:
  iai databases restore my-restored-db --source-database my-db --instances 2 --cpu 1 --memory 2G --storage-size 20G
  iai databases restore my-restored-db --source-database my-db --target-time 2026-05-12T10:00:00Z --instances 2 --cpu 1 --memory 2G --storage-size 20G`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		databaseName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), dbOrganization, dbProject)
		if err != nil {
			return err
		}

		reqBody, err := inputs.BuildRestoreRequestBody(inputs.RestoreInput{
			DatabaseInput: inputs.DatabaseInput{
				Instances:       dbInstances,
				PostgresVersion: dbPostgresVersion,
				CPU:             dbCPU,
				Memory:          dbMemory,
				StorageSize:     dbStorageSize,
				Extensions:      dbExtensions,
				BackupSchedule:  dbBackupSchedule,
				BackupRetention: dbBackupRetention,
				StackId:         dbStackId,
			},
			SourceDatabase: dbSourceDatabase,
			TargetTime:     dbTargetTime,
		})
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting database restore request...")

		serverMessage, err := deployClient.RestoreDatabase(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			databaseName,
			reqBody,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var (
	dbPFPort      int
	dbPFLocalPort int
)

var dbPortForwardCmd = &cobra.Command{
	Use:   "port-forward <database_name>",
	Short: "Forward a local port to a database",
	Long: `Open a local TCP listener and tunnel traffic through the deployment operator
to a PostgreSQL database running in the cluster.

The remote port defaults to 5432. Use --port to override. Use --local-port
to choose the local listening port (defaults to the remote port).

After connecting you can use psql, pgAdmin, or any PostgreSQL client against
localhost:<local-port>.

Examples:
  iai databases port-forward my-db
  iai databases port-forward my-db --local-port 15432`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		databaseName := strings.TrimSpace(args[0])
		remotePort := dbPFPort
		if remotePort == 0 {
			remotePort = 5432
		}
		localPort := dbPFLocalPort
		if localPort == 0 {
			localPort = remotePort
		}
		return runPortForward(cmd.Context(), portForwardOpts{
			resourceType: "databases",
			resourceName: databaseName,
			remotePort:   remotePort,
			localPort:    localPort,
			org:          dbOrganization,
			project:      dbProject,
		})
	},
}

func addDatabaseResourceFlags(cmd *cobra.Command) {
	cmd.Flags().
		IntVar(&dbInstances, "instances", 0, "Number of PostgreSQL instances (minimum 1); values above 1 enable high availability")
	cmd.Flags().
		StringVar(&dbPostgresVersion, "postgres-version", "", "PostgreSQL major or major.minor version (e.g. 17, 16.4); supported range 15–17; defaults to latest if omitted")
	cmd.Flags().
		StringVar(&dbCPU, "cpu", "", "CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m); max 7 vCPU (7000m)")
	cmd.Flags().
		StringVar(&dbMemory, "memory", "", "Memory in megabytes (M) or gigabytes (G) (e.g. 512M, 1G, 2G); max 15G")
	cmd.Flags().
		StringVar(&dbStorageSize, "storage-size", "", "Storage size with G unit (e.g. 20G, 100G); must be between 10G and 200G; cannot be decreased")
	cmd.Flags().
		StringArrayVar(&dbExtensions, "extensions", nil, "PostgreSQL extension to install (can be repeated); replaces the default list, so include \"vector\" explicitly if needed; defaults to [vector] if omitted")
	cmd.Flags().
		StringVar(&dbBackupSchedule, "backup-schedule", "", "Backup schedule as a 6-field cron expression (second minute hour day month weekday, e.g. \"0 0 2 * * *\" for daily at 02:00)")
	cmd.Flags().
		StringVar(&dbBackupRetention, "backup-retention", "", "How long to retain backups (e.g. 30d, 4w, 6m)")
}

func init() {
	// databases list
	dbListCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbListCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")

	// databases describe
	dbDescribeCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbDescribeCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")

	// databases create
	dbCreateCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbCreateCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")
	addDatabaseResourceFlags(dbCreateCmd)
	dbCreateCmd.Flags().
		StringVar(&dbStackId, "stack-id", "", "Stack ID to assign the database to")
	_ = dbCreateCmd.MarkFlagRequired("instances")
	_ = dbCreateCmd.MarkFlagRequired("cpu")
	_ = dbCreateCmd.MarkFlagRequired("memory")
	_ = dbCreateCmd.MarkFlagRequired("storage-size")

	// databases update
	dbUpdateCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbUpdateCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")
	addDatabaseResourceFlags(dbUpdateCmd)
	dbUpdateCmd.Flags().
		BoolVar(&dbClearBackup, "clear-backup", false, "Remove backup configuration from the database")
	dbUpdateCmd.Flags().
		StringVar(&dbStackId, "stack-id", "", "Stack ID to assign the database to")
	dbUpdateCmd.Flags().
		BoolVar(&dbClearStackId, "clear-stack-id", false, "Remove the database from its stack")

	// databases delete
	dbDeleteCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbDeleteCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")

	// databases logs
	dbLogsCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbLogsCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")
	dbLogsCmd.Flags().
		BoolVarP(&dbLogsFollow, "follow", "f", false, "Stream new log entries as they arrive; mutually exclusive with --end-time")
	dbLogsCmd.Flags().
		StringVar(&dbLogsSince, "since", "", "Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time")
	dbLogsCmd.Flags().
		StringVar(&dbLogsStartTime, "start-time", "", "Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window")
	dbLogsCmd.Flags().
		StringVar(&dbLogsEndTime, "end-time", "", "Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow")
	dbLogsCmd.Flags().
		BoolVar(&dbLogsRaw, "raw", false, "Output raw server JSON without formatting")
	dbLogsCmd.Flags().
		StringSliceVar(&dbLogsFields, "fields", nil, "Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for raw output")
	dbLogsCmd.Flags().
		BoolVar(&dbLogsAllFields, "all-fields", false, "Show all extra fields from structured (JSON) logs after the message")
	dbLogsCmd.MarkFlagsMutuallyExclusive("raw", "fields")
	dbLogsCmd.MarkFlagsMutuallyExclusive("raw", "all-fields")
	dbLogsCmd.MarkFlagsMutuallyExclusive("fields", "all-fields")

	// databases backups
	dbBackupsCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbBackupsCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")

	// databases backup
	dbBackupCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbBackupCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")

	// databases restore
	dbRestoreCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbRestoreCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")
	addDatabaseResourceFlags(dbRestoreCmd)
	_ = dbRestoreCmd.MarkFlagRequired("instances")
	_ = dbRestoreCmd.MarkFlagRequired("cpu")
	_ = dbRestoreCmd.MarkFlagRequired("memory")
	_ = dbRestoreCmd.MarkFlagRequired("storage-size")
	dbRestoreCmd.Flags().
		StringVar(&dbSourceDatabase, "source-database", "", "Name of the database to restore from; must have backups enabled")
	dbRestoreCmd.Flags().
		StringVar(&dbTargetTime, "target-time", "", "RFC3339 timestamp for point-in-time recovery (e.g. 2026-05-12T10:00:00Z); omit to restore the latest backup")
	dbRestoreCmd.Flags().
		StringVar(&dbStackId, "stack-id", "", "Stack ID to assign the restored database to")
	_ = dbRestoreCmd.MarkFlagRequired("source-database")

	// Flags for "databases port-forward"
	dbPortForwardCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbPortForwardCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")
	dbPortForwardCmd.Flags().
		IntVar(&dbPFPort, "port", 0, "Remote port on the database (defaults to 5432)")
	dbPortForwardCmd.Flags().
		IntVar(&dbPFLocalPort, "local-port", 0, "Local port to listen on (defaults to the remote port)")

	// Flags for "databases log-fields"
	dbLogFieldsCmd.Flags().
		StringVarP(&dbProject, "project", "p", "", "Project name")
	dbLogFieldsCmd.Flags().
		StringVarP(&dbOrganization, "organization", "o", "", "Organization name")
	dbLogFieldsCmd.Flags().
		StringVar(&dbLogFieldsSince, "since", "1h", "Relative duration to scan (e.g. 5m, 1h)")

	// Wire up command hierarchy
	databasesCmd.AddCommand(
		dbListCmd, dbDescribeCmd, dbCreateCmd, dbUpdateCmd, dbDeleteCmd,
		dbLogsCmd, dbLogFieldsCmd, dbBackupsCmd, dbBackupCmd, dbRestoreCmd, dbPortForwardCmd,
	)
	rootCmd.AddCommand(databasesCmd)
}
