package markdown

import (
	"fmt"
	"strings"
)

// Row is one entry in the generated commands table: the runnable command and
// the manifest it was extracted from.
//
// Line is the 1-based line number in the parsed document the row was read
// from. It is zero for rows produced programmatically (for example by an
// adapter), where no source line exists.
type Row struct {
	Command string
	Source  string
	Line    int
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
// section of content, skipping table headers, separators, and any other table
// (only two-cell rows are treated as command rows). It returns ErrNoMarkers
// when content has no marker pair.
//
// Each row carries the 1-based line number it was found on, so callers can
// tell which rows a given change touched.
func ParseCommands(content string) ([]Row, error) {
	sections, err := FindSections(content)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		return nil, ErrNoMarkers
	}
	// lineOf returns the 1-based document line for a byte offset.
	lineOf := func(off int) int {
		return strings.Count(content[:off], "\n") + 1
	}
	var rows []Row
	for _, s := range sections {
		innerStart := s.Start + len(StartMarker)
		inner := content[innerStart : s.End-len(EndMarker)]
		off := 0
		for _, line := range strings.Split(inner, "\n") {
			lineStart := innerStart + off
			off += len(line) + 1 // +1 for the newline the split consumed
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
				continue
			}
			cells := strings.Split(strings.TrimSuffix(strings.TrimPrefix(trimmed, "|"), "|"), "|")
			if len(cells) != 2 {
				continue
			}
			cmd := strings.Trim(strings.TrimSpace(cells[0]), "`")
			src := strings.Trim(strings.TrimSpace(cells[1]), "`")
			if cmd == "" || cmd == "Command" || strings.Trim(cmd, "-") == "" || strings.HasPrefix(cmd, "_") {
				continue
			}
			rows = append(rows, Row{Command: cmd, Source: src, Line: lineOf(lineStart)})
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

// ModuleRow is one entry in the generated modules table: a detected manifest,
// its language, and its dependencies.
type ModuleRow struct {
	Path     string
	Language string
	Deps     string
}

// RenderModules renders the modules table that follows the commands table.
func RenderModules(rows []ModuleRow) string {
	var b strings.Builder
	b.WriteString("| Module | Language | Dependencies |\n")
	b.WriteString("|---|---|---|\n")
	if len(rows) == 0 {
		b.WriteString("| _no modules detected_ | | |\n")
		return b.String()
	}
	for _, r := range rows {
		b.WriteString("| `" + r.Path + "` | " + r.Language + " | " + r.Deps + " |\n")
	}
	return b.String()
}

// RenderSection renders the canonical block content for a project: the
// commands table under a Commands header, then the modules table under a
// Modules header. ParseCommands skips everything but two-cell command rows,
// so the modules table round-trips without polluting the command list.
func RenderSection(cmdRows []Row, modRows []ModuleRow) string {
	var b strings.Builder
	b.WriteString("## Commands\n\n")
	b.WriteString(RenderTable(cmdRows))
	b.WriteString("\n## Modules\n\n")
	b.WriteString(RenderModules(modRows))
	return strings.TrimSuffix(b.String(), "\n")
}
