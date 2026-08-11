package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/audit"
	"github.com/SrIruma/repoctx/internal/project"
)

const (
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiReset  = "\033[0m"
)

func newAuditCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "audit [dir]",
		Short: "Detect context rot in AGENTS.md / CLAUDE.md",
		Long: "Checks the claims in your context files against the current state of the\n" +
			"code: ghost commands claimed between repoctx markers and stale repository\n" +
			"paths referenced anywhere in the file. Reports a health score per file.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			return runAudit(cmd, dir, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	return cmd
}

func runAudit(cmd *cobra.Command, dir string, jsonOut bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p, err := loadProject(abs)
	if err != nil {
		return err
	}
	var actual []project.Command
	for _, m := range p.Manifests {
		actual = append(actual, m.Commands...)
	}

	var reports []*audit.Report
	for _, name := range contextFiles(abs) {
		r, err := audit.Run(audit.Options{
			Root:   abs,
			File:   filepath.Join(abs, name),
			Actual: actual,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		reports = append(reports, r)
	}
	if len(reports) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no context files (AGENTS.md, CLAUDE.md) found in", abs)
		return nil
	}
	if jsonOut {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(reports)
	}
	for _, r := range reports {
		printReport(cmd.OutOrStdout(), r)
	}
	return nil
}

// contextFiles returns the known context files that exist in dir.
func contextFiles(dir string) []string {
	var out []string
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
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
