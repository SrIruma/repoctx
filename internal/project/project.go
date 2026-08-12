package project

import "path/filepath"

// ManifestKind identifies the ecosystem a manifest belongs to.
type ManifestKind string

const (
	KindNPM       ManifestKind = "npm"
	KindCargo     ManifestKind = "cargo"
	KindGo        ManifestKind = "go"
	KindPyProject ManifestKind = "pyproject"
	KindMake      ManifestKind = "make"
	KindCMake     ManifestKind = "cmake"
	KindGemfile   ManifestKind = "gemfile"
	KindComposer  ManifestKind = "composer"
	KindMaven     ManifestKind = "maven"
	KindGradle    ManifestKind = "gradle"
)

// Command is a named project command (for example "test" -> "npm run test").
type Command struct {
	Name string `json:"name"`
	Cmd  string `json:"cmd"`
}

// Manifest is one detected project manifest and the facts extracted from it.
type Manifest struct {
	Path     string       `json:"path"`
	Kind     ManifestKind `json:"kind"`
	Language string       `json:"language"`
	Scope    string       `json:"scope"`
	Commands []Command    `json:"commands"`
	Deps     []string     `json:"deps,omitempty"`
	// Errors are per-manifest extraction failures. Best-effort extraction
	// keeps the manifest in the project with zero facts and records why.
	Errors []string `json:"errors,omitempty"`
}

// Project is the result of scanning a repository tree.
type Project struct {
	Root          string      `json:"root"`
	Manifests     []*Manifest `json:"manifests"`
	DetectedOther []string    `json:"detected_other,omitempty"`
}

// ScopeOf returns the scope directory for a manifest path (dir of the file).
func ScopeOf(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return "."
	}
	return filepath.ToSlash(dir)
}
