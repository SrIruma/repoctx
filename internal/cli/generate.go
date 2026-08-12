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
	rendered := markdown.RenderSection(tableRows(p), moduleRows(p))
	for _, f := range files {
		path := f
		if !filepath.IsAbs(path) {
			path = filepath.Join(abs, f)
		}
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
