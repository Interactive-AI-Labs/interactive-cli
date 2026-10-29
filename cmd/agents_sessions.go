package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	agentSessionCustomerID    string
	agentSessionTitle         string
	agentSessionLabels        []string
	agentSessionAllowGreeting bool
	agentSessionCreateJSON    bool
	agentSessionCreateYAML    bool

	agentSessionListCustomerID string
	agentSessionListLabels     []string
	agentSessionListLimit      int
	agentSessionListCursor     string
	agentSessionListSort       string
	agentSessionListJSON       bool
	agentSessionListYAML       bool
)

var agentSessionsCmd = &cobra.Command{
	Use:     "sessions",
	Aliases: []string{"session"},
	Short:   "Conversation sessions on an agent",
	Long: `Manage conversation sessions on a deployed agent.

A session holds one conversation. Create one, send it messages with
'iai agents events create' and read the agent's replies with
'iai agents events list', or talk to it from the terminal with 'iai agents chat'.`,
}

var agentSessionCreateCmd = &cobra.Command{
	Use:   "create <agent_name>",
	Short: "Create a conversation session on an agent",
	Long:  `Create a conversation session on a deployed agent and print its ID.`,
	Example: `  iai agents sessions create my-agent
  iai agents sessions create my-agent --customer-id cust_42 --title "Order 123"
  iai agents sessions create my-agent --labels test --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}
		session, err := deployClient.CreateAgentSession(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agentName,
			deployment.AgentSessionCreateRequest{
				CustomerID: agentSessionCustomerID,
				Title:      agentSessionTitle,
				Labels:     agentSessionLabels,
			},
			agentSessionAllowGreeting,
		)
		if err != nil {
			return err
		}
		if agentSessionCreateJSON {
			return output.PrintRawJSON(out, session.Raw)
		}
		if agentSessionCreateYAML {
			return output.PrintRawYAML(out, session.Raw)
		}
		return output.PrintAgentSession(out, session)
	},
}

var agentSessionListCmd = &cobra.Command{
	Use:     "list <agent_name>",
	Aliases: []string{"ls"},
	Short:   "List conversation sessions on an agent",
	Long:    `List conversation sessions on a deployed agent.`,
	Example: `  iai agents sessions list my-agent
  iai agents sessions list my-agent --customer-id cust_42
  iai agents sessions list my-agent --labels test --limit 10
  iai agents sessions list my-agent --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}
		list, err := deployClient.ListAgentSessions(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agentName,
			deployment.AgentSessionListOptions{
				CustomerID: agentSessionListCustomerID,
				Labels:     agentSessionListLabels,
				Limit:      agentSessionListLimit,
				Cursor:     agentSessionListCursor,
				Sort:       agentSessionListSort,
			},
		)
		if err != nil {
			return err
		}
		if agentSessionListJSON {
			return output.PrintRawJSON(out, list.Raw)
		}
		if agentSessionListYAML {
			return output.PrintRawYAML(out, list.Raw)
		}
		return output.PrintAgentSessionList(out, list)
	},
}

func init() {
	f := agentSessionCreateCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.StringVar(&agentSessionCustomerID, "customer-id", "",
		"Customer the conversation belongs to; a guest customer is created when omitted")
	f.StringVar(&agentSessionTitle, "title", "", "Conversation title")
	f.StringArrayVar(&agentSessionLabels, "labels", nil,
		"Label attached to the conversation; repeat the flag once per label")
	f.BoolVar(&agentSessionAllowGreeting, "allow-greeting", false,
		"Let the agent open the conversation with a greeting")
	f.BoolVar(&agentSessionCreateJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&agentSessionCreateYAML, "yaml", false, "Output raw API response as YAML")
	agentSessionCreateCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	f = agentSessionListCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.StringVar(&agentSessionListCustomerID, "customer-id", "", "Only sessions of this customer")
	f.StringArrayVar(&agentSessionListLabels, "labels", nil,
		"Only sessions carrying this label; repeat the flag once per label")
	f.IntVar(&agentSessionListLimit, "limit", 0, "Sessions per page (1-100)")
	indexFlag(agentSessionListCmd, "limit", "1-100")
	f.StringVar(&agentSessionListCursor, "cursor", "", "Cursor from the previous page")
	f.StringVar(&agentSessionListSort, "sort", "", "Creation order: asc or desc (default asc)")
	f.BoolVar(&agentSessionListJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&agentSessionListYAML, "yaml", false, "Output raw API response as YAML")
	agentSessionListCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	agentSessionsCmd.AddCommand(agentSessionCreateCmd, agentSessionListCmd)
	agentsCmd.AddCommand(agentSessionsCmd)
}
