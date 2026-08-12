package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// exitError signals a deliberate non-zero exit. The command has already
// printed its output, so Execute returns the code without an error message.
type exitError struct{ code int }

func (e *exitError) Error() string { return fmt.Sprintf("exit code %d", e.code) }

// version is overridden at build time via -ldflags
// (-X github.com/SrIruma/repoctx/internal/cli.version=...).
var version = "0.0.1"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "repoctx",
		Short: "Keep your AI coding-agent context files truthful",
		Long: "repoctx scans a repository, extracts facts from the real code\n" +
			"(build/test commands, module structure, dependencies) and keeps\n" +
			"AGENTS.md / CLAUDE.md context files accurate.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newInfoCmd())
	root.AddCommand(newGenerateCmd())
	root.AddCommand(newAuditCmd())
	root.AddCommand(newWorkflowCmd())
	return root
}

func Execute() {
	os.Exit(execute(NewRootCmd(), os.Stdout, os.Stderr))
}

// execute runs root and returns the process exit code. Errors wrapped in
// *exitError map to their code silently; anything else is reported and
// returns 1.
func execute(root *cobra.Command, stdout, stderr io.Writer) int {
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			return ee.code
		}
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}
