package adapters

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/SrIruma/repoctx/internal/project"
)

type composerAdapter struct{}

func (composerAdapter) Kind() project.ManifestKind { return project.KindComposer }
func (composerAdapter) Language() string           { return "PHP" }

func (composerAdapter) Read(path string) (*ManifestData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pkg struct {
		Scripts    map[string]string `json:"scripts"`
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		return nil, err
	}

	md := &ManifestData{}
	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		md.Commands = append(md.Commands, project.Command{Name: name, Cmd: "composer run " + name})
	}

	seen := map[string]bool{}
	for _, d := range []map[string]string{pkg.Require, pkg.RequireDev} {
		for name := range d {
			if !seen[name] {
				seen[name] = true
				md.Deps = append(md.Deps, name)
			}
		}
	}
	sort.Strings(md.Deps)
	return md, nil
}
