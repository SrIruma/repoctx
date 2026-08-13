package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SrIruma/repoctx/internal/adapters"
	"github.com/SrIruma/repoctx/internal/project"
)

func fixtureDir(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "tests", "fixtures", "audit", name)
}

// extractActual runs the real scanner + adapters against a fixture so audit is
// tested end to end against the actual extraction pipeline.
func extractActual(t *testing.T, dir string) []project.Command {
	t.Helper()
	p, err := project.NewScanner(dir).Scan()
	if err != nil {
		t.Fatalf("scan %s: %v", dir, err)
	}
	var cmds []project.Command
	for _, m := range p.Manifests {
		ad, ok := adapters.For(m.Kind)
		if !ok {
			continue
		}
		md, err := ad.Read(filepath.Join(dir, m.Path))
		if err != nil {
			t.Fatalf("read %s: %v", m.Path, err)
		}
		cmds = append(cmds, md.Commands...)
	}
	return cmds
}

func checkByName(r *Report, name string) *Check {
	for i := range r.Checks {
		if r.Checks[i].Name == name {
			return &r.Checks[i]
		}
	}
	return nil
}

func TestAuditFlagsGhostCommands(t *testing.T) {
	dir := fixtureDir(t, "ghost")
	r, err := Run(Options{
		Root:   dir,
		File:   filepath.Join(dir, "AGENTS.md"),
		Actual: extractActual(t, dir),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	c := checkByName(r, "commands")
	if c == nil || c.Passed {
		t.Fatalf("expected commands check to fail, got %+v", c)
	}
	if len(c.Issues) != 1 || c.Issues[0].Command != "npm run lint" {
		t.Errorf("expected one ghost command 'npm run lint', got %+v", c.Issues)
	}
	if r.Score >= 100 || r.Passed {
		t.Errorf("expected score < 100 and failed report, got score=%d passed=%v", r.Score, r.Passed)
	}
}

func TestAuditFlagsStalePaths(t *testing.T) {
	dir := fixtureDir(t, "stale")
	r, err := Run(Options{
		Root:   dir,
		File:   filepath.Join(dir, "AGENTS.md"),
		Actual: extractActual(t, dir),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	c := checkByName(r, "paths")
	if c == nil || c.Passed {
		t.Fatalf("expected paths check to fail, got %+v", c)
	}
	if len(c.Issues) != 1 || c.Issues[0].Path != "legacy/schema.md" {
		t.Errorf("expected stale path 'legacy/schema.md', got %+v", c.Issues)
	}
	if cc := checkByName(r, "commands"); cc == nil || !cc.Passed {
		t.Errorf("commands check should pass, got %+v", cc)
	}
}

func TestAuditHealthy(t *testing.T) {
	dir := fixtureDir(t, "healthy")
	r, err := Run(Options{
		Root:   dir,
		File:   filepath.Join(dir, "AGENTS.md"),
		Actual: extractActual(t, dir),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !r.Passed || r.Score != 100 {
		t.Errorf("expected pass with score 100, got score=%d passed=%v\n%+v", r.Score, r.Passed, r.Checks)
	}
}

// TestAuditFlagsLiveRotInCLAUDE is the "live negative" test: a hand-maintained
// CLAUDE.md with real rot — a ghost command inside the markers and a stale
// path in human prose — must be flagged by the audit end to end. This is the
// scenario the repo's own dogfooding can never produce (its context files are
// regenerated right before the audit), so it lives here against a fixture.
func TestAuditFlagsLiveRotInCLAUDE(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "live")
	r, err := Run(Options{
		Root:   dir,
		File:   filepath.Join(dir, "CLAUDE.md"),
		Actual: extractActual(t, dir),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if r.Passed || r.Score >= 100 {
		t.Errorf("expected failed report, got score=%d passed=%v", r.Score, r.Passed)
	}
	cc := checkByName(r, "commands")
	if cc == nil || cc.Passed {
		t.Fatalf("expected commands check to fail, got %+v", cc)
	}
	if len(cc.Issues) != 1 || cc.Issues[0].Command != "npm run deploy" {
		t.Errorf("expected one ghost command 'npm run deploy', got %+v", cc.Issues)
	}
	pc := checkByName(r, "paths")
	if pc == nil || pc.Passed {
		t.Fatalf("expected paths check to fail, got %+v", pc)
	}
	if len(pc.Issues) != 1 || pc.Issues[0].Path != "docs/architecture-2024.md" {
		t.Errorf("expected stale path 'docs/architecture-2024.md', got %+v", pc.Issues)
	}
}

func TestAuditNoMarkersSkipsCommandsCheck(t *testing.T) {
	dir := fixtureDir(t, "healthy")
	r, err := Run(Options{
		Root:   dir,
		File:   filepath.Join(dir, "Makefile"),
		Actual: extractActual(t, dir),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	c := checkByName(r, "commands")
	if c == nil || !c.Passed {
		t.Fatalf("expected commands check to pass when no markers, got %+v", c)
	}
}

func TestCandidatePaths(t *testing.T) {
	content := "See `backend/schema.md` and `docs/README.md#install`. URL `https://example.com/x`.\n" +
		"Command `npm run test` and version `v0.0.1` should be ignored.\n" +
		"`CLAUDE.md` and `internal/cli.version` are conceptual, not paths."
	got := candidatePaths(content)
	want := []string{"backend/schema.md", "docs/README.md"}
	if len(got) != len(want) {
		t.Fatalf("candidatePaths = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("candidatePaths[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestIsPathLike(t *testing.T) {
	cases := map[string]bool{
		"backend/schema.md":    true,
		"package.json":         true,
		"Makefile":             true,
		"cmd/repoctx/":         true,
		"tests/fixtures/":      true,
		".gitignore":           true,
		"v0.0.1":               false,
		"go1.22":               false,
		"internal/cli.version": false,
		"docs/file.unknown":    false,
		"npm run test":         false,
		"https://x.io/a":       false,
		"#section":             false,
	}
	for in, want := range cases {
		if got := isPathLike(in); got != want {
			t.Errorf("isPathLike(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestAuditIgnoresConceptualReferences(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"name":"t","scripts":{"test":"jest"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	agents := "# Notes\n\n" +
		"Supported files are `AGENTS.md` and `CLAUDE.md`.\n" +
		"The version lives in `internal/cli.version`.\n\n" +
		"## Commands\n\n" +
		"<!-- repoctx:start -->\n" +
		"| Command | Source |\n|---|---|\n" +
		"| `npm run test` | `package.json` |\n" +
		"<!-- repoctx:end -->\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agents), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	r, err := Run(Options{
		Root:   dir,
		File:   filepath.Join(dir, "AGENTS.md"),
		Actual: extractActual(t, dir),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !r.Passed || r.Score != 100 {
		t.Errorf("expected pass with score 100, got score=%d passed=%v\n%+v", r.Score, r.Passed, r.Checks)
	}
}
