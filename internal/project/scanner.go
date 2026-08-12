package project

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// defaultSkipDirs are directories never descended into: dependency trees,
// build outputs, VCS internals, editor junk.
var defaultSkipDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"vendor":        true,
	"target":        true,
	"dist":          true,
	"bin":           true,
	".venv":         true,
	"venv":          true,
	"__pycache__":   true,
	".next":         true,
	".cargo":        true,
	".idea":         true,
	".vscode":       true,
	".mypy_cache":   true,
	".pytest_cache": true,
	".tox":          true,
}

// knownOtherManifests are ecosystems we detect but whose adapter is not
// implemented yet. Detection degrades gracefully: they are reported, generic
// checks still apply, extraction simply has no adapter.
var knownOtherManifests = map[string]string{}

var manifestKinds = map[string]ManifestKind{
	"package.json":     KindNPM,
	"Cargo.toml":       KindCargo,
	"go.mod":           KindGo,
	"pyproject.toml":   KindPyProject,
	"Makefile":         KindMake,
	"makefile":         KindMake,
	"GNUmakefile":      KindMake,
	"CMakeLists.txt":   KindCMake,
	"Gemfile":          KindGemfile,
	"composer.json":    KindComposer,
	"pom.xml":          KindMaven,
	"build.gradle":     KindGradle,
	"build.gradle.kts": KindGradle,
	"meson.build":      KindMeson,
}

// Scanner walks a repository tree looking for project manifests.
type Scanner struct {
	Root     string
	MaxDepth int
	SkipDirs []string
}

func NewScanner(root string) *Scanner {
	return &Scanner{Root: root, MaxDepth: 6}
}

// Scan walks the tree and returns the detected Project.
func (s *Scanner) Scan() (*Project, error) {
	p := &Project{Root: s.Root}
	skip := make(map[string]bool, len(defaultSkipDirs)+len(s.SkipDirs))
	for d := range defaultSkipDirs {
		skip[d] = true
	}
	for _, d := range s.SkipDirs {
		skip[d] = true
	}
	maxDepth := s.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 6
	}

	err := filepath.WalkDir(s.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == s.Root {
				return nil
			}
			if skip[d.Name()] {
				return filepath.SkipDir
			}
			rel, rerr := filepath.Rel(s.Root, path)
			if rerr != nil {
				return rerr
			}
			depth := strings.Count(filepath.ToSlash(rel), "/") + 1
			if depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		rel, rerr := filepath.Rel(s.Root, path)
		if rerr != nil {
			return rerr
		}
		depth := strings.Count(filepath.ToSlash(rel), "/") + 1
		if depth > maxDepth {
			return nil
		}
		if kind, ok := manifestKinds[name]; ok {
			p.Manifests = append(p.Manifests, &Manifest{
				Path:  filepath.ToSlash(rel),
				Kind:  kind,
				Scope: ScopeOf(filepath.ToSlash(rel)),
			})
			return nil
		}
		if desc, ok := knownOtherManifests[name]; ok {
			p.DetectedOther = append(p.DetectedOther, desc+" ("+filepath.ToSlash(rel)+")")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(p.Manifests, func(i, j int) bool { return p.Manifests[i].Path < p.Manifests[j].Path })
	sort.Strings(p.DetectedOther)
	return p, nil
}
