package adapters

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/SrIruma/repoctx/internal/project"
)

type cmakeAdapter struct{}

func (cmakeAdapter) Kind() project.ManifestKind { return project.KindCMake }
func (cmakeAdapter) Language() string           { return "C/C++" }

// findPackageRe matches find_package(Foo [version ...]) and captures the name.
var findPackageRe = regexp.MustCompile(`find_package\(\s*([A-Za-z0-9_:]+)`)

// customTargetRe matches a single-line add_custom_target(NAME ...) call.
var customTargetRe = regexp.MustCompile(`add_custom_target\(\s*["']?([A-Za-z0-9_-]+)`)

func (cmakeAdapter) Read(path string) (*ManifestData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var targets []string
	var deps []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if m := findPackageRe.FindStringSubmatch(line); m != nil {
			deps = append(deps, m[1])
			continue
		}
		// add_custom_target may span several lines; accumulate until the
		// parenthesis balance closes (1-level, no nested parens).
		stmt := line
		if strings.HasPrefix(stmt, "add_custom_target(") {
			for strings.Count(stmt, "(") > strings.Count(stmt, ")") && sc.Scan() {
				stmt += " " + strings.TrimSpace(sc.Text())
			}
			if m := customTargetRe.FindStringSubmatch(stmt); m != nil {
				targets = append(targets, m[1])
			}
		}
	}
	sort.Strings(targets)
	var cmds []project.Command
	for _, t := range targets {
		cmds = append(cmds, project.Command{Name: t, Cmd: "cmake --build build --target " + t})
	}
	sort.Strings(deps)
	return &ManifestData{Commands: cmds, Deps: dedupSorted(deps)}, nil
}
