package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
