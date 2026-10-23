package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSkillCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Print the canonical SKILL.md for repoctx",
		Long: "Emits the canonical SKILL.md describing how to use repoctx from a coding\n" +
			"agent. The same content is published to every agent channel, so there is\n" +
			"one source of truth. Output is deterministic and written to stdout.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch format {
			case "", "agents", "claude", "cursor", "opencode":
			default:
				return fmt.Errorf("unknown --format %q (claude, cursor, agents, opencode)", format)
			}
			fmt.Fprint(cmd.OutOrStdout(), skillDocument())
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "agents",
		"target agent hint (claude, cursor, agents, opencode); content is identical")
	return cmd
}

func skillDocument() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: repoctx\n")
	b.WriteString("description: Use when AGENTS.md, CLAUDE.md, .cursorrules or any agent context file must be checked against the real repository — regenerates the code-derived sections (commands, modules, dependencies) and detects rot (ghost commands, stale paths). Triggers on \"update AGENTS.md\", \"context file is out of date\", \"audit agent context\", \"repoctx\".\n")
	b.WriteString("license: MIT\n")
	b.WriteString("compatibility: Requires the repoctx CLI (npx @sriruma/repoctx / brew install repoctx). Works in any repo with a manifest.\n")
	b.WriteString("---\n\n")
	b.WriteString("# repoctx\n\n")
	b.WriteString("repoctx keeps AI coding-agent context files (AGENTS.md, CLAUDE.md, .cursorrules) truthful by regenerating the code-derived sections from the real repository and detecting rot.\n\n")
	b.WriteString("## When to use\n\n")
	b.WriteString("- The user asks to update or refresh AGENTS.md / CLAUDE.md / .cursorrules.\n")
	b.WriteString("- A context file looks out of date (commands, paths, modules).\n")
	b.WriteString("- Before finishing a task that changed commands, scripts, dependencies, manifests or package layout.\n")
	b.WriteString("- In CI, to gate merges on context drift (use `--since` so only drift the change introduced fails).\n\n")
	b.WriteString("## When NOT to use\n\n")
	b.WriteString("- Editing human-written sections (overview, conventions, TODOs) — those are owned by people.\n")
	b.WriteString("- Repositories with no manifest (repoctx has nothing to extract).\n\n")
	b.WriteString("## Core workflow\n\n")
	b.WriteString("1. `repoctx info [dir]` — see what was detected (manifests, commands, modules).\n")
	b.WriteString("2. `repoctx init [dir]` — bootstrap a context file where none exists.\n")
	b.WriteString("3. `repoctx generate [dir]` — regenerate the code-derived tables (preserves human content).\n")
	b.WriteString("4. `repoctx audit [dir] --check` — fail if the context drifted.\n\n")
	b.WriteString("## Rules\n\n")
	b.WriteString("- Never hand-edit the tables between `<!-- repoctx:start -->` and `<!-- repoctx:end -->`.\n")
	b.WriteString("- Run `repoctx generate .` after changing commands, scripts, dependencies or manifests.\n")
	b.WriteString("- Run `repoctx audit . --check` before finishing; fix anything it reports.\n")
	b.WriteString("- In CI, add `--since <rev>` to `audit --check` so only findings on lines changed since `<rev>` fail; inherited rot is reported but does not block the merge.\n")
	b.WriteString("- Prefer `--file` to target a specific context file; otherwise repoctx discovers all AGENTS.md/CLAUDE.md in the tree.\n\n")
	b.WriteString("## Output format\n\n")
	b.WriteString("When you report an audit result, use this shape:\n\n")
	b.WriteString("```markdown\n")
	b.WriteString("## Context health\n\n")
	b.WriteString("**File:** AGENTS.md · **Score:** <score>/100\n\n")
	b.WriteString("| Check | Result |\n")
	b.WriteString("|---|---|\n")
	b.WriteString("| commands | 11 claimed, all found |\n")
	b.WriteString("| paths | 9 referenced, 1 stale |\n\n")
	b.WriteString("### What to fix\n")
	b.WriteString("- `docs/api.md` — referenced in the context file, but the file moved.\n\n")
	b.WriteString("### How to fix\n")
	b.WriteString("- update the path, or run `repoctx generate .` if the command list changed\n")
	b.WriteString("```\n\n")
	b.WriteString("- Quote the exact path or command the audit reported. Never paraphrase it.\n")
	b.WriteString("- If every check passes, say so in one line and stop. Do not invent findings.\n\n")
	b.WriteString("## Troubleshooting\n\n")
	b.WriteString("- `no context files found`: the repo has no AGENTS.md/CLAUDE.md; run `repoctx init`.\n")
	b.WriteString("- `audit` reports stale paths: the referenced file moved or was renamed; update or remove the reference.\n")
	b.WriteString("- `audit` reports ghost commands: a claimed command no longer exists in any manifest; remove it.\n")
	b.WriteString("- `audit --since` says `pre-existing`: the rot predates the base revision; it does not fail the run. Fix it or leave it as accepted debt.\n")
	b.WriteString("- `audit --since` errors: the root is not a git working tree, or the revision cannot be resolved.\n")
	b.WriteString("- Markers missing: `generate` refuses to overwrite a file without markers; add `<!-- repoctx:start -->` / `<!-- repoctx:end -->` or use `init`.\n\n")
	b.WriteString(workflowBlock([]string{"AGENTS.md"}) + "\n")
	return b.String()
}
