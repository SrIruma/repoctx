package adapters

import (
	"os"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/SrIruma/repoctx/internal/project"
)

type pyprojectAdapter struct{}

func (pyprojectAdapter) Kind() project.ManifestKind { return project.KindPyProject }
func (pyprojectAdapter) Language() string           { return "Python" }

func (pyprojectAdapter) Read(path string) (*ManifestData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	md := &ManifestData{Commands: []project.Command{{Name: "test", Cmd: "pytest"}}}
	if t, ok := doc["project"].(map[string]any); ok {
		if reqs, ok := t["dependencies"].([]any); ok {
			for _, r := range reqs {
				if s, ok := r.(string); ok {
					md.Deps = append(md.Deps, reqName(s))
				}
			}
		}
	}
	if tool, ok := doc["tool"].(map[string]any); ok {
		if _, ok := tool["ruff"]; ok {
			md.Commands = append(md.Commands, project.Command{Name: "lint", Cmd: "ruff check ."})
		}
		if _, ok := tool["mypy"]; ok {
			md.Commands = append(md.Commands, project.Command{Name: "typecheck", Cmd: "mypy ."})
		}
	}
	sort.Strings(md.Deps)
	return md, nil
}

// reqName strips version constraints from a requirement string.
// "requests>=2.0" -> "requests", "click==8.1" -> "click".
func reqName(req string) string {
	for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">", "<", "@"} {
		if i := strings.Index(req, sep); i > 0 {
			return req[:i]
		}
	}
	return strings.TrimSpace(req)
}
