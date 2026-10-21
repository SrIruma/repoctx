package cli

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name     string
		injected string
		info     *debug.BuildInfo
		want     string
	}{
		{
			name:     "injected value wins over build info",
			injected: "v0.7.0",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v0.6.0"}},
			want:     "v0.7.0",
		},
		{
			name:     "module version from go install",
			injected: "",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v0.7.0"}},
			want:     "v0.7.0",
		},
		{
			name:     "devel build reports dev",
			injected: "",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			want:     "dev",
		},
		{
			name:     "empty module version reports dev",
			injected: "",
			info:     &debug.BuildInfo{Main: debug.Module{Version: ""}},
			want:     "dev",
		},
		{
			name:     "missing build info reports dev",
			injected: "",
			info:     nil,
			want:     "dev",
		},
		{
			name:     "injected dev is reported verbatim",
			injected: "dev",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v0.7.0"}},
			want:     "dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveVersion(tt.injected, tt.info)
			if got != tt.want {
				t.Fatalf("resolveVersion(%q, ...) = %q, want %q", tt.injected, got, tt.want)
			}
		})
	}
}

func TestRootCommandReportsAVersion(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Version == "" {
		t.Fatal("root command has no version")
	}
	if cmd.Version == "0.0.1" {
		t.Fatal("root command falls back to the old hardcoded version")
	}
}
