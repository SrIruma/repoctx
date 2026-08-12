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

func TestCargoAdapterWorkspaceRoot(t *testing.T) {
	ad := cargoAdapter{}
	md, err := ad.Read(fixture(t, "scanner/workspace-cargo/Cargo.toml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := map[string]string{
		"build":  "cargo build",
		"test":   "cargo test",
		"fmt":    "cargo fmt --check",
		"clippy": "cargo clippy",
	}
	for name, cmd := range want {
		if !hasCommand(md.Commands, name, cmd) {
			t.Errorf("missing cargo %s (%q) in workspace root: %v", name, cmd, commandNames(md.Commands))
		}
	}
	if hasCommand(md.Commands, "run", "cargo run") {
		t.Errorf("virtual workspace root must not expose cargo run: %v", md.Commands)
	}
	for _, dep := range []string{"serde", "anyhow"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing workspace dependency %q in %v", dep, md.Deps)
		}
	}
}

func TestCargoAdapterWorkspaceMember(t *testing.T) {
	ad := cargoAdapter{}
	md, err := ad.Read(fixture(t, "scanner/workspace-cargo/crates/a/Cargo.toml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !hasCommand(md.Commands, "run", "cargo run") {
		t.Errorf("workspace member crate should keep cargo run: %v", md.Commands)
	}
	if !contains(md.Deps, "serde") || !contains(md.Deps, "tokio") {
		t.Errorf("expected deps [serde tokio], got %v", md.Deps)
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

func TestCMakeAdapter(t *testing.T) {
	ad := cmakeAdapter{}
	md, err := ad.Read(fixture(t, "cmake/CMakeLists.txt"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, name := range []string{"build", "test"} {
		if !hasCommand(md.Commands, name, "cmake --build build --target "+name) {
			t.Errorf("missing cmake target %q: %v", name, commandNames(md.Commands))
		}
	}
	for _, dep := range []string{"Boost", "OpenSSL"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing find_package dep %q in %v", dep, md.Deps)
		}
	}
}

func TestGemfileAdapter(t *testing.T) {
	ad := gemfileAdapter{}
	md, err := ad.Read(fixture(t, "ruby/Gemfile"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, name := range []string{"rails", "puma", "rspec"} {
		if !hasCommand(md.Commands, name, "bundle exec "+name) {
			t.Errorf("missing bundle exec %q command: %v", name, commandNames(md.Commands))
		}
		if !contains(md.Deps, name) {
			t.Errorf("missing gem %q in %v", name, md.Deps)
		}
	}
}

func TestComposerAdapter(t *testing.T) {
	ad := composerAdapter{}
	md, err := ad.Read(fixture(t, "composer/composer.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !hasCommand(md.Commands, "test", "composer run test") {
		t.Errorf("missing composer run test: %v", md.Commands)
	}
	if !hasCommand(md.Commands, "lint", "composer run lint") {
		t.Errorf("missing composer run lint: %v", md.Commands)
	}
	for _, dep := range []string{"monolog/monolog", "phpunit/phpunit"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
}

func TestMavenAdapter(t *testing.T) {
	ad := mavenAdapter{}
	md, err := ad.Read(fixture(t, "maven/pom.xml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, dep := range []string{
		"org.springframework.boot:spring-boot-starter-web:3.2.0",
		"junit:junit",
	} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
	if len(md.Commands) != 0 {
		t.Errorf("pom.xml should expose no commands, got %v", commandNames(md.Commands))
	}
}

func TestGradleAdapter(t *testing.T) {
	ad := gradleAdapter{}
	md, err := ad.Read(fixture(t, "gradle/build.gradle"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, dep := range []string{
		"com.google.guava:guava:33.2.0-jre",
		"com.fasterxml.jackson.core:jackson-databind:2.17.0",
		"org.projectlombok:lombok:1.18.32",
		"org.junit.jupiter:junit-jupiter:5.10.2",
	} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
	if contains(md.Deps, "project(':lib')") || contains(md.Deps, ":lib") {
		t.Errorf("project references should be ignored: %v", md.Deps)
	}
}

func TestMesonAdapter(t *testing.T) {
	ad := mesonAdapter{}
	md, err := ad.Read(fixture(t, "meson/meson.build"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := map[string]string{
		"setup":   "meson setup build",
		"compile": "meson compile -C build",
		"test":    "meson test -C build",
	}
	for name, cmd := range want {
		if !hasCommand(md.Commands, name, cmd) {
			t.Errorf("missing meson %q command (%q): %v", name, cmd, md.Commands)
		}
	}
	for _, dep := range []string{"gtk+-3.0", "glib-2.0"} {
		if !contains(md.Deps, dep) {
			t.Errorf("missing dependency %q in %v", dep, md.Deps)
		}
	}
}
