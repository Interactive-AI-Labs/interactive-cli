package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/conversation"
	"github.com/spf13/cobra"
)

var (
	chatSessionID  string
	chatCustomerID string
	chatTitle      string
)

var agentChatCmd = &cobra.Command{
	Use:   "chat <agent_name>",
	Short: "Talk to an agent from the terminal",
	Long: `Talk to a deployed agent from the terminal.

Each line you enter is sent as a customer message and the agent's reply is shown
as it is written. Ctrl+C or end of input stops the chat; the session stays on
the agent.`,
	Example: `  iai agents chat my-agent
  iai agents chat my-agent --session-id 9mtKfiDIaZ
  iai agents chat my-agent --customer-id cust_42 --title "Order 123"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName := strings.TrimSpace(args[0])
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		pCtx, _, deployClient, err := resolveProject(
			ctx, agentOrganization, agentProject, resolveOpts{deployTimeout: 0},
		)
		if err != nil {
			return err
		}
		return conversation.Chat(ctx, conversation.ChatDeps{
			API:    deployClient,
			Stdin:  cmd.InOrStdin(),
			Stdout: cmd.OutOrStdout(),
			Stderr: cmd.ErrOrStderr(),
		}, conversation.ChatOptions{
			OrgID:      pCtx.orgId,
			ProjectID:  pCtx.projectId,
			AgentName:  agentName,
			SessionID:  chatSessionID,
			CustomerID: chatCustomerID,
			Title:      chatTitle,
		})
	},
}

func init() {
	f := agentChatCmd.Flags()
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	f.StringVar(
		&chatSessionID,
		"session-id",
		"",
		"Resume this conversation session, showing its latest 50 messages first; a new one is created when omitted",
	)
	f.StringVar(&chatCustomerID, "customer-id", "",
		"Customer the new conversation belongs to; a guest customer is created when omitted")
	f.StringVar(&chatTitle, "title", "", "Title of the new conversation")
	agentChatCmd.MarkFlagsMutuallyExclusive("session-id", "customer-id")
	agentChatCmd.MarkFlagsMutuallyExclusive("session-id", "title")

	agentsCmd.AddCommand(agentChatCmd)
}
