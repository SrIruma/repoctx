package project

import "strings"

// nonCanonicalSegments are directory names whose contents are not considered
// part of the project's real surface: test trees, fixtures, examples, vendored
// code and build outputs. Manifests below them are detected and reported but
// excluded from the generated command/module tables, so a repository like
// repoctx (which ships dozens of fixture manifests) does not drown its own
// context file in noise.
var nonCanonicalSegments = map[string]bool{
	"test":         true,
	"tests":        true,
	"testdata":     true,
	"fixture":      true,
	"fixtures":     true,
	"example":      true,
	"examples":     true,
	"vendor":       true,
	"third_party":  true,
	"node_modules": true,
	"__tests__":    true,
	"e2e":          true,
	"spec":         true,
	"specs":        true,
}

// IsCanonical reports whether a path (relative, POSIX or native separators)
// belongs to the canonical project surface. A path is non-canonical when any
// of its directory segments is a recognised test/fixture/example/vendor name.
func IsCanonical(path string) bool {
	for _, seg := range strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '\\' }) {
		if nonCanonicalSegments[seg] {
			return false
		}
	}
	return true
}
