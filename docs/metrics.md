# Metrics and kill criteria — Fase 2

Fase 2 decides whether the skill keeps finding users. These are the signals
to track and the thresholds that trigger a decision.

## What to measure

| Signal | Source | How |
|---|---|---|
| Binary installs | GitHub release downloads | `gh release view` / insights |
| npm installs | npm downloads | `npm view @sriruma/repoctx` downloads |
| Repo stars | GitHub | star count over time |
| Skill stars | GitHub (skill repo) / Agensi | star count |
| Skill installs per channel | Claude marketplace, Agensi, `.well-known` hits | channel dashboards |
| Teams with >3 repos using it | qualitative (issues, emails) | manual |

## Cadence

Weekly snapshot during Fase 2 (8 weeks). Record in a simple table or issue.

## Kill criteria (week 8)

If **all** of the following hold, repoctx stays an OSS/portfolio project and
no further investment goes into it (a valid decision, not a failure):

- < 50 stars on the skill
- < 5 teams using it
- 0 responses from teams with >3 repos

If any signal is strong, keep going.

## Decision date

2027-01-26 (per PLAN.md §8 timeline). Record the decision in `docs/roadmap.md`.
