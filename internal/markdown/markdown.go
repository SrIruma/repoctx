// Package markdown provides marker-aware parsing and rendering of AI coding-agent
// context files (AGENTS.md, CLAUDE.md).
//
// repoctx-managed content lives between HTML comment markers:
//
//	<!-- repoctx:start -->
//	| Command | Source |
//	<!-- repoctx:end -->
//
// Everything outside the markers is treated as human-written and is never touched.
package markdown

import (
	"errors"
	"fmt"
	"strings"
)

const (
	StartMarker = "<!-- repoctx:start -->"
	EndMarker   = "<!-- repoctx:end -->"
)

// ErrNoMarkers is returned when a document has no repoctx marker pair.
var ErrNoMarkers = errors.New("no repoctx markers found; add " + StartMarker + " and " + EndMarker)

// Section is one marker-delimited region of a document, as byte offsets into
// the original content. Start points at the start marker, End just past the
// end marker.
type Section struct {
	Start int
	End   int
}

// FindSections returns every marker-delimited region in content, in document
// order. An unclosed start marker is an error.
func FindSections(content string) ([]Section, error) {
	var sections []Section
	rest := content
	base := 0
	for {
		start := strings.Index(rest, StartMarker)
		if start < 0 {
			return sections, nil
		}
		endRel := strings.Index(rest[start+len(StartMarker):], EndMarker)
		if endRel < 0 {
			return nil, fmt.Errorf("unclosed %s at offset %d", StartMarker, base+start)
		}
		sections = append(sections, Section{
			Start: base + start,
			End:   base + start + len(StartMarker) + endRel + len(EndMarker),
		})
		base = sections[len(sections)-1].End
		rest = content[base:]
	}
}

// HasMarkers reports whether content contains a complete marker pair.
func HasMarkers(content string) bool {
	sections, err := FindSections(content)
	return err == nil && len(sections) > 0
}

// CanonicalBlock wraps rendered content between the start and end markers.
func CanonicalBlock(rendered string) string {
	return StartMarker + "\n" + rendered + "\n" + EndMarker
}

// Update replaces every marker-delimited section with the canonical block
// wrapping rendered, preserving all human-written content verbatim. It returns
// ErrNoMarkers when content has no marker pair.
func Update(content, rendered string) (string, error) {
	sections, err := FindSections(content)
	if err != nil {
		return "", err
	}
	if len(sections) == 0 {
		return "", ErrNoMarkers
	}
	block := CanonicalBlock(strings.TrimSpace(rendered))
	var out strings.Builder
	last := 0
	for _, s := range sections {
		out.WriteString(content[last:s.Start])
		out.WriteString(block)
		last = s.End
	}
	out.WriteString(content[last:])
	return out.String(), nil
}
