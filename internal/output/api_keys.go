package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var projectAPIKeyColumnMap = map[string]struct {
	Header string
	Value  func(clients.ProjectAPIKey) string
}{
	"id":         {"ID", func(k clients.ProjectAPIKey) string { return k.ID }},
	"public_key": {"PUBLIC KEY", func(k clients.ProjectAPIKey) string { return k.PublicKey }},
	"secret":     {"SECRET", func(k clients.ProjectAPIKey) string { return k.DisplaySecretKey }},
	"note":       {"NOTE", func(k clients.ProjectAPIKey) string { return stringValue(k.Note) }},
	"status": {
		"STATUS",
		func(k clients.ProjectAPIKey) string { return expiryStatus(k.ExpiresAt) },
	},
	"expires_at": {
		"EXPIRES",
		func(k clients.ProjectAPIKey) string { return formatExpiresAt(k.ExpiresAt) },
	},
	"last_used_at": {
		"LAST USED",
		func(k clients.ProjectAPIKey) string { return formatOptionalTime(k.LastUsedAt) },
	},
	"created_at": {
		"CREATED",
		func(k clients.ProjectAPIKey) string { return LocalTime(k.CreatedAt) },
	},
}

var routerAPIKeyColumnMap = map[string]struct {
	Header string
	Value  func(clients.RouterAPIKey) string
}{
	"id":   {"ID", func(k clients.RouterAPIKey) string { return k.ID }},
	"name": {"NAME", func(k clients.RouterAPIKey) string { return k.Name }},
	"description": {
		"DESCRIPTION",
		func(k clients.RouterAPIKey) string { return stringValue(k.Description) },
	},
	"status": {"STATUS", keyStatus},
	"key":    {"KEY", func(k clients.RouterAPIKey) string { return k.KeyPreview }},
	"disabled": {
		"DISABLED",
		func(k clients.RouterAPIKey) string { return fmt.Sprint(k.Disabled) },
	},
	"limit": {"LIMIT", func(k clients.RouterAPIKey) string { return formatUSD(k.Limit) }},
	"remaining": {
		"REMAINING",
		func(k clients.RouterAPIKey) string { return formatUSD(k.LimitRemaining) },
	},
	"limit_reset": {
		"RESET",
		func(k clients.RouterAPIKey) string { return stringValue(k.LimitReset) },
	},
	"expires_at": {
		"EXPIRES",
		func(k clients.RouterAPIKey) string { return formatExpiresAt(k.ExpiresAt) },
	},
	"last_used_at": {
		"LAST USED",
		func(k clients.RouterAPIKey) string { return formatOptionalTime(k.LastUsedAt) },
	},
	"created_at": {
		"CREATED",
		func(k clients.RouterAPIKey) string { return LocalTime(k.CreatedAt) },
	},
	"updated_at": {
		"UPDATED",
		func(k clients.RouterAPIKey) string { return formatOptionalTime(k.UpdatedAt) },
	},
	"project_id": {"PROJECT ID", func(k clients.RouterAPIKey) string { return k.ProjectID }},
	"user_id":    {"USER ID", func(k clients.RouterAPIKey) string { return k.UserID }},
}

func PrintProjectAPIKeyList(out io.Writer, keys []clients.ProjectAPIKey, columns []string) error {
	fmt.Fprintln(out)
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = projectAPIKeyColumnMap[col].Header
	}
	rows := make([][]string, 0, len(keys))
	for _, key := range keys {
		row := make([]string, len(columns))
		for i, col := range columns {
			row[i] = projectAPIKeyColumnMap[col].Value(key)
		}
		rows = append(rows, row)
	}
	return PrintTable(out, headers, rows)
}

func PrintRouterAPIKeyList(out io.Writer, keys []clients.RouterAPIKey, columns []string) error {
	fmt.Fprintln(out)
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = routerAPIKeyColumnMap[col].Header
	}
	rows := make([][]string, 0, len(keys))
	for _, key := range keys {
		row := make([]string, len(columns))
		for i, col := range columns {
			row[i] = routerAPIKeyColumnMap[col].Value(key)
		}
		rows = append(rows, row)
	}
	return PrintTable(out, headers, rows)
}

func keyStatus(key clients.RouterAPIKey) string {
	if key.Disabled {
		return "disabled"
	}
	return expiryStatus(key.ExpiresAt)
}

func expiryStatus(expiresAt *string) string {
	if expiresAt == nil || *expiresAt == "" {
		return "active"
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano} {
		parsed, err := time.Parse(layout, *expiresAt)
		if err == nil && time.Now().After(parsed) {
			return "expired"
		}
		if err == nil {
			return "active"
		}
	}
	return "active"
}

func formatExpiresAt(expiresAt *string) string {
	if expiresAt == nil || *expiresAt == "" {
		return "never"
	}
	return LocalTime(*expiresAt)
}

func formatOptionalTime(value *string) string {
	if value == nil || *value == "" {
		return "never"
	}
	return LocalTime(*value)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
