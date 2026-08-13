# Live fixture — a hand-maintained CLAUDE.md

This file simulates a context file that a human edits by hand, so it contains
rot on purpose. repoctx does not manage it; the audit must still find both
kinds of drift below.

The entry point is `src/index.js`. See `README.md` for the setup. The design
notes in `docs/architecture-2024.md` were deleted last quarter but this file
still references them (stale path).

## Commands

<!-- repoctx:start -->
| Command | Source |
|---|---|
| `npm run test` | package.json |
| `npm run lint` | package.json |
| `npm run deploy` | package.json |
<!-- repoctx:end -->
