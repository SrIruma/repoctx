# repoctx

Keep `AGENTS.md`, `CLAUDE.md` and `.cursorrules` truthful against the real
repository. repoctx reads the manifests, extracts the commands and the module
structure as facts, regenerates the code-derived sections of your context file
between the markers, and audits the file for rot — commands that no longer
exist and paths that moved.

It is free and MIT. The engine is open source:
https://github.com/SrIruma/repoctx

## Requirements

- The repoctx CLI. `npx @sriruma/repoctx` runs it with no install; `brew install
  repoctx` and `winget install repoctx` install it permanently.
- A repository with a manifest: `package.json`, `go.mod`, `pom.xml`,
  `build.gradle`, `Cargo.toml`, `pyproject.toml`, `Makefile`, `CMakeLists.txt`,
  `Gemfile`, `composer.json`. repoctx reads the manifest to find the real
  commands — with no manifest there is nothing to extract.

## What the skill does

The agent loads `SKILL.md` and follows it. In practice:

1. `repoctx info .` — what was detected: manifests, commands, modules.
2. `repoctx init .` — bootstrap a context file where there is none.
3. `repoctx generate .` — regenerate the code-derived tables, leave your prose
   alone.
4. `repoctx audit . --check` — fail if the context drifted. Exit code 1 on
   drift, so it works as a CI gate.

## The one rule that matters

Never hand-edit the tables between `<!-- repoctx:start -->` and
`<!-- repoctx:end -->`. Run `repoctx generate .` instead: it rewrites only
those tables and leaves every human-written section — overview, conventions,
TODOs — exactly as it was.

`repoctx init --force` is the opposite: it rebuilds the file from scratch and
overwrites what you wrote. Use `generate`.

## Report format

When you report an audit result, use this shape:

```markdown
## Context health

**File:** AGENTS.md · **Score:** <score>/100

| Check | Result |
|---|---|
| commands | 11 claimed, all found |
| paths | 9 referenced, 1 stale |

### What to fix
- `docs/api.md` — referenced in the Structure section, but the file moved to
  `docs/reference/api.md`.

### How to fix
- update the path, or run `repoctx generate .` if the command list changed
```

- Quote the exact path or command the audit reported. Never paraphrase it.
- If every check passes, say so in one line and stop.

## Common rot

`references/patterns.md` lists what the audit catches and how to fix each
case: ghost commands, stale paths, missing markers, nested context files.

## Troubleshooting

- `repoctx: command not found` — install the CLI: `npx @sriruma/repoctx`, `brew install
  repoctx` or `winget install repoctx`.
- `generate` refuses to touch the file — the markers are missing. Add
  `<!-- repoctx:start -->` and `<!-- repoctx:end -->` around the tables, or
  run `repoctx init .` to create the file with them.
- The audit says the context drifted right after generating — the file changed
  between the two commands. Regenerate and commit them together.
- A repo is not detected — it has no manifest at the scanned depth, or the
  manifest is in a subdirectory. `repoctx info .` shows what was found.

## License

MIT. Use it anywhere, including commercially. See
https://github.com/SrIruma/repoctx for the source and the full documentation.
