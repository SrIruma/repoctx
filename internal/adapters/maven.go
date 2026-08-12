package adapters

import (
	"encoding/xml"
	"os"
	"sort"

	"github.com/SrIruma/repoctx/internal/project"
)

type mavenAdapter struct{}

func (mavenAdapter) Kind() project.ManifestKind { return project.KindMaven }
func (mavenAdapter) Language() string           { return "Java" }

type pomProject struct {
	XMLName      xml.Name `xml:"project"`
	Dependencies []struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"dependencies>dependency"`
}

func (mavenAdapter) Read(path string) (*ManifestData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var p pomProject
	if err := xml.NewDecoder(f).Decode(&p); err != nil {
		return nil, err
	}

	var deps []string
	for _, d := range p.Dependencies {
		if d.GroupID == "" || d.ArtifactID == "" {
			continue
		}
		if d.Version == "" {
			deps = append(deps, d.GroupID+":"+d.ArtifactID)
		} else {
			deps = append(deps, d.GroupID+":"+d.ArtifactID+":"+d.Version)
		}
	}
	sort.Strings(deps)
	return &ManifestData{Deps: dedupSorted(deps)}, nil
}
