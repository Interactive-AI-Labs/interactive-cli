package cmd

import "github.com/spf13/cobra"

func init() {
	registerPromptType(PromptTypeConfig{
		TypeName: "skill",
		Plural:   "skills",
		Aliases:  []string{"skill"},
		GroupID:  groupCopilot,
		Short:    "Manage Interactive Copilot skills (not to be confused with context items that configure the Interactive Agent)",
		Long: `Manage Interactive Copilot skills for the interactive-copilot service.

IMPORTANT: These are Interactive Copilot skills, not to be confused with
context items that configure the Interactive Agent. Skills are loaded by the
Copilot runtime and injected into Copilot conversations as context. They
have no effect on the Interactive Agent.

Each Copilot skill is a free-form markdown bundle. It carries a short description
and an "intents" list of natural-language triggers (stored in config.skill) that
the Copilot uses to route incoming queries to the right skill at runtime.`,
		RouteSegment:          "skills",
		GlobalScope:           true,
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

The server automatically assigns the "latest" label to new versions. Copilot
loads the version labeled "active", so assign it with --labels active to make
a skill the one Copilot uses.`,
		CreateExample: `  iai skills create summarize-trace --file ./skill.md \
    --description "Summarize a Langfuse trace" \
    --intents "summarize trace" --intents "explain trace"
  iai skills create summarize-trace --file ./skill.md --labels active
  iai skills create summarize-trace --file ./skill.md -m "initial trace summary skill"`,
		ListLong: `List Copilot skills in a project.

A skill has one of two scopes, shown in the SCOPE column. Project skills are
yours to edit. Global skills are owned by Interactive and served to every
project; they are read-only here. Both scopes are listed, so a name that exists
in both appears twice, once per scope — at runtime the Copilot prioritizes the
project skill over a global one of the same name.

The listing is complete rather than paginated. Folders are shown with a trailing
"/" (colored when stdout is a terminal) and can be browsed into with --folder,
which lists the project alone: global skills are project-wide and sit in no
folder.`,
		ListExample: `  iai skills list
  iai skills list --folder my-folder
  iai skills list --json`,
		GetLong: `Show a Copilot skill in detail, including its config and full content.

Reads the project's own skills. Pass --scope global for a skill Interactive
serves to every project; those are read-only here. A name that exists in both
scopes is two different skills: at runtime the Copilot prioritizes the project
one over the global one.

Without flags, returns the version the server resolves by default. Copilot
loads the "active" version, so use --label active to fetch the version Copilot
uses. Use --version to retrieve a specific version number, or --label to
resolve any other label. A global skill has only its "active" version.`,
		GetExample: `  iai skills get summarize-trace
  iai skills get routines --scope global
  iai skills get summarize-trace --version 3
  iai skills get summarize-trace --label active`,
		UpdateLong: `Update a Copilot skill by creating a new version with updated content.

Each update creates a brand-new version with exactly the content and config
provided on the command line — the previous version is preserved unchanged
but is not inherited from. In particular, if --description or --intents are
omitted the new version's config.skill block will be empty, even if the
prior version had values for them. Pass them again on every update if you
want the new version to keep them.

Pass --intents once per intent (the flag is repeatable).`,
		UpdateExample: `  iai skills update summarize-trace --file ./skill.md \
    --description "Summarize a Langfuse trace" \
    --intents "summarize trace" --intents "explain trace"
  iai skills update summarize-trace --file ./skill.md --labels active
  iai skills update summarize-trace --file ./skill.md -m "handle traces with no observations"`,
		DeleteLong: `Delete a Copilot skill and all its versions, or delete specific versions.

Without flags, deletes the skill and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions
with a specific label. Use -f to skip the confirmation prompt.`,
		DeleteExample: `  iai skills delete summarize-trace
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
