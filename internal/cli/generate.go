package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/markdown"
	"github.com/SrIruma/repoctx/internal/project"
)

func newGenerateCmd() *cobra.Command {
	var files []string
	var dryRun bool
	var opts scanFlags
	cmd := &cobra.Command{
		Use:   "generate [dir]",
		Short: "Regenerate the code-derived sections of context files",
		Long: "Scans the repository, extracts the real build/test commands, and rewrites\n" +
			"the content between repoctx markers in AGENTS.md / CLAUDE.md. Human-written\n" +
			"content outside the markers is preserved verbatim.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			flags := scanFlags{
				maxDepth:    opts.maxDepth,
				skipDirs:    opts.skipDirs,
				configPath:  opts.configPath,
				files:       files,
				maxDepthSet: cmd.Flags().Changed("max-depth"),
				skipDirsSet: cmd.Flags().Changed("skip-dirs"),
				filesSet:    cmd.Flags().Changed("file"),
			}
			res, err := flags.resolve(dir)
			if err != nil {
				return err
			}
			return runGenerate(cmd, dir, res.files, res, dryRun)
		},
	}
	cmd.Flags().StringSliceVar(&files, "file", nil,
		"context file to update (repeatable, default AGENTS.md)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"report what would change without writing any file")
	addScanFlags(cmd, &opts)
	return cmd
}

func runGenerate(cmd *cobra.Command, dir string, files []string, opts resolved, dryRun bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p, err := loadProject(abs, opts)
	if err != nil {
		return err
	}
	targets := contextTargets(p, opts, files)
	for _, f := range targets {
		path := f
		if !filepath.IsAbs(path) {
			path = filepath.Join(abs, f)
		}
		rendered := markdown.RenderSection(tableRows(p, scopeOf(f)), moduleRows(p, scopeOf(f)))
		out, changed, err := planFile(path, rendered)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		if dryRun {
			if changed {
				fmt.Fprintf(cmd.OutOrStdout(), "would update %s\n", f)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is up to date\n", f)
			}
			continue
		}
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "generated %s\n", f)
	}
	return nil
}

// scopeOf returns the directory a context file owns. A context file at
// docs/AGENTS.md owns the "docs" subtree; a file at the repository root owns
// everything and is represented by ".".
func scopeOf(file string) string {
	dir := filepath.ToSlash(filepath.Dir(file))
	if dir == "." || dir == "" {
		return "."
	}
	return dir
}

// planFile computes the content path would get after a generation and whether
// it differs from what is on disk. A missing file is planned for creation with
// a fresh marked section. Existing files without markers are an error, never
// half-rewritten.
func planFile(path, rendered string) (out string, changed bool, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		doc := "# Commands\n\n" + markdown.CanonicalBlock(strings.TrimSpace(rendered)) + "\n"
		return doc, true, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	out, err = markdown.Update(string(data), rendered)
	if err != nil {
		return "", false, err
	}
	return out, out != string(data), nil
}

// tableRows flattens every manifest in scope into renderable rows, in the
// deterministic manifest-then-command order produced by the scanner. Paths are
// relative to scope so a nested context file describes its own subtree.
func tableRows(p *project.Project, scope string) []markdown.Row {
	var rows []markdown.Row
	for _, m := range p.Manifests {
		if !m.Canonical || !inScope(m.Scope, scope) {
			continue
		}
		for _, c := range m.Commands {
			rows = append(rows, markdown.Row{Command: c.Cmd, Source: relToScope(m.Path, scope)})
		}
	}
	return rows
}

// moduleRows flattens every manifest in scope into a renderable modules table
// row, with paths relative to scope.
func moduleRows(p *project.Project, scope string) []markdown.ModuleRow {
	var rows []markdown.ModuleRow
	for _, m := range p.Manifests {
		if !m.Canonical || !inScope(m.Scope, scope) {
			continue
		}
		rows = append(rows, markdown.ModuleRow{
			Path:     relToScope(m.Path, scope),
			Language: m.Language,
			Deps:     strings.Join(m.Deps, ", "),
		})
	}
	return rows
}

// inScope reports whether a manifest located in manifestScope belongs to the
// subtree owned by scope. The root scope (".") owns everything.
func inScope(manifestScope, scope string) bool {
	if scope == "." {
		return true
	}
	return manifestScope == scope || strings.HasPrefix(manifestScope, scope+"/")
}

// relToScope rewrites a manifest path to be relative to scope. The root scope
// keeps the path unchanged.
func relToScope(path, scope string) string {
	if scope == "." {
		return path
	}
	rel, err := filepath.Rel(scope, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
