package audit

import (
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
		"Command `npm run test` and version `v0.0.1` should be ignored."
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
		"backend/schema.md": true,
		"package.json":      true,
		"Makefile":          true,
		"v0.0.1":            false,
		"go1.22":            false,
		"npm run test":      false,
		"https://x.io/a":    false,
		"#section":          false,
	}
	for in, want := range cases {
		if got := isPathLike(in); got != want {
			t.Errorf("isPathLike(%q) = %v, want %v", in, got, want)
		}
	}
}
