# Agensi listing — repoctx (free)

Copy-paste material for the free listing in the Agensi creator dashboard.

This listing is the top of the funnel: it is what a buyer discovers, and it
is what leads to `repoctx+`. Publish it even if the paid listing is still in
review — a free listing can go live on its own, and it does not need payouts
connected.

Specs verified against Agensi's creator docs on 2026-10-21. Where a number
appears, it is a platform limit.

## Name

```
repoctx — keep your agent context file honest
```

Slug: `repoctx`. The frontmatter `name:` in `SKILL.md` is `repoctx` and must
match the package folder name exactly.

## Summary (card, max 200 characters)

```
Keep AGENTS.md, CLAUDE.md and .cursorrules truthful: regenerate the code-derived sections and audit for rot — ghost commands, stale paths.
```

## Description (long)

> Your agent reads `AGENTS.md` every session. It was written once and nobody
> updated it. The `npm run deploy` command was renamed, the package moved,
> the dependency was dropped. Your agent does not know: it reads a file that
> lies.
>
> repoctx keeps that file honest. It reads the real manifests in your
> repository, extracts the commands and module structure as facts, and
> regenerates the code-derived sections of your context file between the
> markers. Then it audits the file and reports what is stale.
>
> **What it catches**
> - **Ghost commands** — a command the context file claims that no longer
>   exists in any manifest.
> - **Stale paths** — a file it references that moved or was deleted.
>
> **What it never touches**
> The prose. The overview, the conventions, the TODOs are yours; repoctx
> only owns the tables between the markers.
>
> **It works in CI.** `repoctx audit . --check` exits non-zero when the
> context drifted, so a context file that lies fails the build instead of the
> next agent that reads it.
>
> repoctx is free and MIT. The engine is open source:
> https://github.com/SrIruma/repoctx
>
> Need it for a fleet of repositories instead of one? That is the paid tier,
> `repoctx+`: per-stack playbooks, tuned `repoctx.toml` policies, `AGENTS.md`
> templates and a fleet-wide audit.

## Category / tags

Up to 3 categories; the first is the primary and decides the marketplace
page. Picking categories rewrites the category tags below.

1. Code Quality & Review (primary)
2. Technical Documentation
3. DevOps & Cloud

Tags: `agents-md`, `claude-md`, `context-drift`, `code-quality`, `documentation`,
`ci`, `devops`, `open-source`.

## Price

**Free.** Free listings can be published without payouts connected; the
payout connection is only required for a paid skill.

## Logo

- **512×512 px**, PNG, JPG, WebP or GIF, up to 2 MB.
- The repoctx logo: dark background, the wordmark in the repoctx style.
- A logo plus one screenshot measurably increases installs.

## Screenshots and demo (optional)

- Screenshots: optional, up to 10 MB each.
- Demo video: a 30-second screen recording roughly doubles conversion. The
  best 30 seconds for this skill: run `repoctx audit .` on a repository whose
  context file is stale and show the report.

## Package

- File: `dist/repoctx-skill.zip` (built with `make skill-package`).
- Structure: one top-level folder `repoctx/` with `SKILL.md` at its root.
- Contents: `SKILL.md`, `README.md`, `references/patterns.md`. Text files
  only — no binaries, no symlinks, no hidden OS files.
- The marketplace can auto-generate name, description, tags, use cases, FAQ
  and a sample input/output from the package. Read the generated listing
  before submitting: the review checks that the listing says what the skill
  does.

## Review

24-48 hours: automated security scan, an AI reviewer that reads the skill for
safety and honest intent, then a human review. Updates to a live skill are
free for existing buyers.

## Checklist

- [x] ZIP built with `make skill-package` → `dist/repoctx-skill.zip`
- [x] One top-level folder `repoctx/`, `SKILL.md` at its root
- [x] Frontmatter `name: repoctx` matches the folder
- [x] Supporting file included (`references/patterns.md`)
- [x] `README.md` inside the package folder
- [x] Price set to free
- [ ] Logo uploaded (512×512, ≤2 MB)
- [ ] At least one screenshot uploaded
- [ ] Description pasted from this file
- [ ] Link to the paid listing added to the description and to the package
      README — **add it once the repoctx+ listing URL exists**
