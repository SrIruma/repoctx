package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// exitError signals a deliberate non-zero exit. The command has already
// printed its output, so Execute returns the code without an error message.
type exitError struct{ code int }

func (e *exitError) Error() string { return fmt.Sprintf("exit code %d", e.code) }

// version is overridden at build time via -ldflags
// (-X github.com/SrIruma/repoctx/internal/cli.version=...).
// It is empty by default, which means no ldflags were applied: the version
// is then read from the build information the Go toolchain embeds in the
// binary, so that `go install module@version` reports the version it was
// installed from.
var version = ""

// resolveVersion picks the version to report. The ldflags-injected value
// wins when present. Otherwise the module version recorded in the build
// information is used, which `go install module@version` populates. A binary
// built from a checkout has no usable version and reports "dev".
func resolveVersion(injected string, info *debug.BuildInfo) string {
	if injected != "" {
		return injected
	}
	if info != nil {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

// buildVersion returns the version of the running binary.
func buildVersion() string {
	info, _ := debug.ReadBuildInfo()
	return resolveVersion(version, info)
}

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "repoctx",
		Short: "Keep your AI coding-agent context files truthful",
		Long: "repoctx scans a repository, extracts facts from the real code\n" +
			"(build/test commands, module structure, dependencies) and keeps\n" +
			"AGENTS.md / CLAUDE.md context files accurate.",
		Version:       buildVersion(),
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
