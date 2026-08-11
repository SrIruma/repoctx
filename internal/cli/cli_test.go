package cli

import (
	"bytes"
	"encoding/json"
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
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}); err != nil {
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
	if strings.Contains(got, "| stale |") {
		t.Errorf("stale table not replaced:\n%s", got)
	}
}

func TestGenerateCreatesMissingFile(t *testing.T) {
	dir := writeTestProject(t, "")
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runGenerate(cmd, dir, []string{"CLAUDE.md"}); err != nil {
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
	if err := runGenerate(cmd, dir, []string{"AGENTS.md"}); err == nil {
		t.Fatal("expected error for a file without markers")
	}
}

func TestAuditCLIHuman(t *testing.T) {
	dir := filepath.Join("..", "..", "tests", "fixtures", "audit", "ghost")
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runAudit(cmd, dir, false); err != nil {
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
	if err := runAudit(cmd, dir, true); err != nil {
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
