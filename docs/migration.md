# Migration guide: 0.x → 1.0.0 (upcoming)

repoctx is on 0.x. The machine contract documented in
[docs/contract.md](contract.md) is **not frozen** yet: breaking changes are
allowed without a major bump while the tool is validated in real use. `1.0.0`
will be released when that contract is frozen. This guide covers what moving
between 0.x releases looks like, so the path to 1.0.0 stays smooth.

## Upgrade in three steps

1. **Install the new binary.**

   ```sh
   go install github.com/SrIruma/repoctx/cmd/repoctx@latest
   ```

2. **Regenerate your context files.** `generate` rewrites only the content
   between the repoctx markers, so your human-written prose is preserved:

   ```sh
   repoctx generate .
   ```

3. **Verify with the CI gate.** `audit --check` exits non-zero when a context
   file has drifted, so it can gate merges:

   ```sh
   repoctx audit . --check
   ```

## What changed on the way to 1.0.0

Everything below was merged during the `0.x` series. It is what a `1.0.0`
release will solidify once the contract is frozen.

### Configuration

- `repoctx.toml` in the scanned directory configures `max_depth`, `skip_dirs`
  and `files`; an explicit path can be given with `--config`. Values merge in
  `flags > config > defaults` precedence order.

### New flags

- `info`, `generate` and `audit` share `--max-depth` (scan depth limit) and
  `--skip-dirs` (repeatable, adds to the built-in skip list).
- `generate` gained `--dry-run` (preview without writing) and `--file`
  (repeatable; overrides the `files` config).
- `audit` gained `--check` (CI gating, exit `1` on rot) and `--json`.

### `info --json` contract changes

- A corrupt manifest is now a **soft failure**: the scan keeps the manifest
  with zero facts and an `errors` field, and the exit code stays `0`. Older
  releases could drop the manifest or fail the scan.
- `deps` is omitted when a manifest has no dependencies.
- `errors` is present only when extraction failed.

### `audit --json` contract changes

- `audit --json` emits an array of per-file reports
  (`file`, `checks`, `score`, `passed`); with `--check` the exit code reflects
  the result.
- Rot is reported per check with `issues` entries (`command`/`path` + `detail`).

### New ecosystems

- All 11 detected manifests have adapters, so nothing is reported as
  "adapter pending" anymore. Beyond the original five (`package.json`,
  `Cargo.toml`, `go.mod`, `pyproject.toml`, `Makefile`), 1.0.0 covers
  `CMakeLists.txt`, `Gemfile`, `composer.json`, `pom.xml`, `build.gradle`,
  and `meson.build`.

### Workspaces

- npm/pnpm and cargo workspaces are scanned naturally: every manifest is
  detected, commands are attributed per manifest, and identical commands from
  different packages are disambiguated by the Source column. A cargo virtual
  workspace root reports the workspace-wide commands but not `cargo run`.

## Checklist for JSON consumers

If your scripts or CI parse `repoctx --json` output, verify these before
building on the contract:

- [ ] Handle a non-zero exit from `audit --check` as "context is rotting".
- [ ] Treat a manifest with an `errors` field as unparsed (its `commands` is
      `null`), not as a scan failure.
- [ ] Use the `--file`, `--config`, `--max-depth` and `--skip-dirs` flags only
      with the documented contracts in mind; today their behaviour can still
      change (0.x), and it will be frozen at `1.0.0`.
