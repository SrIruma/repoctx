package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SrIruma/repoctx/internal/project"
)

type npmAdapter struct{}

func (npmAdapter) Kind() project.ManifestKind { return project.KindNPM }
func (npmAdapter) Language() string           { return "JavaScript/TypeScript" }

func (npmAdapter) Read(path string) (*ManifestData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pkg struct {
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		PackageManager  string            `json:"packageManager"`
	}
	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		return nil, err
	}

	pm := detectPackageManager(pkg.PackageManager, filepath.Dir(path))

	md := &ManifestData{PackageManager: pm}
	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		md.Commands = append(md.Commands, project.Command{Name: name, Cmd: pm + " run " + name})
	}

	seen := map[string]bool{}
	for _, d := range []map[string]string{pkg.Dependencies, pkg.DevDependencies} {
		for name := range d {
			if !seen[name] {
				seen[name] = true
				md.Deps = append(md.Deps, name)
			}
		}
	}
	sort.Strings(md.Deps)
	return md, nil
}

// detectPackageManager decides which tool runs a package.json's scripts.
// Precedence: the corepack "packageManager" field when it names a known tool,
// then a sibling lockfile, then npm as the default.
func detectPackageManager(packageManager, dir string) string {
	if pm, ok := packageManagerName(packageManager); ok {
		return pm
	}
	for _, lock := range [...]struct{ name, pm string }{
		{"yarn.lock", "yarn"},
		{"pnpm-lock.yaml", "pnpm"},
		{"bun.lock", "bun"},
		{"bun.lockb", "bun"},
		{"package-lock.json", "npm"},
		{"npm-shrinkwrap.json", "npm"},
	} {
		if _, err := os.Stat(filepath.Join(dir, lock.name)); err == nil {
			return lock.pm
		}
	}
	return "npm"
}

// packageManagerName extracts the tool name from a corepack "packageManager"
// field ("yarn@4.18.0", "pnpm@10.0.0", ...). It reports ok=false when the field
// is empty or names something repoctx does not know.
func packageManagerName(field string) (string, bool) {
	name, _, _ := strings.Cut(field, "@")
	switch name {
	case "npm", "yarn", "pnpm", "bun":
		return name, true
	}
	return "", false
}
