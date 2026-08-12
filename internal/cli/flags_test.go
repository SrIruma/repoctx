package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SrIruma/repoctx/internal/markdown"
)

const testPkg = `{"name":"t","scripts":{"test":"jest"}}`

func writeNestedProject(t *testing.T, dir string) {
	t.Helper()
	write := func(rel string) {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(testPkg), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("package.json")
	write("backend/go.mod")
	write("hidden/package.json")
	write("other/package.json")
}

func writeToml(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func execCLI(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	root := NewRootCmd()
	root.SetArgs(args)
	var out, errBuf bytes.Buffer
	code := execute(root, &out, &errBuf)
	return out.String(), errBuf.String(), code
}

// infoPaths runs `info <dir> --json` with extra args and returns the detected
// manifest paths, sorted as the scanner reports them.
func infoPaths(t *testing.T, dir string, extra ...string) []string {
	t.Helper()
	args := append([]string{"info", dir, "--json"}, extra...)
	out, errOut, code := execCLI(t, args...)
	if code != 0 {
		t.Fatalf("info exit code = %d, stderr:\n%s", code, errOut)
	}
	var p struct {
		Manifests []struct {
			Path string `json:"path"`
		} `json:"manifests"`
	}
	if err := json.Unmarshal([]byte(out), &p); err != nil {
		t.Fatalf("invalid info JSON: %v\n%s", err, out)
	}
	paths := make([]string, 0, len(p.Manifests))
	for _, m := range p.Manifests {
		paths = append(paths, m.Path)
	}
	return paths
}

func TestMaxDepthFlagLimitsScan(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)

	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"default depth scans everything", nil, []string{
			"backend/go.mod", "hidden/package.json", "other/package.json", "package.json",
		}},
		{"depth 1 root only", []string{"--max-depth", "1"}, []string{"package.json"}},
		{"depth 2 top-level packages", []string{"--max-depth", "2"}, []string{
			"backend/go.mod", "hidden/package.json", "other/package.json", "package.json",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := infoPaths(t, dir, tc.args...)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("manifests = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMaxDepthRejectsNonPositive(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	for _, v := range []string{"0", "-1"} {
		out, errOut, code := execCLI(t, "info", dir, "--max-depth", v)
		if code == 0 {
			t.Errorf("--max-depth %s: expected non-zero exit\nstdout:\n%s", v, out)
		}
		if !strings.Contains(errOut, "--max-depth") {
			t.Errorf("--max-depth %s: stderr should mention --max-depth, got:\n%s", v, errOut)
		}
	}
}

func TestSkipDirsFlagSkipsDirectory(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)

	got := infoPaths(t, dir, "--skip-dirs", "hidden")
	want := []string{"backend/go.mod", "other/package.json", "package.json"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("manifests = %v, want %v", got, want)
	}

	got = infoPaths(t, dir, "--skip-dirs", "hidden", "--skip-dirs", "other")
	want = []string{"backend/go.mod", "package.json"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("repeatable skip-dirs: manifests = %v, want %v", got, want)
	}
}

func TestSkipDirsPreservesBuiltinSkips(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	writeToml(t, filepath.Join(dir, "node_modules", "fake", "package.json"), testPkg)

	got := infoPaths(t, dir, "--skip-dirs", "hidden")
	for _, p := range got {
		if strings.HasPrefix(p, "node_modules/") {
			t.Errorf("builtin skip lost, found %q", p)
		}
	}
}

func TestGenerateMaxDepthLimitsTable(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	agents := "# P\n\n## Commands\n\n" + markdown.StartMarker + "\n| stale |\n" + markdown.EndMarker + "\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agents), 0o644); err != nil {
		t.Fatal(err)
	}

	_, errOut, code := execCLI(t, "generate", dir, "--max-depth", "1")
	if code != 0 {
		t.Fatalf("generate exit code = %d, stderr:\n%s", code, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "| `npm run test` | `package.json` |") {
		t.Errorf("root command missing:\n%s", got)
	}
	if strings.Contains(got, "go build ./...") || strings.Contains(got, "backend/go.mod") {
		t.Errorf("depth-limited scan leaked backend manifest:\n%s", got)
	}
}

func TestAuditMaxDepthTurnsCommandIntoGhost(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	agents := "# P\n\n## Commands\n\n" + markdown.StartMarker +
		"\n| `go test ./...` | `backend/go.mod` |\n" + markdown.EndMarker + "\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agents), 0o644); err != nil {
		t.Fatal(err)
	}

	_, errOut, code := execCLI(t, "audit", dir, "--check")
	if code != 0 {
		t.Errorf("default depth: expected pass, got code %d\nstderr:\n%s", code, errOut)
	}

	_, errOut, code = execCLI(t, "audit", dir, "--check", "--max-depth", "1")
	if code == 0 {
		t.Errorf("--max-depth 1: expected ghost-command failure, got success")
	}
}

func TestGenerateSkipDirsLimitsTable(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	agents := "# P\n\n## Commands\n\n" + markdown.StartMarker + "\n| stale |\n" + markdown.EndMarker + "\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agents), 0o644); err != nil {
		t.Fatal(err)
	}

	_, errOut, code := execCLI(t, "generate", dir, "--skip-dirs", "backend", "--skip-dirs", "other")
	if code != 0 {
		t.Fatalf("generate exit code = %d, stderr:\n%s", code, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, "backend/go.mod") || strings.Contains(got, "other/package.json") {
		t.Errorf("skipped manifests leaked into table:\n%s", got)
	}
	if !strings.Contains(got, "hidden/package.json") {
		t.Errorf("non-skipped manifest missing:\n%s", got)
	}
}

func TestAuditSkipDirsTurnsCommandIntoGhost(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	agents := "# P\n\n## Commands\n\n" + markdown.StartMarker +
		"\n| `go test ./...` | `backend/go.mod` |\n" + markdown.EndMarker + "\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agents), 0o644); err != nil {
		t.Fatal(err)
	}

	_, errOut, code := execCLI(t, "audit", dir, "--check")
	if code != 0 {
		t.Errorf("default: expected pass, got code %d\nstderr:\n%s", code, errOut)
	}

	_, errOut, code = execCLI(t, "audit", dir, "--check", "--skip-dirs", "backend")
	if code == 0 {
		t.Errorf("--skip-dirs backend: expected ghost-command failure, got success")
	}
}

func TestConfigPrecedenceFlagsOverConfig(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	writeToml(t, filepath.Join(dir, "repoctx.toml"),
		"max_depth = 1\nskip_dirs = [\"hidden\"]\n")

	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"config values apply", nil, []string{"package.json"}},
		{"flag max-depth beats config", []string{"--max-depth", "3"}, []string{
			"backend/go.mod", "other/package.json", "package.json",
		}},
		{"flag skip-dirs replaces config", []string{"--max-depth", "3", "--skip-dirs", "other"}, []string{
			"backend/go.mod", "hidden/package.json", "package.json",
		}},
		{"flag max-depth 0 still rejected", []string{"--max-depth", "0"}, []string{"error"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want[0] == "error" {
				out, errOut, code := execCLI(t, append([]string{"info", dir, "--json"}, tc.args...)...)
				if code == 0 {
					t.Errorf("expected error, got code 0\nstdout:\n%s", out)
				}
				if !strings.Contains(errOut, "--max-depth") {
					t.Errorf("stderr should mention --max-depth:\n%s", errOut)
				}
				return
			}
			got := infoPaths(t, dir, tc.args...)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("manifests = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerateUsesConfigFilesWhenFlagAbsent(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	writeToml(t, filepath.Join(dir, "repoctx.toml"), "files = [\"CLAUDE.md\"]\n")

	out, errOut, code := execCLI(t, "generate", dir)
	if code != 0 {
		t.Fatalf("generate exit code = %d, stderr:\n%s", code, errOut)
	}
	if !strings.Contains(out, "CLAUDE.md") {
		t.Errorf("generate should target CLAUDE.md from config, output:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md not created from config files")
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("AGENTS.md should not be created when config names CLAUDE.md")
	}
}

func TestGenerateFileFlagBeatsConfigFiles(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	writeToml(t, filepath.Join(dir, "repoctx.toml"), "files = [\"CLAUDE.md\"]\n")

	_, errOut, code := execCLI(t, "generate", dir, "--file", "AGENTS.md")
	if code != 0 {
		t.Fatalf("generate exit code = %d, stderr:\n%s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Errorf("AGENTS.md not created via --file")
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Errorf("CLAUDE.md should not be created when --file overrides config")
	}
}

func TestConfigFlagOverridesLocation(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	writeToml(t, filepath.Join(dir, "repoctx.toml"), "max_depth = 3\n")
	custom := filepath.Join(dir, "cfg", "custom.toml")
	writeToml(t, custom, "max_depth = 1\n")

	got := infoPaths(t, dir, "--config", custom)
	want := []string{"package.json"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("--config not applied: manifests = %v, want %v", got, want)
	}
}

func TestConfigInvalidPathErrors(t *testing.T) {
	dir := t.TempDir()
	writeNestedProject(t, dir)
	out, errOut, code := execCLI(t, "info", dir, "--config", filepath.Join(dir, "nope.toml"))
	if code == 0 {
		t.Errorf("expected error for missing --config path\nstdout:\n%s", out)
	}
	if !strings.Contains(errOut, "config") {
		t.Errorf("stderr should mention config:\n%s", errOut)
	}
}
