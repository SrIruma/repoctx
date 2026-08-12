package adapters

import (
	"os"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/SrIruma/repoctx/internal/project"
)

type cargoAdapter struct{}

func (cargoAdapter) Kind() project.ManifestKind { return project.KindCargo }
func (cargoAdapter) Language() string           { return "Rust" }

func (cargoAdapter) Read(path string) (*ManifestData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		// Best-effort: a corrupt manifest still yields the conventional cargo
		// commands, just no dependencies.
		return &ManifestData{Commands: cargoCommands(false)}, nil
	}
	return &ManifestData{
		Commands: cargoCommands(isVirtualWorkspaceRoot(doc)),
		Deps: tomlDependencyNames(doc,
			"dependencies", "dev-dependencies", "build-dependencies", "workspace.dependencies"),
	}, nil
}

// isVirtualWorkspaceRoot reports whether doc is a workspace-root Cargo.toml
// with a [workspace] table and no [package] table of its own.
func isVirtualWorkspaceRoot(doc map[string]any) bool {
	_, hasWorkspace := doc["workspace"].(map[string]any)
	_, hasPackage := doc["package"].(map[string]any)
	return hasWorkspace && !hasPackage
}

// cargoCommands returns the conventional cargo commands. A virtual workspace
// root cannot `cargo run` (there is no default binary), so that command is
// omitted there; the remaining commands operate on the whole workspace.
func cargoCommands(virtualRoot bool) []project.Command {
	cmds := []project.Command{
		{Name: "build", Cmd: "cargo build"},
		{Name: "test", Cmd: "cargo test"},
	}
	if !virtualRoot {
		cmds = append(cmds, project.Command{Name: "run", Cmd: "cargo run"})
	}
	cmds = append(cmds,
		project.Command{Name: "fmt", Cmd: "cargo fmt --check"},
		project.Command{Name: "clippy", Cmd: "cargo clippy"},
	)
	return cmds
}

// tomlDependencyNames extracts dependency names from the given tables of a
// decoded TOML document. Values may be plain version strings or tables
// (including { workspace = true } references), so they are decoded generically.
func tomlDependencyNames(doc map[string]any, tables ...string) []string {
	seen := map[string]bool{}
	var names []string
	for _, table := range tables {
		var t map[string]any
		if i := strings.IndexByte(table, '.'); i >= 0 {
			parent, ok := doc[table[:i]].(map[string]any)
			if !ok {
				continue
			}
			t, ok = parent[table[i+1:]].(map[string]any)
			if !ok {
				continue
			}
		} else {
			t, _ = doc[table].(map[string]any)
		}
		for name := range t {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	return names
}
