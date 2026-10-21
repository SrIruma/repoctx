# Roadmap

Estado actual: **v0.7.0** (Fase 0 cerrada). Contrato en 0.x, no congelado.

## En curso — Fase 1 · v0.8.0 "bootstrap"

- [x] `repoctx init` — esqueleto de contexto desde cero
- [x] `repoctx skill` — SKILL.md canónico generado
- [ ] `docs/roadmap.md` (este archivo)

## Próximo — Fase 2 · v0.9.0 "fleet + skill gratis"

- [ ] `repoctx audit` multi-repo
- [ ] `fleet.toml` / roots en config
- [ ] `audit --json` con `roots[]` + `reports[]`
- [ ] SKILL.md canónico publicado en 8+ canales
- [ ] Instrumentación de métricas

## Después

- Fase 3 — skill de pago (playbooks, policy, fleet) · v0.10.0
- Fase 4 — MCP server (enforcement + histórico) · v0.11.0 (repo aparte)

## Notas

- El motor `repoctx` es MIT para siempre. Las fases 1-2 son OSS igual que la 0.
- Semver: fix → patch, feature → minor, cambio de contrato → major (y se avisa).
- Cada fase termina con un release tagueado y un milestone cerrado en GitHub.
