# Validation

repoctx's own repository is dogfooded: `AGENTS.md` is regenerated in CI and
the build fails if it drifts. That gate proves *consistency* — the file always
matches what repoctx currently emits — but it cannot prove *truth*. The facts
repoctx emits are the only ground truth it is compared against, so a scanner
bug that silently drops a command would pass CI with the file still wrong, and
the audit can never fail because the file is regenerated right before it runs.

This directory (and the scripts behind it) is the counterweight: validation
that does not depend on repoctx's own output.

## Mechanisms

### 1. Live-rot fixture and test

`tests/fixtures/live/CLAUDE.md` is a hand-maintained context file with rot left
in on purpose — a ghost command inside the markers and a stale path in human
prose. `TestAuditFlagsLiveRotInCLAUDE` runs the full scan + audit pipeline
against it and asserts both issues are reported and the health score drops.
This is the scenario the repo's own dogfooding can never produce, and it runs
in every `go test ./...`:

```sh
go test ./internal/audit -run LiveRot
```

### 2. Semantic self-assertion test

`TestSelfAGENTSKeepsCriticalFacts` pins the facts this repo's own `AGENTS.md`
must never silently lose (`make test`, `go test ./...`, `go vet ./...`, the
`go.mod` module row). The CI dogfood gate can't tell an intentional fact change
from an adapter regression that drops a command; this test can.

### 3. External-corpus benchmark

`scripts/corpus.sh` clones real repositories at pinned SHAs, runs
`repoctx info`, and diffs the extracted commands against committed golden
files. The golden files are the anti-circularity check: they were reviewed by
hand against the pinned checkout (see [results](#results)), so they encode
*expected* facts, not whatever repoctx happens to emit. A regression that
silently adds or drops a command shows up as a diff.

```sh
scripts/corpus.sh            # clone + compare, report
scripts/corpus.sh --fail     # exit non-zero on mismatch (CI)
scripts/corpus.sh --update   # adopt current output as the new golden
```

Add a repository by appending `name<TAB>url<TAB>sha` to
`testdata/corpus/repos.tsv`, run `--update`, and review the diff by hand.

### 4. Dogfood rotation

`scripts/dogfood.sh` runs repoctx against repositories you actually work on
and reports what the scanner extracts and what would change in their context
files. It is read-only and CI-agnostic — the point is to exercise the scanner
against messy, real-world code that the fixture suite never touches.

```sh
scripts/dogfood.sh /path/to/repo [...]
```

## Results

### Corpus (pinned SHA review, 2026-08-12)

| Repo | Ecosystem | Manifests | Commands | Review |
|---|---|---|---|---|
| `spf13/cobra` | Go | `Makefile`, `go.mod` | 12 | Makefile targets match the real targets (`all`, `fmt`, `lint`, `richtest`, ...); Go commands are adapter conventions, plausible for a Go project |
| `expressjs/express` | npm | `package.json` | 6 | All scripts match `package.json` verbatim (`lint`, `test`, `test-ci`, ...) |
| `psf/black` | Python | `pyproject.toml` ×6, `docs/Makefile` | 9 | `pytest` / `mypy .` match the configured tools; `make help` matches `docs/Makefile` |
| `rust-lang/rustlings` | Rust + npm | `Cargo.toml` ×4, `website/package.json` | 20 | Cargo commands are adapter conventions; the root is a `[package]` crate so `cargo run` is legitimate; `website/package.json` contributes none |

Every command in the golden files was checked against the pinned checkout; the
adapter-convention commands (`go build ./...`, `cargo clippy`, ...) are the
documented adapter contract, not literal manifest keys.

Findings that came out of this first pass:

- **`manifests[].commands` is `null`, not `[]`, when a manifest has no
  commands** (e.g. `rustlings/website/package.json`). The documented contract
  (docs/contract.md) says `null` means *extraction failed*, so "no scripts" is
  indistinguishable from "adapter broke". Not changed here (contract is frozen
  post-1.0); worth a contract amendment or a `[]` in a future major.
- **The scanner descends into test-data manifests.** Black's
  `tests/data/*/pyproject.toml` files are compatibility fixtures, not projects,
  but the scanner emits `pytest` for each. The skip list is name-based
  (`.git`, `node_modules`, ...), so there is no way to say "this is test data".
  This inflates command counts on real repos and is the main source of noise.

### Dogfood rotation (2026-08-12)

Run against two real, non-fixture projects:

- `oneblock` (npm + Go polyglot): `package.json` (6 scripts: `build`,
  `build:all`, `build:dev`, `go:build`, `go:build:all`, `package`, `watch`) and
  a nested `cli/go.mod` (10 modules). No `AGENTS.md` / `CLAUDE.md`, so
  `audit`/`generate` were skipped — the first step for this project is
  bootstrapping a context file (`repoctx generate .`).
- `prueba` (Rust single crate): `Cargo.toml` with 3 deps, full command set
  (`cargo build`, `test`, `run`, `fmt --check`, `clippy`). Same bootstrap
  note; its existing `CONTEXTO.md` has no repoctx markers, and files without
  markers are never rewritten by design.

Neither project could be audited yet, which is itself a signal: the audit
story only starts once a context file with markers exists. The rotation script
is the tool to bootstrap and then keep that file honest over time.
