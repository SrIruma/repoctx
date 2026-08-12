package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestLoadMissingFileReturnsZeroValue(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load on empty dir: %v", err)
	}
	if cfg.SkipDirs != nil || cfg.MaxDepth != 0 || cfg.Files != nil {
		t.Errorf("expected zero-value Config, got %+v", cfg)
	}
}

func TestLoadParsesFields(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `max_depth = 3
skip_dirs = ["vendor", "third_party"]
files = ["AGENTS.md", "CLAUDE.md"]
`)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", cfg.MaxDepth)
	}
	if len(cfg.SkipDirs) != 2 || cfg.SkipDirs[0] != "vendor" || cfg.SkipDirs[1] != "third_party" {
		t.Errorf("SkipDirs = %v, want [vendor third_party]", cfg.SkipDirs)
	}
	if len(cfg.Files) != 2 || cfg.Files[0] != "AGENTS.md" || cfg.Files[1] != "CLAUDE.md" {
		t.Errorf("Files = %v, want [AGENTS.md CLAUDE.md]", cfg.Files)
	}
}

func TestLoadPartialFields(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "max_depth = 2\n")
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MaxDepth != 2 || cfg.SkipDirs != nil || cfg.Files != nil {
		t.Errorf("expected only MaxDepth set, got %+v", cfg)
	}
}

func TestLoadMalformedReturnsErrorWithPath(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "skip_dirs = [\"unterminated\n")
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for malformed TOML")
	}
	if !strings.Contains(err.Error(), FileName) {
		t.Errorf("error should mention %s, got %v", FileName, err)
	}
}

func TestLoadFileMissingReturnsError(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "nope.toml"))
	if err == nil {
		t.Fatal("expected error for missing explicit config path")
	}
}

func TestLoadFileParses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.toml")
	if err := os.WriteFile(path, []byte("max_depth = 4\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if cfg.MaxDepth != 4 {
		t.Errorf("MaxDepth = %d, want 4", cfg.MaxDepth)
	}
}
