package markdown

import (
	"fmt"
	"strings"
)

// Row is one entry in the generated commands table: the runnable command and
// the manifest it was extracted from.
type Row struct {
	Command string
	Source  string
}

// RenderTable renders the commands table that is written between markers.
func RenderTable(rows []Row) string {
	var b strings.Builder
	b.WriteString("| Command | Source |\n")
	b.WriteString("|---|---|\n")
	if len(rows) == 0 {
		b.WriteString("| _no commands detected_ | |\n")
		return b.String()
	}
	for _, r := range rows {
		b.WriteString("| `" + r.Command + "` | `" + r.Source + "` |\n")
	}
	return b.String()
}

// ParseCommands reads the "| `cmd` | `source` |" rows back out of every marked
// section of content, skipping table headers and separators. It returns
// ErrNoMarkers when content has no marker pair.
func ParseCommands(content string) ([]Row, error) {
	sections, err := FindSections(content)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		return nil, ErrNoMarkers
	}
	var rows []Row
	for _, s := range sections {
		inner := content[s.Start+len(StartMarker) : s.End-len(EndMarker)]
		for _, line := range strings.Split(inner, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
				continue
			}
			cells := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|"), "|")
			if len(cells) < 2 {
				continue
			}
			cmd := strings.Trim(strings.TrimSpace(cells[0]), "`")
			src := strings.Trim(strings.TrimSpace(cells[1]), "`")
			if cmd == "" || cmd == "Command" || strings.Trim(cmd, "-") == "" || strings.HasPrefix(cmd, "_") {
				continue
			}
			rows = append(rows, Row{Command: cmd, Source: src})
		}
	}
	return rows, nil
}

// RowString formats a row for error messages and reports.
func RowString(r Row) string {
	if r.Source == "" {
		return fmt.Sprintf("%q", r.Command)
	}
	return fmt.Sprintf("%q (from %s)", r.Command, r.Source)
}
