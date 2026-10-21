package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/audit"
	"github.com/SrIruma/repoctx/internal/config"
	"github.com/SrIruma/repoctx/internal/project"
)

const (
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiReset  = "\033[0m"
)

func newAuditCmd() *cobra.Command {
	var jsonOut, check bool
	var files []string
	var opts scanFlags
	cmd := &cobra.Command{
		Use:   "audit [dir]",
		Short: "Detect context rot in AGENTS.md / CLAUDE.md",
		Long: "Checks the claims in your context files against the current state of the\n" +
			"code: ghost commands claimed between repoctx markers and stale repository\n" +
			"paths referenced anywhere in the file. Reports a health score per file.\n" +
			"With --check the command exits non-zero when any file fails, so it can\n" +
			"gate merges in CI.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			roots := args
			if len(roots) == 0 {
				roots = []string{"."}
			}
			// Fleet mode: a single root whose config lists `roots` expands to
			// multiple roots, resolved relative to the config directory.
			if len(roots) == 1 {
				cfg, err := config.Load(roots[0])
				if err == nil && len(cfg.Roots) > 0 {
					base := roots[0]
					absBase, _ := filepath.Abs(base)
					expanded := make([]string, 0, len(cfg.Roots))
					for _, r := range cfg.Roots {
						if filepath.IsAbs(r) {
							expanded = append(expanded, r)
						} else {
							expanded = append(expanded, filepath.Join(absBase, r))
						}
					}
					roots = expanded
				}
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
			return runAudit(cmd, roots, jsonOut, check, flags)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&check, "check", false, "exit non-zero when any file fails (CI gating)")
	cmd.Flags().StringSliceVar(&files, "file", nil,
		"context file to audit (repeatable, replaces config/defaults)")
	addScanFlags(cmd, &opts)
	return cmd
}

func runAudit(cmd *cobra.Command, roots []string, jsonOut, check bool, flags scanFlags) error {
	type rootReport struct {
		Root  string          `json:"root"`
		Files []*audit.Report `json:"files"`
	}
	var allReports []rootReport
	var flatReports []*audit.Report
	totalPassed := 0
	totalFiles := 0
	anyFailed := false

	for _, r := range roots {
		abs, err := filepath.Abs(r)
		if err != nil {
			return err
		}
		res, err := flags.resolve(r)
		if err != nil {
			return err
		}
		p, err := loadProject(abs, res)
		if err != nil {
			return err
		}
		var actual []project.Command
		for _, m := range p.Manifests {
			if !m.Canonical {
				continue
			}
			actual = append(actual, m.Commands...)
		}
		targets := contextTargets(p, res, defaultContextFiles)
		var reports []*audit.Report
		for _, name := range contextFiles(abs, targets) {
			rep, err := audit.Run(audit.Options{
				Root:   abs,
				File:   filepath.Join(abs, name),
				Actual: actual,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			reports = append(reports, rep)
		}
		for _, rep := range reports {
			if !rep.Passed {
				anyFailed = true
			}
			totalFiles++
			if rep.Passed {
				totalPassed++
			}
		}
		allReports = append(allReports, rootReport{Root: abs, Files: reports})
		flatReports = append(flatReports, reports...)

		if !jsonOut && len(roots) > 1 {
			fmt.Fprintf(cmd.OutOrStdout(), "=== %s ===\n", abs)
		}
		if len(reports) == 0 {
			if !jsonOut {
				fmt.Fprintf(cmd.OutOrStdout(), "no context files (%s) found in %s\n",
					strings.Join(targets, ", "), abs)
			}
		}
		if !jsonOut {
			for _, rep := range reports {
				printReport(cmd.OutOrStdout(), rep)
			}
		}
	}

	if jsonOut {
		if len(roots) == 1 {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			if err := enc.Encode(flatReports); err != nil {
				return err
			}
		} else {
			fleetScore := 0
			if totalFiles > 0 {
				fleetScore = totalPassed * 100 / totalFiles
			}
			out := struct {
				Roots      []string     `json:"roots"`
				Reports    []rootReport `json:"reports"`
				FleetScore int          `json:"fleet_score"`
				Passed     bool         `json:"passed"`
			}{
				Roots: func() []string {
					r := make([]string, 0, len(allReports))
					for _, x := range allReports {
						r = append(r, x.Root)
					}
					return r
				}(),
				Reports:    allReports,
				FleetScore: fleetScore,
				Passed:     !anyFailed,
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			if err := enc.Encode(out); err != nil {
				return err
			}
		}
	}

	if len(roots) > 1 && !jsonOut {
		fleetScore := 0
		if totalFiles > 0 {
			fleetScore = totalPassed * 100 / totalFiles
		}
		status := "PASS"
		if anyFailed {
			status = "FAIL"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\nFleet: %s  score %d/100  (%d/%d files passed)\n",
			status, fleetScore, totalPassed, totalFiles)
	}

	if check && anyFailed {
		return &exitError{code: 1}
	}
	return nil
}

// defaultContextFiles are the context files audited when the project does not
// configure its own and the scan discovers none.
var defaultContextFiles = []string{"AGENTS.md", "CLAUDE.md"}

// contextFiles returns the known context files that exist in dir.
func contextFiles(dir string, names []string) []string {
	var out []string
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			out = append(out, name)
		}
	}
	return out
}

func printReport(w io.Writer, r *audit.Report) {
	status := colorize(w, ansiGreen, "PASS")
	if !r.Passed {
		status = colorize(w, ansiRed, "FAIL")
	}
	fmt.Fprintf(w, "%s  %s  score %d/100\n", status, r.File, r.Score)
	for _, c := range r.Checks {
		mark := colorize(w, ansiGreen, "ok ")
		if !c.Passed {
			mark = colorize(w, ansiRed, "!! ")
		}
		fmt.Fprintf(w, "  %s %s: %s\n", mark, c.Name, c.Detail)
		for _, iss := range c.Issues {
			fmt.Fprintf(w, "    - %s: %s\n", colorize(w, ansiYellow, issueLabel(iss)), iss.Detail)
		}
	}
}

func issueLabel(iss audit.Issue) string {
	switch {
	case iss.Command != "":
		return iss.Command
	case iss.Path != "":
		return iss.Path
	}
	return "issue"
}

func colorize(w io.Writer, color, s string) string {
	if isTerminal(w) {
		return color + s + ansiReset
	}
	return s
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
