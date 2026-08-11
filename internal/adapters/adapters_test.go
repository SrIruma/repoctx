package adapters

import (
	"path/filepath"
	"testing"

	"github.com/SrIruma/repoctx/internal/project"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "tests", "fixtures", name)
}

func commandNames(cmds []project.Command) []string {
	out := make([]string, 0, len(cmds))
	for _, c := range cmds {
		out = append(out, c.Name)
	}
	return out
}

func hasCommand(cmds []project.Command, name, cmd string) bool {
	for _, c := range cmds {
		if c.Name == name && c.Cmd == cmd {
			return true
		}
	}
	return false
}

func TestNPMAdapter(t *testing.T) {
	ad := npmAdapter{}
	md, err := ad.Read(fixture(t, "npm/package.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(md.Commands) != 3 {
		t.Errorf("expected 3 commands, got %v", commandNames(md.Commands))
	}
	if !hasCommand(md.Commands, "test", "npm run test") {
		t.Errorf("missing npm run test command: %v", md.Commands)
	}
	for _, dep := range []string{"react", "typescript"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
}

func TestCargoAdapter(t *testing.T) {
	ad := cargoAdapter{}
	md, err := ad.Read(fixture(t, "cargo/Cargo.toml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !hasCommand(md.Commands, "build", "cargo build") {
		t.Errorf("missing cargo build: %v", md.Commands)
	}
	if !hasCommand(md.Commands, "test", "cargo test") {
		t.Errorf("missing cargo test: %v", md.Commands)
	}
	for _, dep := range []string{"serde", "anyhow", "criterion"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
}

func TestGoAdapter(t *testing.T) {
	ad := goAdapter{}
	md, err := ad.Read(fixture(t, "go/go.mod"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !hasCommand(md.Commands, "test", "go test ./...") {
		t.Errorf("missing go test command: %v", md.Commands)
	}
	for _, dep := range []string{"github.com/spf13/cobra", "github.com/stretchr/testify"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
	if contains(md.Deps, "example.com/mock") {
		t.Errorf("module header should not be reported as a dependency: %v", md.Deps)
	}
}

func TestPyProjectAdapter(t *testing.T) {
	ad := pyprojectAdapter{}
	md, err := ad.Read(fixture(t, "pyproject/pyproject.toml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !hasCommand(md.Commands, "test", "pytest") {
		t.Errorf("missing pytest command: %v", md.Commands)
	}
	if !hasCommand(md.Commands, "lint", "ruff check .") {
		t.Errorf("missing ruff lint command: %v", md.Commands)
	}
	for _, dep := range []string{"requests", "click"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
}

func TestMakeAdapter(t *testing.T) {
	ad := makeAdapter{}
	md, err := ad.Read(fixture(t, "make/Makefile"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, name := range []string{"build", "test", "clean"} {
		if !hasCommand(md.Commands, name, "make "+name) {
			t.Errorf("missing make target %q: %v", name, commandNames(md.Commands))
		}
	}
	for _, bad := range []string{".PHONY", "%.o"} {
		if contains(commandNames(md.Commands), bad) {
			t.Errorf("target %q should have been excluded: %v", bad, commandNames(md.Commands))
		}
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
