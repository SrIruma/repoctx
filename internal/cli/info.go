package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/adapters"
	"github.com/SrIruma/repoctx/internal/project"
)

func newInfoCmd() *cobra.Command {
	var jsonOut bool
	var opts scanFlags
	cmd := &cobra.Command{
		Use:   "info [dir]",
		Short: "Detect manifests and extract facts from a repository",
		Args:  cobra.MaximumNArgs(1),
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
			return runInfo(cmd, dir, jsonOut, res)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	addScanFlags(cmd, &opts)
	return cmd
}

func runInfo(cmd *cobra.Command, dir string, jsonOut bool, opts resolved) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p, err := loadProject(abs, opts)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if jsonOut {
		return printJSON(w, p)
	}
	return printHuman(w, abs, p)
}

// loadProject scans dir and enriches every manifest with the facts extracted
// by its adapter. Extraction is best-effort: unreadable manifests are kept as
// detected but without facts.
func loadProject(dir string, opts resolved) (*project.Project, error) {
	sc := project.NewScanner(dir)
	sc.MaxDepth = opts.maxDepth
	sc.SkipDirs = opts.skipDirs
	p, err := sc.Scan()
	if err != nil {
		return nil, err
	}
	for _, m := range p.Manifests {
		ad, ok := adapters.For(m.Kind)
		if !ok {
			continue
		}
		md, err := ad.Read(filepath.Join(dir, m.Path))
		if err != nil {
			m.Errors = append(m.Errors, err.Error())
			continue
		}
		m.Language = ad.Language()
		m.Commands = md.Commands
		if m.Commands == nil {
			// A successful extraction with no commands must serialize as []
			// (not null). null is reserved for extraction failure, signalled
			// by the errors field below.
			m.Commands = []project.Command{}
		}
		m.Deps = md.Deps
	}
	return p, nil
}

func printJSON(w io.Writer, p *project.Project) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

func printHuman(w io.Writer, root string, p *project.Project) error {
	if len(p.Manifests) == 0 && len(p.DetectedOther) == 0 {
		fmt.Fprintf(w, "No manifests detected in %s\n", root)
		return nil
	}
	fmt.Fprintf(w, "Detected manifests in %s:\n", root)
	for _, m := range p.Manifests {
		names := make([]string, 0, len(m.Commands))
		for _, c := range m.Commands {
			names = append(names, c.Name)
		}
		fmt.Fprintf(w, "  %-28s %-10s %-22s commands: [%s]  (%d deps)\n",
			m.Path, m.Kind, m.Language, strings.Join(names, ", "), len(m.Deps))
		for _, err := range m.Errors {
			fmt.Fprintf(w, "  ! %s: %s\n", m.Path, err)
		}
	}
	for _, o := range p.DetectedOther {
		fmt.Fprintf(w, "  ! %s\n", o)
	}
	return nil
}
