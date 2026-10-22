# Rot patterns

What `repoctx audit` catches, what the report looks like, and how to fix each
case. Every example here is a real output of the audit.

## 1. Ghost command

**What it is.** The context file claims a command that no longer exists in any
manifest. The script was renamed, the `scripts` key was dropped from
`package.json`, the Maven plugin was removed, the Makefile target was
renamed.

**What the audit says.**

```
  !!  commands: 1 ghost command(s) out of 2 claimed
    - npm run deploy: ghost command: no longer present in the project
```

**How to fix.** Three options, in order of preference:

- The command still exists under a different name — update the table.
- The command was deleted on purpose — remove the row.
- The command lives in a manifest the adapter does not read (a custom
  `Makefile` include, a `justfile`, a `Taskfile.yml`) — repoctx cannot see it.
  Move it to a manifest it does read, or keep the row and accept the report.

## 2. Stale path

**What it is.** The context file references a file that is not in the
repository. The package moved, the doc was renamed, the directory was
reorganized and the context file was not updated.

**What the audit says.**

```
  !!  paths: 2 stale path(s)
    - docs/api.md: stale path: not found in repository
    - internal/cli/audit.go: stale path: not found in repository
```

**How to fix.** Update the reference to where the file is now, or delete the
reference if the file is gone. A path inside a fenced code block is not
checked — only prose references are, so a command example does not produce
this issue.

## 3. Missing markers

**What it is.** `generate` refuses to write a context file that does not have
the `<!-- repoctx:start -->` / `<!-- repoctx:end -->` markers. Without them
there is no way to tell the code-derived tables from the human-written prose,
so the safe thing is to do nothing.

**What the audit says.** The commands check is skipped and the report says so:

```
  ok  commands: no repoctx markers; commands check skipped
```

**How to fix.** Put the tables between the markers:

```markdown
<!-- repoctx:start -->
## Commands

| Command | Source |
|---|---|
| `make test` | `Makefile` |
<!-- repoctx:end -->
```

`repoctx init .` creates a file that already has them.

## 4. Nested context files

**What it is.** A monorepo with a context file per package. Generating the
root file with the wrong scope pulls in every package's commands, and the
root context file ends up claiming commands it does not own.

**What the audit says.** Nothing directly — the audit is per file, and a wrong
scope shows up as ghost commands in the root file, because the root table
claims commands that belong to a package.

**How to fix.** Generate per scope and keep the root file to the commands that
run from the root:

```sh
repoctx generate . --file AGENTS.md
repoctx generate ./services/api --file AGENTS.md
```

`repoctx info .` shows which manifests were found and at what depth.

## 5. Manifest not detected

**What it is.** The repository has no manifest repoctx reads, or the manifest
is deeper than the scan goes. Nothing is extracted, so the generated tables
are empty.

**What the audit says.**

```
  ok  commands: 0 commands claimed
  ok  paths: all referenced paths exist
```

An empty table is the tell. `repoctx info .` lists the manifests it found and
the commands it extracted from each one; a manifest missing from that list is
a manifest repoctx did not read.

**How to fix.** Move the manifest to the repository root, or point repoctx at
the right directory. The supported manifests are `package.json`, `go.mod`,
`pom.xml`, `build.gradle`, `Cargo.toml`, `pyproject.toml`, `Makefile`,
`CMakeLists.txt`, `Gemfile` and `composer.json`.

## 6. Drifted right after generating

**What it is.** The audit fails immediately after a successful `generate`. The
file changed between the two commands — a formatter, an editor with autosave,
or a hook that rewrites the tables.

**How to fix.** Generate, then audit, then commit both in the same commit:

```sh
repoctx generate . && repoctx audit . --check
```

If something in your toolchain rewrites the file after every write, find it
before trusting the gate.

## The report

`repoctx audit . --check` prints one line per check and exits 1 on any
failure:

```
FAIL  /path/to/repo/AGENTS.md  score 50/100
  ok  commands: 1 commands claimed
  !!  paths: 1 stale path(s)
    - docs/api.md: stale path: not found in repository
```

The score is 100 when every check passes. In CI, `--check` turns the failure
into a non-zero exit code, which is the whole point: a context file that
lies should fail the build, not the next agent that reads it.
