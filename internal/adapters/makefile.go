package adapters

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/SrIruma/repoctx/internal/project"
)

type makeAdapter struct{}

func (makeAdapter) Kind() project.ManifestKind { return project.KindMake }
func (makeAdapter) Language() string           { return "Generic (Make)" }

// makeTarget matches a target rule line: "build:" but not "%.o: %.c" or ".PHONY:".
var makeTarget = regexp.MustCompile(`(?m)^([A-Za-z0-9._/-]+):[^=]`)

func (makeAdapter) Read(path string) (*ManifestData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	md := &ManifestData{}
	for _, m := range makeTarget.FindAllStringSubmatch(string(data), -1) {
		t := strings.TrimSpace(m[1])
		if t == "" || strings.HasPrefix(t, ".") || strings.Contains(t, "%") {
			continue
		}
		if !seen[t] {
			seen[t] = true
			md.Commands = append(md.Commands, project.Command{Name: t, Cmd: "make " + t})
		}
	}
	sort.Slice(md.Commands, func(i, j int) bool { return md.Commands[i].Name < md.Commands[j].Name })
	return md, nil
}
