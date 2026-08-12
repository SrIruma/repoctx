package adapters

import (
	"bufio"
	"os"
	"regexp"
	"sort"

	"github.com/SrIruma/repoctx/internal/project"
)

type mesonAdapter struct{}

func (mesonAdapter) Kind() project.ManifestKind { return project.KindMeson }
func (mesonAdapter) Language() string           { return "C/C++" }

// dependencyRe matches dependency('foo' [, ...]) and captures the name.
var dependencyRe = regexp.MustCompile(`dependency\(\s*['"]([^'"]+)`)

func (mesonAdapter) Read(path string) (*ManifestData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var deps []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m := dependencyRe.FindStringSubmatch(sc.Text()); m != nil {
			deps = append(deps, m[1])
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.Strings(deps)
	return &ManifestData{
		Commands: []project.Command{
			{Name: "setup", Cmd: "meson setup build"},
			{Name: "compile", Cmd: "meson compile -C build"},
			{Name: "test", Cmd: "meson test -C build"},
		},
		Deps: dedupSorted(deps),
	}, nil
}
