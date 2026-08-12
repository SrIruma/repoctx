package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/audit"
	"github.com/SrIruma/repoctx/internal/markdown"
	"github.com/SrIruma/repoctx/internal/project"
)

func writeTestProject(t *testing.T, agentsContent string) string {
	t.Helper()
	dir := t.TempDir()
	pkg := `{"name":"t","scripts":{"test":"jest"},"devDependencies":{"typescript":"^5.0.0"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if agentsContent != "" {
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agentsContent), 0o644); err != nil {
			t.Fatalf("write AGENTS.md: %v", err)
		}
	}
	return dir
}

func TestGenerateRoundTrip(t *testing.T) {
	dir := writeTestProject(t, "# My Project\n\nHuman notes.\n\n## Commands\n\n"+
		markdown.StartMarker+"\n| stale |\n"+markdown.EndMarker+"\n")
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, false); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	got := string(data)
	for _, keep := range []string{"# My Project", "Human notes."} {
		if !strings.Contains(got, keep) {
			t.Errorf("human content %q lost:\n%s", keep, got)
		}
	}
	if !strings.Contains(got, "| `npm run test` | `package.json` |") {
		t.Errorf("table not updated:\n%s", got)
	}
	if !strings.Contains(got, "| `package.json` | JavaScript/TypeScript | typescript |") {
		t.Errorf("modules table not rendered:\n%s", got)
	}
	if strings.Contains(got, "| stale |") {
		t.Errorf("stale table not replaced:\n%s", got)
	}
}

func TestGenerateModulesRoundTripKeepsCommandsOnly(t *testing.T) {
	dir := writeTestProject(t, "# P\n\n## Commands\n\n"+
		markdown.StartMarker+"\n| stale |\n"+markdown.EndMarker+"\n")
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, false); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	rows, err := markdown.ParseCommands(string(data))
	if err != nil {
		t.Fatalf("ParseCommands: %v", err)
	}
	if len(rows) != 1 || rows[0].Command != "npm run test" || rows[0].Source != "package.json" {
		t.Errorf("expected exactly the npm command row, got %+v", rows)
	}
}

func TestGenerateCreatesMissingFile(t *testing.T) {
	dir := writeTestProject(t, "")
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runGenerate(cmd, dir, []string{"CLAUDE.md"}, resolved{}, false); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md was not created: %v", err)
	}
	if !strings.Contains(string(data), markdown.StartMarker) {
		t.Errorf("CLAUDE.md missing markers:\n%s", data)
	}
}

func TestGenerateFailsWithoutMarkers(t *testing.T) {
	dir := writeTestProject(t, "# no markers here")
	cmd := &cobra.Command{}
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, false); err == nil {
		t.Fatal("expected error for a file without markers")
	}
}

func TestGenerateDryRunReportsUpdateWithoutWriting(t *testing.T) {
	dir := writeTestProject(t, "# My Project\n\nHuman notes.\n\n## Commands\n\n"+
		markdown.StartMarker+"\n| stale |\n"+markdown.EndMarker+"\n")
	path := filepath.Join(dir, "AGENTS.md")
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatalf("set mtime: %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, true); err != nil {
		t.Fatalf("runGenerate dry-run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "would update AGENTS.md") {
		t.Errorf("dry-run should report what would change, got:\n%s", out)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("dry-run wrote the file:\n%s", after)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.ModTime().Equal(old) {
		t.Errorf("mtime changed: got %v, want %v", fi.ModTime(), old)
	}
}

func TestGenerateDryRunUpToDate(t *testing.T) {
	dir := writeTestProject(t, "# P\n\n## Commands\n\n"+
		markdown.StartMarker+"\n| stale |\n"+markdown.EndMarker+"\n")
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, false); err != nil {
		t.Fatalf("generate: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}

	cmd = &cobra.Command{}
	buf.Reset()
	cmd.SetOut(&buf)
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, true); err != nil {
		t.Fatalf("runGenerate dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "is up to date") {
		t.Errorf("dry-run should report up-to-date file, got:\n%s", buf.String())
	}
	after, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, after) {
		t.Errorf("dry-run modified an up-to-date file")
	}
}

func TestGenerateDryRunVersusRealRun(t *testing.T) {
	dir := writeTestProject(t, "# P\n\n## Commands\n\n"+
		markdown.StartMarker+"\n| stale |\n"+markdown.EndMarker+"\n")
	path := filepath.Join(dir, "AGENTS.md")

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, true); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "| stale |") {
		t.Errorf("dry-run wrote content; file changed:\n%s", data)
	}

	cmd = &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}, false); err != nil {
		t.Fatalf("real run: %v", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "| stale |") {
		t.Errorf("real run did not replace stale content:\n%s", data)
	}
}

func TestLoadProjectKeepsManifestErrors(t *testing.T) {
	dir := writeTestProject(t, "")
	broken := filepath.Join(dir, "broken")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "package.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := loadProject(dir, resolved{})
	if err != nil {
		t.Fatalf("loadProject: %v", err)
	}
	if len(p.Manifests) != 2 {
		t.Fatalf("expected both manifests despite one being corrupt, got %+v", p.Manifests)
	}
	byPath := map[string]*project.Manifest{}
	for _, m := range p.Manifests {
		byPath[m.Path] = m
	}
	if len(byPath["package.json"].Errors) != 0 {
		t.Errorf("healthy manifest should have no errors, got %v", byPath["package.json"].Errors)
	}
	if len(byPath["package.json"].Commands) == 0 {
		t.Errorf("healthy manifest should keep its facts")
	}
	if len(byPath["broken/package.json"].Errors) == 0 {
		t.Errorf("corrupt manifest should record an extraction error")
	}
	if len(byPath["broken/package.json"].Commands) != 0 {
		t.Errorf("corrupt manifest should have no commands")
	}
}

func TestInfoJSONSurfacesManifestErrors(t *testing.T) {
	dir := writeTestProject(t, "")
	broken := filepath.Join(dir, "broken")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "package.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, errOut, code := execCLI(t, "info", dir, "--json")
	if code != 0 {
		t.Fatalf("info exit code = %d, stderr:\n%s", code, errOut)
	}
	var p struct {
		Manifests []struct {
			Path   string   `json:"path"`
			Errors []string `json:"errors"`
		} `json:"manifests"`
	}
	if err := json.Unmarshal([]byte(out), &p); err != nil {
		t.Fatalf("invalid info JSON: %v\n%s", err, out)
	}
	var got []string
	for _, m := range p.Manifests {
		if m.Path == "broken/package.json" {
			got = m.Errors
		}
	}
	if len(got) == 0 {
		t.Errorf("info --json should list the extraction error for broken/package.json, got manifests %+v", p.Manifests)
	}
}

func TestInfoHumanWarnsOnManifestErrors(t *testing.T) {
	dir := writeTestProject(t, "")
	broken := filepath.Join(dir, "broken")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "package.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, errOut, code := execCLI(t, "info", dir)
	if code != 0 {
		t.Fatalf("info exit code = %d, stderr:\n%s", code, errOut)
	}
	if !strings.Contains(out, "! broken/package.json") {
		t.Errorf("human output should warn about the failed manifest, got:\n%s", out)
	}
}

func TestAuditCLIHuman(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "audit", "ghost")
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runAudit(cmd, dir, false, false, resolved{}); err != nil {
		t.Fatalf("runAudit: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "FAIL") || !strings.Contains(out, "npm run lint") {
		t.Errorf("expected FAIL report with ghost command, got:\n%s", out)
	}
}

func TestAuditCLIJSON(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "audit", "healthy")
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runAudit(cmd, dir, true, false, resolved{}); err != nil {
		t.Fatalf("runAudit: %v", err)
	}
	var reports []audit.Report
	if err := json.Unmarshal(buf.Bytes(), &reports); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if len(reports) != 1 || reports[0].Score != 100 || !reports[0].Passed {
		t.Errorf("expected one passing 100-score report, got %+v", reports)
	}
}

func TestAuditCLICheckFails(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "audit", "ghost")
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	err := runAudit(cmd, dir, false, true, resolved{})
	var ee *exitError
	if !errors.As(err, &ee) || ee.code != 1 {
		t.Fatalf("expected *exitError{code:1}, got %v", err)
	}
}

func TestAuditCLICheckPasses(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "audit", "healthy")
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runAudit(cmd, dir, false, true, resolved{}); err != nil {
		t.Fatalf("expected nil error on healthy fixture, got %v", err)
	}
}

func TestAuditCLICheckJSON(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "audit", "ghost")
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := runAudit(cmd, dir, true, true, resolved{})
	var ee *exitError
	if !errors.As(err, &ee) || ee.code != 1 {
		t.Fatalf("expected *exitError{code:1}, got %v", err)
	}
	var reports []audit.Report
	if err := json.Unmarshal(buf.Bytes(), &reports); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if reports[0].Passed {
		t.Errorf("expected failing report in JSON output, got %+v", reports[0])
	}
}

func TestExecuteAuditCheckExitCode(t *testing.T) {
	for _, tc := range []struct {
		name string
		dir  string
		want int
	}{
		{"failing fixture exits 1", filepath.Join("..", "..", "tests", "fixtures", "audit", "ghost"), 1},
		{"healthy fixture exits 0", filepath.Join("..", "..", "tests", "fixtures", "audit", "healthy"), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCmd()
			root.SetArgs([]string{"audit", tc.dir, "--check"})
			var out, errBuf bytes.Buffer
			if got := execute(root, &out, &errBuf); got != tc.want {
				t.Errorf("execute exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", got, tc.want, out.String(), errBuf.String())
			}
		})
	}
}

func TestWorkflowTemplateDefault(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runWorkflow(cmd, nil); err != nil {
		t.Fatalf("runWorkflow: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"## Repo Context Maintenance",
		"<!-- repoctx:start -->",
		"<!-- repoctx:end -->",
		"`AGENTS.md`",
		"`repoctx generate .`",
		"`repoctx audit . --check`",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("template missing %q:\n%s", want, out)
		}
	}
}

func TestWorkflowTemplateMultipleFiles(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runWorkflow(cmd, []string{"AGENTS.md", "CLAUDE.md"}); err != nil {
		t.Fatalf("runWorkflow: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"`AGENTS.md` and `CLAUDE.md`", "--file AGENTS.md --file CLAUDE.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("template missing %q:\n%s", want, out)
		}
	}
}

// Workspace policy: commands are attributed per manifest. The same command
// string from different manifests (e.g. `npm run test` at the workspace root
// and in each package) is intentional and disambiguated by the Source column.
// The rendered Commands table never contains the same (command, source) row
// twice. Scopes are asserted at the scanner level; this test pins the
// rendered output of a typical npm/pnpm monorepo.
func TestWorkspaceNPMCommandsPolicy(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "scanner", "workspace-npm")
	p, err := loadProject(dir, resolved{})
	if err != nil {
		t.Fatalf("loadProject: %v", err)
	}
	rows := tableRows(p)

	want := []markdown.Row{
		{Command: "npm run build", Source: "package.json"},
		{Command: "npm run test", Source: "package.json"},
		{Command: "npm run test", Source: "packages/app/package.json"},
		{Command: "npm run test", Source: "packages/lib/package.json"},
	}
	if len(rows) != len(want) {
		t.Fatalf("expected %d command rows, got %d: %+v", len(want), len(rows), rows)
	}
	for _, w := range want {
		if !containsRow(rows, w) {
			t.Errorf("missing command row %+v in %+v", w, rows)
		}
	}
	if dup := duplicateRows(rows); len(dup) > 0 {
		t.Errorf("duplicate command rows: %+v", dup)
	}
}

func containsRow(rows []markdown.Row, want markdown.Row) bool {
	for _, r := range rows {
		if r == want {
			return true
		}
	}
	return false
}

func duplicateRows(rows []markdown.Row) []markdown.Row {
	seen := map[markdown.Row]int{}
	var dup []markdown.Row
	for _, r := range rows {
		if seen[r] == 1 {
			dup = append(dup, r)
		}
		seen[r]++
	}
	return dup
}
