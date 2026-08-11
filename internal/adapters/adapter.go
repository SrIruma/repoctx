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
	}
	return nil, false
}
