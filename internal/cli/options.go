package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SrIruma/repoctx/internal/config"
)

// scanFlags carries the tunable flags shared by info, generate and audit that
// can also come from repoctx.toml. The *Set booleans mark flags the user
// passed explicitly, so config merging knows which ones win.
type scanFlags struct {
	maxDepth    int
	skipDirs    []string
	configPath  string
	files       []string
	maxDepthSet bool
	skipDirsSet bool
	filesSet    bool
}

// resolved holds the effective values after merging repoctx.toml with the
// explicitly-set flags. Precedence: flags > config > defaults. maxDepth 0
// means "scanner default" and is valid: the scanner falls back to its own
// default when non-positive.
type resolved struct {
	maxDepth int
	skipDirs []string
	files    []string
}

// resolve loads repoctx.toml (or the --config override) for dir and merges it
// with the flags the user set on the command line.
func (f scanFlags) resolve(dir string) (resolved, error) {
	var (
		cfg config.Config
		err error
	)
	if f.configPath != "" {
		cfg, err = config.LoadFile(f.configPath)
	} else {
		cfg, err = config.Load(dir)
	}
	if err != nil {
		return resolved{}, err
	}

	out := resolved{
		maxDepth: cfg.MaxDepth,
		skipDirs: cfg.SkipDirs,
	}
	if f.maxDepthSet {
		if f.maxDepth <= 0 {
			return resolved{}, fmt.Errorf("--max-depth must be a positive integer")
		}
		out.maxDepth = f.maxDepth
	}
	if f.skipDirsSet {
		out.skipDirs = f.skipDirs
	}

	out.files = cfg.Files
	if f.filesSet {
		out.files = f.files
	}
	if len(out.files) == 0 {
		out.files = []string{"AGENTS.md"}
	}
	return out, nil
}

// addScanFlags registers the flags that tune the scan on info, generate and
// audit. Values land in opts, which callers copy and mark as explicitly set
// via cmd.Flags().Changed.
func addScanFlags(cmd *cobra.Command, opts *scanFlags) {
	cmd.Flags().IntVar(&opts.maxDepth, "max-depth", 0,
		"maximum directory depth to scan (default 6)")
	cmd.Flags().StringSliceVar(&opts.skipDirs, "skip-dirs", nil,
		"directory name to skip, repeatable")
	cmd.Flags().StringVar(&opts.configPath, "config", "",
		"path to repoctx.toml (default: <dir>/repoctx.toml)")
}
