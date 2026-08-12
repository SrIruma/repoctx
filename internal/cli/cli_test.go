package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/audit"
	"github.com/SrIruma/repoctx/internal/markdown"
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
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}); err != nil {
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
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}); err != nil {
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
	if err := runGenerate(cmd, dir, []string{"CLAUDE.md"}, resolved{}); err != nil {
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
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}, resolved{}); err == nil {
		t.Fatal("expected error for a file without markers")
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
