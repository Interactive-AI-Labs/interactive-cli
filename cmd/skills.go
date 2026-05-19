package cmd

import "github.com/spf13/cobra"

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "skill",
		Plural:   "skills",
		Aliases:  []string{"skill"},
		GroupID:  groupCopilot,
		Short:    "Manage Copilot (interactive-copilot) skills — NOT interactive-agent behaviors",
		Long: `Manage Copilot skills for the interactive-copilot service.

IMPORTANT: These are Copilot skills — they are NOT interactive-agent behaviors.
Skills here are loaded by the Copilot runtime and injected into Copilot
conversations as context. They have no effect on the conversational agent
(interactive-agent).

Each Copilot skill is a free-form markdown bundle. It carries a short description
and an "intents" list of natural-language triggers (stored in config.skill) that
the Copilot uses to route incoming queries to the right skill at runtime.`,
		RouteSegment:          "skills",
		BindPromptConfigFlags: bindSkillConfigFlags,
		CreateLong: `Create a new Copilot skill for the interactive-copilot service.

The skill body is provided as markdown via --file. Optional --description and
--intents populate the config.skill block consumed by the Copilot runtime to
assemble its intent → skill routing table.

Pass --intents once per intent; the flag is repeatable so individual
intents may contain commas (e.g. "summarize, then explain").

Example (skill.md):
  # Summarize Trace

  Given a Langfuse trace ID, fetch the trace and summarize key steps,
  latencies, and any errors.

The server automatically assigns the "latest" label to new versions. To make
a version retrievable via the default 'get' (which resolves "production"),
assign the "production" label with --labels production.

Examples:
  iai skills create summarize-trace --file ./skill.md \
    --description "Summarize a Langfuse trace" \
    --intents "summarize trace" --intents "explain trace"
  iai skills create summarize-trace --file ./skill.md --labels production`,
		ListLong: `List Copilot skills in a project.

Returns all Copilot skills with their name, labels, tags, and last update time.
Folders are shown with a trailing "/" (colored when stdout is a terminal) and
can be browsed into with --folder.

Examples:
  iai skills list
  iai skills list --folder my-folder
  iai skills list --page 2 --limit 10`,
		GetLong: `Show a Copilot skill in detail, including its config and full content.

By default returns the version labeled "production". Use --version to retrieve a
specific version number, or --label to resolve a different label.

Examples:
  iai skills describe summarize-trace
  iai skills describe summarize-trace --version 3
  iai skills describe summarize-trace --label staging`,
		UpdateLong: `Update a Copilot skill by creating a new version with updated content.

Each update creates a brand-new version with exactly the content and config
provided on the command line — the previous version is preserved unchanged
but is not inherited from. In particular, if --description or --intents are
omitted the new version's config.skill block will be empty, even if the
prior version had values for them. Pass them again on every update if you
want the new version to keep them.

Pass --intents once per intent (the flag is repeatable).

Examples:
  iai skills update summarize-trace --file ./skill.md \
    --description "Summarize a Langfuse trace" \
    --intents "summarize trace" --intents "explain trace"
  iai skills update summarize-trace --file ./skill.md --labels production,staging`,
		DeleteLong: `Delete a Copilot skill and all its versions, or delete specific versions.

Without flags, deletes the skill and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions
with a specific label. Use -f to skip the confirmation prompt.

Examples:
  iai skills delete summarize-trace
  iai skills delete summarize-trace -f
  iai skills delete summarize-trace --version 3
  iai skills delete summarize-trace --label staging`,
	})
}

// bindSkillConfigFlags registers --description and --intents and builds the
// config.skill payload block.
func bindSkillConfigFlags(cmd *cobra.Command) ConfigFlagBuilder {
	var (
		description string
		intents     []string
	)
	cmd.Flags().StringVar(
		&description, "description", "",
		"Short description of the skill (stored in config.skill.description)",
	)
	cmd.Flags().StringArrayVar(
		&intents, "intents", nil,
		"Natural-language intent that triggers this skill — repeat the flag "+
			"once per intent (stored in config.skill.intents)",
	)

	return func() map[string]any {
		skill := map[string]any{}
		if description != "" {
			skill["description"] = description
		}
		if len(intents) > 0 {
			skill["intents"] = intents
		}
		if len(skill) == 0 {
			return nil
		}
		return map[string]any{"skill": skill}
	}
}
