package cli

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateGolden regenerates the snapshots in testdata/ instead of comparing:
//
//	go test ./internal/cli -run Golden -update
var updateGolden = flag.Bool("update", false, "update golden files in testdata/")

// TestGoldenJSON pins the machine-readable contract of `info --json` and
// `audit --json` against committed snapshots (see issue #18 and #19). The
// snapshots are the documented schemas in docs/contract.md: any change to the
// JSON shape that is not deliberate (and snapshot-updated) fails here.
//
// Absolute fixture roots are normalized to the literal `<repo>` placeholder so
// the snapshots are stable across machines and checkouts. The schemas
// themselves are validated on the normalized output.
func TestGoldenJSON(t *testing.T) {
	repo := filepath.Join("..", "..")
	cases := []struct {
		name string
		args []string
	}{
		{
			name: "info_mono",
			args: []string{"info", filepath.Join(repo, "tests", "fixtures", "scanner", "mono"), "--json"},
		},
		{
			name: "info_corrupt",
			args: []string{"info", filepath.Join(repo, "tests", "fixtures", "scanner", "corrupt"), "--json"},
		},
		{
			name: "info_yarn",
			args: []string{"info", filepath.Join(repo, "tests", "fixtures", "yarn"), "--json"},
		},
		{
			name: "audit_healthy",
			args: []string{"audit", filepath.Join(repo, "tests", "fixtures", "audit", "healthy"), "--json"},
		},
		{
			name: "audit_stale",
			args: []string{"audit", filepath.Join(repo, "tests", "fixtures", "audit", "stale"), "--json"},
		},
		{
			name: "audit_ghost",
			args: []string{"audit", filepath.Join(repo, "tests", "fixtures", "audit", "ghost"), "--json"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, errOut, code := execCLI(t, tc.args...)
			if code != 0 {
				t.Fatalf("exit code = %d, stderr:\n%s", code, errOut)
			}
			dir := filepath.Join(repo, "tests", "fixtures")
			got := normalizeGolden(out, dir)
			golden := filepath.Join("testdata", tc.name+".json")
			if *updateGolden {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden %s: %v (regenerate with -update)", golden, err)
			}
			if got != string(want) {
				t.Errorf("golden %s mismatch:\n--- got ---\n%s\n--- want ---\n%s\n(regenerate with -update)", golden, got, want)
			}
		})
	}
}

// normalizeGolden replaces the absolute fixture-root prefix with the literal
// placeholder `<repo>` so snapshots are portable. The JSON contract (root,
// file, path fields) is otherwise left untouched.
func normalizeGolden(out, root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return out
	}
	out = strings.ReplaceAll(out, filepath.ToSlash(abs), "<repo>")
	out = strings.ReplaceAll(out, abs, "<repo>")
	return out
}
