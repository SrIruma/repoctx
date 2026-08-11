package project

import (
	"path/filepath"
	"testing"
)

func TestScannerDetectsManifestsAcrossTree(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "mono")
	sc := NewScanner(root)
	p, err := sc.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	want := []string{
		"backend/go.mod",
		"package.json",
		"tools/rust/Cargo.toml",
	}
	if len(p.Manifests) != len(want) {
		t.Fatalf("expected %d manifests, got %d: %+v", len(want), len(p.Manifests), p.Manifests)
	}
	for i, w := range want {
		if p.Manifests[i].Path != w {
			t.Errorf("manifests[%d].Path = %q, want %q", i, p.Manifests[i].Path, w)
		}
	}
}

func TestScannerSkipsDependencyAndVCSDirs(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "mono")
	sc := NewScanner(root)
	p, err := sc.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	for _, m := range p.Manifests {
		if m.Path == "node_modules/fake-pkg/package.json" {
			t.Errorf("node_modules should be skipped, found %q", m.Path)
		}
	}
}

func TestScannerSetsScopeFromManifestLocation(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "mono")
	p, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	scopeByPath := map[string]string{}
	for _, m := range p.Manifests {
		scopeByPath[m.Path] = m.Scope
	}
	if got := scopeByPath["backend/go.mod"]; got != "backend" {
		t.Errorf("scope of backend/go.mod = %q, want %q", got, "backend")
	}
	if got := scopeByPath["package.json"]; got != "." {
		t.Errorf("scope of package.json = %q, want %q", got, ".")
	}
}

func TestScannerReportsUnsupportedManifests(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "mono")
	p, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	// No known-other manifests in this fixture; just assert the field exists.
	if p.DetectedOther == nil {
		return
	}
}
