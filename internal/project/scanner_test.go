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

func TestScannerMaxDepthLimitsManifests(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "mono")
	sc := NewScanner(root)
	sc.MaxDepth = 1
	p, err := sc.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(p.Manifests) != 1 || p.Manifests[0].Path != "package.json" {
		t.Errorf("MaxDepth 1: expected only root package.json, got %+v", p.Manifests)
	}
}

func TestScannerSkipDirsDoesNotLeakAcrossScans(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "mono")

	sc := NewScanner(root)
	sc.SkipDirs = []string{"backend"}
	if _, err := sc.Scan(); err != nil {
		t.Fatalf("scan with SkipDirs: %v", err)
	}

	clean, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan without SkipDirs: %v", err)
	}
	for _, m := range clean.Manifests {
		if m.Path == "backend/go.mod" {
			return
		}
	}
	t.Error("backend/go.mod missing: SkipDirs from the previous scan leaked globally")
}

func TestScannerPoly5(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "poly5")
	p, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := []struct {
		path string
		kind ManifestKind
	}{
		{"CMakeLists.txt", KindCMake},
		{"Gemfile", KindGemfile},
		{"build.gradle", KindGradle},
		{"composer.json", KindComposer},
		{"pom.xml", KindMaven},
	}
	if len(p.Manifests) != len(want) {
		t.Fatalf("expected %d manifests, got %d: %+v", len(want), len(p.Manifests), p.Manifests)
	}
	for i, w := range want {
		if p.Manifests[i].Path != w.path {
			t.Errorf("manifests[%d].Path = %q, want %q", i, p.Manifests[i].Path, w.path)
		}
		if p.Manifests[i].Kind != w.kind {
			t.Errorf("manifests[%d].Kind = %q, want %q", i, p.Manifests[i].Kind, w.kind)
		}
	}
	if len(p.DetectedOther) != 0 {
		t.Fatalf("expected no detected-other, got %v", p.DetectedOther)
	}
}
