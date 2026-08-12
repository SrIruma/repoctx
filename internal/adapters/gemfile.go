package adapters

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/SrIruma/repoctx/internal/project"
)

type gemfileAdapter struct{}

func (gemfileAdapter) Kind() project.ManifestKind { return project.KindGemfile }
func (gemfileAdapter) Language() string           { return "Ruby" }

// gemLineRe matches `gem "name", "~> 1.0"` and `gem :name, ">= 2"` lines,
// skipping comments and single-quoted gems. The leading `:` or quotes are
// tolerated.
var gemLineRe = regexp.MustCompile(`^\s*gem\s+["']?:?["']?([A-Za-z0-9_.-]+)`)

func (gemfileAdapter) Read(path string) (*ManifestData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var names []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "#") || !strings.HasPrefix(line, "gem ") {
			continue
		}
		if m := gemLineRe.FindStringSubmatch(line); m != nil {
			names = append(names, m[1])
		}
	}
	sort.Strings(names)
	var cmds []project.Command
	for _, n := range names {
		cmds = append(cmds, project.Command{Name: n, Cmd: "bundle exec " + n})
	}
	return &ManifestData{Commands: cmds, Deps: dedupSorted(names)}, nil
}
