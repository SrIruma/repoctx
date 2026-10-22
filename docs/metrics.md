# Metrics and kill criteria — Fase 2

Fase 2 decides whether the paid tier (Fase 3/4) is justified. These are the
signals to track and the thresholds that trigger a decision.

## What to measure

| Signal | Source | How |
|---|---|---|
| Binary installs | GitHub release downloads | `gh release view` / insights |
| npm installs | npm downloads | `npm view @sriruma/repoctx` downloads |
| Repo stars | GitHub | star count over time |
| Skill stars | GitHub (skill repo) / Agensi | star count |
| Skill installs per channel | Claude marketplace, Agensi, `.well-known` hits | channel dashboards |
| Teams with >3 repos using it | qualitative (issues, emails) | manual |
| Paid listing funnel | Agensi Creator Studio | page views → add to carts → conversions → revenue |

## Cadence

Weekly snapshot during Fase 2 (8 weeks). Record in a simple table or issue.

The paid listing on Agensi is the shortest path to a real signal: it shows
page views, carts, conversions and revenue for the thing you are actually
selling. Check it weekly from the day the listing goes live.

## Kill criteria (week 8)

If **all** of the following hold, repoctx stays OSS/portfolio and the paid tier
is postponed (a valid decision, not a failure):

- < 50 stars on the skill
- < 5 teams using it
- 0 responses from teams with >3 repos

If any signal is strong, proceed to Fase 3.

## Decision date

2026-12-14 (per PLAN.md §8 timeline). Record the decision in `docs/roadmap.md`.

The date moved up from 2027-01-26 on 2026-10-21: the paid listing on Agensi
is the measurement instrument, and it exists from the day it is published, so
there is no reason to wait for a separate phase to finish.
