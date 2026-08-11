package adapters

import (
	"os"
	"regexp"
	"sort"

	"github.com/SrIruma/repoctx/internal/project"
)

type goAdapter struct{}

func (goAdapter) Kind() project.ManifestKind { return project.KindGo }
func (goAdapter) Language() string           { return "Go" }

func (goAdapter) Read(path string) (*ManifestData, error) {
	md := &ManifestData{
		Commands: []project.Command{
			{Name: "build", Cmd: "go build ./..."},
			{Name: "test", Cmd: "go test ./..."},
			{Name: "vet", Cmd: "go vet ./..."},
			{Name: "fmt", Cmd: "gofmt -l ."},
		},
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	md.Deps = goModDeps(data)
	return md, nil
}

var goRequireLine = regexp.MustCompile(`(?m)^\s*([\w./~-]+)\s+v\d+(?:\.\d+)*(?:-[\w.]+)?(?:\s+\/\/.*)?$`)

// goModDeps extracts module paths from require lines (excludes the module
// header line, which has no version, and the go version directive).
func goModDeps(data []byte) []string {
	seen := map[string]bool{}
	var deps []string
	for _, m := range goRequireLine.FindAllSubmatch(data, -1) {
		name := string(m[1])
		if !seen[name] {
			seen[name] = true
			deps = append(deps, name)
		}
	}
	sort.Strings(deps)
	return deps
}
