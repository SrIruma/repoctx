# Roadmap

Estado actual: **v0.9.0** (Fase 2 cerrada). Contrato en 0.x, no congelado.

## Completado — Fase 1 · v0.8.0 "bootstrap"

- [x] `repoctx init` — esqueleto de contexto desde cero
- [x] `repoctx skill` — SKILL.md canónico generado
- [x] `docs/roadmap.md` público

## Completado — Fase 2 · v0.9.0 "fleet + skill gratis"

- [x] `repoctx audit` multi-repo
- [x] `fleet.toml` / roots en config
- [x] `audit --json` con `roots[]` + `reports[]`
- [x] SKILL.md canónico publicado en 8+ canales (in-repo)
- [x] Instrumentación de métricas (docs/metrics.md)

## Después

- Fase 3 — skill de pago (playbooks, policy, fleet) · v0.10.0
- Fase 4 — MCP server (enforcement + histórico) · v0.11.0 (repo aparte)

## Notas

- El motor `repoctx` es MIT para siempre. Las fases 1-2 son OSS igual que la 0.
- Semver: fix → patch, feature → minor, cambio de contrato → major (y se avisa).
- Cada fase termina con un release tagueado y un milestone cerrado en GitHub.
