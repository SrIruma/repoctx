package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSelfAGENTSKeepsCriticalFacts pins the semantic contract the repo's own
// context file must never silently lose. The CI dogfood gate (`generate` +
// `git diff`) only proves AGENTS.md matches what repoctx currently emits; it
// cannot tell a fact change caused by code churn from one caused by an adapter
// regression that silently drops a command. This test pins the facts that
// must survive regardless of adapter changes.
func TestSelfAGENTSKeepsCriticalFacts(t *testing.T) {
	path := filepath.Join("..", "..", "AGENTS.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	content := string(data)
	for _, want := range []string{
		"<!-- repoctx:start -->",
		"| `make test` |",
		"| `go test ./...` |",
		"| `go vet ./...` |",
		"| `go.mod` | Go |",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("AGENTS.md lost critical fact %q (adapter regression?)", want)
		}
	}
}
