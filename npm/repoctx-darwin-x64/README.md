# repoctx

Keep `AGENTS.md`, `CLAUDE.md`, and other agent context files accurate against
the actual repository. repoctx extracts facts from project manifests, refreshes
the code-derived sections of context files, and detects stale commands and
paths.

## Install

```sh
npx @sriruma/repoctx --help
```

Or install it globally:

```sh
npm install --global @sriruma/repoctx
repoctx --help
```

The main package selects the matching prebuilt binary for Windows, macOS, or
Linux. No Go toolchain is required.

## Quick start

```sh
repoctx info .
repoctx init .
repoctx generate .
repoctx audit . --check
```

`generate` updates only the repoctx-managed sections between
`<!-- repoctx:start -->` and `<!-- repoctx:end -->`. Keep human-written
documentation outside those markers.

## More information

- [Documentation and source](https://github.com/SrIruma/repoctx)
- [Releases](https://github.com/SrIruma/repoctx/releases)
- [License](https://github.com/SrIruma/repoctx/blob/main/LICENSE)

Platform-specific packages are implementation details used by
`@sriruma/repoctx`; install the main package rather than a platform package
directly.
