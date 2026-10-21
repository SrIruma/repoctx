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

func newInitCmd() *cobra.Command {
	var file string
	var force, dryRun bool
	var opts scanFlags
	cmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Bootstrap a new context file from the repository",
		Long: "Creates a context file (AGENTS.md by default) from scratch: an overview,\n" +
			"structure notes, the code-derived commands and modules tables between\n" +
			"repoctx markers, and a maintenance workflow block. Human-only sections are\n" +
			"left as TODO markers. Existing files are never overwritten unless --force.",
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
				maxDepthSet: cmd.Flags().Changed("max-depth"),
				skipDirsSet: cmd.Flags().Changed("skip-dirs"),
			}
			res, err := flags.resolve(dir)
			if err != nil {
				return err
			}
			return runInit(cmd, dir, file, res, force, dryRun)
		},
	}
	cmd.Flags().StringVar(&file, "file", "AGENTS.md",
		"context file to create (default AGENTS.md)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the skeleton without writing")
	addScanFlags(cmd, &opts)
	return cmd
}

func runInit(cmd *cobra.Command, dir, file string, opts resolved, force, dryRun bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p, err := loadProject(abs, opts)
	if err != nil {
		return err
	}
	target := file
	if !filepath.IsAbs(target) {
		target = filepath.Join(abs, file)
	}
	if !force {
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("%s already exists; use --force to overwrite", file)
		}
	}
	rendered := markdown.RenderSection(tableRows(p, "."), moduleRows(p, "."))
	skeleton := initSkeleton(filepath.Base(abs), file, rendered, p)
	if dryRun {
		fmt.Fprint(cmd.OutOrStdout(), skeleton)
		return nil
	}
	if err := os.WriteFile(target, []byte(skeleton), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", file)
	return nil
}

func initSkeleton(repoName, file, rendered string, p *project.Project) string {
	hint := conventionHint(p)
	var b strings.Builder
	b.WriteString("# " + repoName + "\n\n")
	b.WriteString("<!-- TODO: one-line description of this repository -->\n\n")
	b.WriteString("## Overview\n\n")
	b.WriteString("<!-- TODO: what this project does and who it's for -->\n\n")
	if hint != "" {
		b.WriteString("## Conventions\n\n")
		b.WriteString(hint + "\n\n")
	}
	b.WriteString("## Structure\n\n")
	b.WriteString("<!-- TODO: high-level layout of the repository -->\n\n")
	b.WriteString(markdown.CanonicalBlock(strings.TrimSpace(rendered)) + "\n\n")
	b.WriteString(workflowBlock([]string{file}) + "\n")
	return b.String()
}

func conventionHint(p *project.Project) string {
	if len(p.Manifests) == 0 {
		return ""
	}
	kinds := map[project.ManifestKind]string{
		project.KindNPM:       "JavaScript/TypeScript project: scripts live in package.json; run them with the detected package manager.",
		project.KindCargo:     "Rust project: build and test with cargo.",
		project.KindGo:        "Go project: build and test with go.",
		project.KindPyProject: "Python project: dependencies and scripts in pyproject.toml.",
		project.KindMake:      "Make-based project: targets are defined in the Makefile.",
		project.KindCMake:     "C/C++ project built with CMake.",
		project.KindGemfile:   "Ruby project: dependencies in the Gemfile.",
		project.KindComposer:  "PHP project: dependencies in composer.json.",
		project.KindMaven:     "Java project built with Maven.",
		project.KindGradle:    "JVM project built with Gradle.",
		project.KindMeson:     "Project built with Meson.",
	}
	// Use the first manifest's kind as the primary hint.
	if h, ok := kinds[p.Manifests[0].Kind]; ok {
		return h
	}
	return ""
}
