package adapters

import (
	"os"
	"sort"

	"github.com/pelletier/go-toml/v2"

	"github.com/SrIruma/repoctx/internal/project"
)

type cargoAdapter struct{}

func (cargoAdapter) Kind() project.ManifestKind { return project.KindCargo }
func (cargoAdapter) Language() string           { return "Rust" }

func (cargoAdapter) Read(path string) (*ManifestData, error) {
	md := &ManifestData{
		Commands: []project.Command{
			{Name: "build", Cmd: "cargo build"},
			{Name: "test", Cmd: "cargo test"},
			{Name: "run", Cmd: "cargo run"},
			{Name: "fmt", Cmd: "cargo fmt --check"},
			{Name: "clippy", Cmd: "cargo clippy"},
		},
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	md.Deps = tomlDependencyNames(data, "dependencies", "dev-dependencies", "build-dependencies")
	return md, nil
}

// tomlDependencyNames extracts dependency names from the given TOML tables.
// Values may be plain version strings or tables, so they are decoded generically.
func tomlDependencyNames(data []byte, tables ...string) []string {
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil
	}
	seen := map[string]bool{}
	var names []string
	for _, table := range tables {
		if t, ok := doc[table].(map[string]any); ok {
			for name := range t {
				if !seen[name] {
					seen[name] = true
					names = append(names, name)
				}
			}
		}
	}
	sort.Strings(names)
	return names
}
