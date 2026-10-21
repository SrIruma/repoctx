package cli

import "github.com/SrIruma/repoctx/internal/project"

// contextTargets decides which context files generate and audit operate on.
//
// Precedence:
//   - an explicit --file flag or a `files` list in repoctx.toml wins outright;
//     exactly those names are used and the tree is not searched.
//   - otherwise the context files discovered in the tree are used, so a
//     monorepo keeps one context file per package.
//   - when discovery finds nothing, fallback is returned so the commands keep
//     their historical default (generate creates the root file; audit reports
//     that there is nothing to audit).
func contextTargets(p *project.Project, opts resolved, fallback []string) []string {
	if opts.filesConfigured {
		return opts.files
	}
	if len(p.ContextFiles) > 0 {
		return p.ContextFiles
	}
	return fallback
}
