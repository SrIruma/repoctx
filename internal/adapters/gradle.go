package adapters

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/SrIruma/repoctx/internal/project"
)

type gradleAdapter struct{}

func (gradleAdapter) Kind() project.ManifestKind { return project.KindGradle }
func (gradleAdapter) Language() string           { return "Gradle" }

// gradleConfigs are the dependency configurations we extract from.
var gradleConfigs = map[string]bool{
	"implementation":      true,
	"api":                 true,
	"compileOnly":         true,
	"runtimeOnly":         true,
	"testImplementation":  true,
	"annotationProcessor": true,
	"kapt":                true,
	"classpath":           true,
}

// groupNameVersionRe matches `group: 'g', name: 'a', version: 'v'`.
var groupNameVersionRe = regexp.MustCompile(`group:\s*["']([^"']+)["'].*name:\s*["']([^"']+)["'].*version:\s*["']([^"']+)["']`)

// gradleDependency extracts a dependency from one dependency declaration
// line. It supports the `implementation 'g:a:v'` shorthand, the paren form
// `implementation("g:a:v")` and the `group:…, name:…, version:…` map form.
// project(...) and files(...) references are ignored.
func gradleDependency(line string) (string, bool) {
	rest := strings.TrimSpace(line)
	i := strings.IndexAny(rest, " \t")
	if i < 0 {
		return "", false
	}
	if !gradleConfigs[rest[:i]] {
		return "", false
	}
	rest = strings.TrimSpace(rest[i+1:])
	if strings.HasPrefix(rest, "{") || strings.HasPrefix(rest, "project") || strings.HasPrefix(rest, "files") {
		return "", false
	}
	if m := groupNameVersionRe.FindStringSubmatch(rest); m != nil {
		return m[1] + ":" + m[2] + ":" + m[3], true
	}
	inner := rest
	if strings.HasPrefix(inner, "(") {
		if j := strings.IndexByte(inner, ')'); j > 0 {
			inner = inner[1:j]
		} else {
			return "", false
		}
	}
	inner = strings.TrimSpace(inner)
	if len(inner) < 2 {
		return "", false
	}
	q := inner[0]
	if q != '\'' && q != '"' {
		return "", false
	}
	if j := strings.IndexByte(inner[1:], q); j >= 0 {
		val := inner[1 : j+1]
		if strings.HasPrefix(val, "project") || strings.HasPrefix(val, "files") {
			return "", false
		}
		return val, true
	}
	return "", false
}

func (gradleAdapter) Read(path string) (*ManifestData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var deps []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "*") {
			continue
		}
		if d, ok := gradleDependency(line); ok {
			deps = append(deps, d)
		}
	}
	sort.Strings(deps)
	return &ManifestData{Deps: dedupSorted(deps)}, nil
}
