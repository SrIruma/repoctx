package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseDiffLines(t *testing.T) {
	diff := "diff --git a/AGENTS.md b/AGENTS.md\n" +
		"index 1..2 100644\n" +
		"--- a/AGENTS.md\n" +
		"+++ b/AGENTS.md\n" +
		"@@ -3,0 +4,2 @@\n" +
		"+| `npm run lint` | `package.json` |\n" +
		"+| `npm run build` | `package.json` |\n" +
		"@@ -10 +12 @@\n" +
		"-old\n" +
		"+new\n"
	got := parseDiffLines(diff)
	for _, n := range []int{4, 5, 12} {
		if !got[n] {
			t.Errorf("expected line %d in changed set, got %v", n, got)
		}
	}
	if got[3] || got[11] {
		t.Errorf("unexpected lines in changed set: %v", got)
	}
}

// gitOrSkip skips the test when git is unavailable.
func gitOrSkip(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func runGitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestChangedLinesIncremental drives changedLines against a real repository:
// a committed context file, then an edit, and checks only the touched lines
// are flagged.
func TestChangedLinesIncremental(t *testing.T) {
	gitOrSkip(t)
	dir := t.TempDir()
	runGitIn(t, dir, "init")
	runGitIn(t, dir, "config", "user.email", "t@t")
	runGitIn(t, dir, "config", "user.name", "t")

	file := filepath.Join(dir, "AGENTS.md")
	base := "# Context\n\nline 3\nline 4\n"
	if err := os.WriteFile(file, []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitIn(t, dir, "add", "AGENTS.md")
	runGitIn(t, dir, "commit", "-m", "base")

	// Change only line 4.
	if err := os.WriteFile(file, []byte("# Context\n\nline 3\nCHANGED\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := changedLines(dir, file, "HEAD")
	if err != nil {
		t.Fatalf("changedLines: %v", err)
	}
	if !got[4] {
		t.Errorf("expected line 4 changed, got %v", got)
	}
	if got[3] {
		t.Errorf("line 3 unchanged but flagged: %v", got)
	}
}

// TestChangedLinesNewFile: a file that does not exist at the base is entirely
// new, so every line counts.
func TestChangedLinesNewFile(t *testing.T) {
	gitOrSkip(t)
	dir := t.TempDir()
	runGitIn(t, dir, "init")
	runGitIn(t, dir, "config", "user.email", "t@t")
	runGitIn(t, dir, "config", "user.name", "t")
	other := filepath.Join(dir, "keep.txt")
	if err := os.WriteFile(other, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitIn(t, dir, "add", "keep.txt")
	runGitIn(t, dir, "commit", "-m", "base")

	file := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(file, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := changedLines(dir, file, "HEAD")
	if err != nil {
		t.Fatalf("changedLines: %v", err)
	}
	for _, n := range []int{1, 2, 3} {
		if !got[n] {
			t.Errorf("new file: expected line %d changed, got %v", n, got)
		}
	}
}

func TestChangedLinesBadRev(t *testing.T) {
	gitOrSkip(t)
	dir := t.TempDir()
	runGitIn(t, dir, "init")
	if _, err := changedLines(dir, filepath.Join(dir, "AGENTS.md"), "no-such-rev"); err == nil {
		t.Error("expected error for unresolvable revision")
	}
}
