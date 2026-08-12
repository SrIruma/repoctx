package adapters

import (
	"github.com/SrIruma/repoctx/internal/project"
)

// ManifestData is the extraction result of an adapter.
type ManifestData struct {
	Commands []project.Command
	Deps     []string
}

// Adapter knows how to read one manifest format and turn it into facts.
type Adapter interface {
	Kind() project.ManifestKind
	Language() string
	Read(path string) (*ManifestData, error)
}

// For returns the adapter that handles the given manifest kind.
func For(k project.ManifestKind) (Adapter, bool) {
	switch k {
	case project.KindNPM:
		return npmAdapter{}, true
	case project.KindCargo:
		return cargoAdapter{}, true
	case project.KindGo:
		return goAdapter{}, true
	case project.KindPyProject:
		return pyprojectAdapter{}, true
	case project.KindMake:
		return makeAdapter{}, true
	case project.KindCMake:
		return cmakeAdapter{}, true
	case project.KindGemfile:
		return gemfileAdapter{}, true
	case project.KindComposer:
		return composerAdapter{}, true
	case project.KindMaven:
		return mavenAdapter{}, true
	case project.KindGradle:
		return gradleAdapter{}, true
	}
	return nil, false
}

// dedupSorted removes adjacent duplicates from a sorted slice.
func dedupSorted(in []string) []string {
	if len(in) < 2 {
		return in
	}
	out := in[:1]
	for _, s := range in[1:] {
		if s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return out
}
