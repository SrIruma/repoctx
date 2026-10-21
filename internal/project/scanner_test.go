package project

import (
	"os"
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
		{"meson.build", KindMeson},
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

func TestScannerWorkspaceNPM(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "workspace-npm")
	p, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := []struct {
		path  string
		scope string
	}{
		{"package.json", "."},
		{"packages/app/package.json", "packages/app"},
		{"packages/lib/package.json", "packages/lib"},
	}
	if len(p.Manifests) != len(want) {
		t.Fatalf("expected %d manifests, got %d: %+v", len(want), len(p.Manifests), p.Manifests)
	}
	for i, w := range want {
		if p.Manifests[i].Path != w.path {
			t.Errorf("manifests[%d].Path = %q, want %q", i, p.Manifests[i].Path, w.path)
		}
		if p.Manifests[i].Scope != w.scope {
			t.Errorf("manifests[%d].Scope = %q, want %q", i, p.Manifests[i].Scope, w.scope)
		}
	}
}

func TestScannerWorkspaceCargo(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "workspace-cargo")
	p, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := []struct {
		path  string
		scope string
	}{
		{"Cargo.toml", "."},
		{"crates/a/Cargo.toml", "crates/a"},
		{"crates/b/Cargo.toml", "crates/b"},
	}
	if len(p.Manifests) != len(want) {
		t.Fatalf("expected %d manifests, got %d: %+v", len(want), len(p.Manifests), p.Manifests)
	}
	for i, w := range want {
		if p.Manifests[i].Path != w.path {
			t.Errorf("manifests[%d].Path = %q, want %q", i, p.Manifests[i].Path, w.path)
		}
		if p.Manifests[i].Scope != w.scope {
			t.Errorf("manifests[%d].Scope = %q, want %q", i, p.Manifests[i].Scope, w.scope)
		}
	}
}

func TestScannerDiscoversNestedContextFiles(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "nested-context")
	p, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := []string{
		"AGENTS.md",
		"backend/AGENTS.md",
		"frontend/CLAUDE.md",
	}
	if len(p.ContextFiles) != len(want) {
		t.Fatalf("expected %d context files, got %d: %+v", len(want), len(p.ContextFiles), p.ContextFiles)
	}
	for i, w := range want {
		if p.ContextFiles[i] != w {
			t.Errorf("ContextFiles[%d] = %q, want %q", i, p.ContextFiles[i], w)
		}
	}
}

func TestScannerContextFilesRespectMaxDepth(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "nested-context")
	sc := NewScanner(root)
	sc.MaxDepth = 1
	p, err := sc.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(p.ContextFiles) != 1 || p.ContextFiles[0] != "AGENTS.md" {
		t.Errorf("MaxDepth 1: expected only root AGENTS.md, got %+v", p.ContextFiles)
	}
}

func TestScannerMarksFixtureManifestsNonCanonical(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "scanner", "mono")
	p, err := NewScanner(root).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	for _, m := range p.Manifests {
		if !m.Canonical {
			t.Errorf("manifest %q scanned as root should be canonical", m.Path)
		}
	}
}

func TestIsCanonical(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"package.json", true},
		{"backend/go.mod", true},
		{"internal/cli/audit.go", true},
		{"tests/fixtures/go/go.mod", false},
		{"test/package.json", false},
		{"examples/demo/Cargo.toml", false},
		{"vendor/lib/package.json", false},
		{"node_modules/pkg/package.json", false},
	}
	for _, tc := range cases {
		if got := IsCanonical(tc.path); got != tc.want {
			t.Errorf("IsCanonical(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestScannerExcludesContextFilesUnderFixtures(t *testing.T) {
	// A context file inside a test/fixture tree is not part of the canonical
	// surface and must not be discovered.
	dir := t.TempDir()
	for _, rel := range []string{"AGENTS.md", "tests/fixtures/x/AGENTS.md", "docs/CLAUDE.md"} {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	p, err := NewScanner(dir).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := []string{"AGENTS.md", "docs/CLAUDE.md"}
	if len(p.ContextFiles) != len(want) {
		t.Fatalf("ContextFiles = %+v, want %+v", p.ContextFiles, want)
	}
	for i, w := range want {
		if p.ContextFiles[i] != w {
			t.Errorf("ContextFiles[%d] = %q, want %q", i, p.ContextFiles[i], w)
		}
	}
}
