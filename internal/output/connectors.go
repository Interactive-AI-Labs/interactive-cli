package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintMcpConnectionList(out io.Writer, conns []clients.McpConnection) error {
	if len(conns) == 0 {
		fmt.Fprintln(out, "No connectors found.")
		return nil
	}
	headers := []string{"ID", "NAME", "TYPE", "STATUS", "TOOLS", "ENDPOINT", "UPDATED"}
	rows := make([][]string, len(conns))
	for i, c := range conns {
		rows[i] = []string{
			c.ID,
			c.Name,
			c.Type,
			c.Status,
			fmt.Sprintf("%d", c.ToolCount),
			c.EndpointURL,
			LocalTime(c.UpdatedAt),
		}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpConnectionDetail(out io.Writer, conn *clients.McpConnectionDetail) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", conn.ID)
	fmt.Fprintf(w, "Name:\t%s\n", conn.Name)
	fmt.Fprintf(w, "Type:\t%s\n", conn.Type)
	fmt.Fprintf(w, "Status:\t%s\n", conn.Status)
	if conn.Slug != "" {
		fmt.Fprintf(w, "Slug:\t%s\n", conn.Slug)
	}
	if conn.Description != "" {
		fmt.Fprintf(w, "Description:\t%s\n", conn.Description)
	}
	fmt.Fprintf(w, "Endpoint:\t%s\n", conn.EndpointURL)
	fmt.Fprintf(w, "Transport:\t%s\n", conn.Transport)
	fmt.Fprintf(w, "Auth Type:\t%s\n", conn.AuthType)
	// Show whether a credential is stored for connectors that need auth — the
	// credential itself is never returned, so this is the only confirmation.
	if conn.AuthType != "" && conn.AuthType != "none" {
		fmt.Fprintf(w, "Credential Set:\t%t\n", conn.HasCredential)
	}
	if conn.CatalogID != "" {
		fmt.Fprintf(w, "Catalog ID:\t%s\n", conn.CatalogID)
	}
	if conn.ProtocolVersion != "" {
		fmt.Fprintf(w, "Protocol:\t%s\n", conn.ProtocolVersion)
	}
	if conn.LastVerifiedAt != "" {
		fmt.Fprintf(w, "Last Verified:\t%s\n", LocalTime(conn.LastVerifiedAt))
	}
	if conn.LastError != "" {
		if conn.LastErrorClass != "" {
			fmt.Fprintf(w, "Last Error:\t%s (%s)\n", conn.LastError, conn.LastErrorClass)
		} else {
			fmt.Fprintf(w, "Last Error:\t%s\n", conn.LastError)
		}
	}
	if len(conn.ConnectedAgents) > 0 {
		names := make([]string, len(conn.ConnectedAgents))
		for i, a := range conn.ConnectedAgents {
			names[i] = a.Name
		}
		fmt.Fprintf(w, "Connected Agents:\t%s\n", strings.Join(names, ", "))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if len(conn.Tools) == 0 {
		fmt.Fprintln(out, "\nNo tools discovered yet. Run 'iai connectors verify' to refresh.")
		return nil
	}
	return printMcpToolsTable(out, conn.Tools)
}

// InputSchema is omitted — too verbose for terminal display.
func printMcpToolsTable(out io.Writer, tools []clients.McpTool) error {
	fmt.Fprintln(out, "\nTools:")
	headers := []string{"NAME", "ENABLED", "DESCRIPTION"}
	rows := make([][]string, len(tools))
	for i, tl := range tools {
		rows[i] = []string{tl.Name, fmt.Sprintf("%t", tl.Enabled), tl.Description}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpCatalog(out io.Writer, entries []clients.McpCatalogEntry) error {
	if len(entries) == 0 {
		fmt.Fprintln(out, "No catalog entries found.")
		return nil
	}
	headers := []string{"ID", "NAME", "CATEGORY", "TYPE", "AUTH"}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{e.ID, e.Name, e.Category, e.Type, TruncateList(e.AuthMethods, 3)}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpVerifyResult(out io.Writer, res *clients.McpVerifyData) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Status:\t%s\n", res.Status)
	if res.ProtocolVersion != "" {
		fmt.Fprintf(w, "Protocol:\t%s\n", res.ProtocolVersion)
	}
	if res.ErrorClass != "" {
		fmt.Fprintf(w, "Error Class:\t%s\n", res.ErrorClass)
	}
	if res.ErrorMessage != "" {
		fmt.Fprintf(w, "Error:\t%s\n", res.ErrorMessage)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if len(res.Tools) == 0 {
		return nil
	}
	return printMcpToolsTable(out, res.Tools)
}

func PrintMcpToolResult(out io.Writer, res *clients.McpToolCallData) error {
	fmt.Fprintf(out, "Status: %s\n", res.Status)
	if res.ErrorClass != "" {
		fmt.Fprintf(out, "Error Class: %s\n", res.ErrorClass)
	}
	if res.ErrorMessage != "" {
		fmt.Fprintf(out, "Error: %s\n", res.ErrorMessage)
	}
	if len(res.Result) > 0 {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, res.Result, "", "  "); err != nil {
			return err
		}
		fmt.Fprintf(out, "\nResult:\n%s\n", pretty.String())
	}
	return nil
}
