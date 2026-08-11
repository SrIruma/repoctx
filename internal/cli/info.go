package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/adapters"
	"github.com/SrIruma/repoctx/internal/project"
)

func newInfoCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "info [dir]",
		Short: "Detect manifests and extract facts from a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			return runInfo(dir, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	return cmd
}

func runInfo(dir string, jsonOut bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p, err := loadProject(abs)
	if err != nil {
		return err
	}
	if jsonOut {
		return printJSON(p)
	}
	return printHuman(abs, p)
}

// loadProject scans dir and enriches every manifest with the facts extracted
// by its adapter. Extraction is best-effort: unreadable manifests are kept as
// detected but without facts.
func loadProject(dir string) (*project.Project, error) {
	sc := project.NewScanner(dir)
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
			continue
		}
		m.Language = ad.Language()
		m.Commands = md.Commands
		m.Deps = md.Deps
	}
	return p, nil
}

func printJSON(p *project.Project) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

func printHuman(root string, p *project.Project) error {
	if len(p.Manifests) == 0 && len(p.DetectedOther) == 0 {
		fmt.Println("No manifests detected in", root)
		return nil
	}
	fmt.Printf("Detected manifests in %s:\n", root)
	for _, m := range p.Manifests {
		names := make([]string, 0, len(m.Commands))
		for _, c := range m.Commands {
			names = append(names, c.Name)
		}
		fmt.Printf("  %-28s %-10s %-22s commands: [%s]  (%d deps)\n",
			m.Path, m.Kind, m.Language, strings.Join(names, ", "), len(m.Deps))
	}
	for _, o := range p.DetectedOther {
		fmt.Printf("  ! %s\n", o)
	}
	return nil
}
