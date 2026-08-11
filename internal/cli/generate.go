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
			return runGenerate(cmd, dir, files)
		},
	}
	cmd.Flags().StringSliceVar(&files, "file", []string{"AGENTS.md"},
		"context file to update (repeatable, default AGENTS.md)")
	return cmd
}

func runGenerate(cmd *cobra.Command, dir string, files []string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p, err := loadProject(abs)
	if err != nil {
		return err
	}
	rendered := markdown.RenderSection(tableRows(p), moduleRows(p))
	for _, f := range files {
		path := f
		if !filepath.IsAbs(path) {
			path = filepath.Join(abs, f)
		}
		if err := generateFile(path, rendered); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "generated %s\n", f)
	}
	return nil
}

// generateFile writes rendered content into the marked section of path. If the
// file does not exist yet it is created with a fresh marked section. Existing
// files without markers are left untouched (error), never half-rewritten.
func generateFile(path, rendered string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		doc := "# Commands\n\n" + markdown.CanonicalBlock(strings.TrimSpace(rendered)) + "\n"
		return os.WriteFile(path, []byte(doc), 0o644)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := markdown.Update(string(data), rendered)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

// tableRows flattens every manifest's commands into renderable rows, in the
// deterministic manifest-then-command order produced by the scanner.
func tableRows(p *project.Project) []markdown.Row {
	var rows []markdown.Row
	for _, m := range p.Manifests {
		for _, c := range m.Commands {
			rows = append(rows, markdown.Row{Command: c.Cmd, Source: m.Path})
		}
	}
	return rows
}

// moduleRows flattens every manifest into a renderable modules table row.
func moduleRows(p *project.Project) []markdown.ModuleRow {
	var rows []markdown.ModuleRow
	for _, m := range p.Manifests {
		rows = append(rows, markdown.ModuleRow{
			Path:     m.Path,
			Language: m.Language,
			Deps:     strings.Join(m.Deps, ", "),
		})
	}
	return rows
}
