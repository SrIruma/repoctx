package markdown

import (
	"errors"
	"strings"
	"testing"
)

func TestRenderTable(t *testing.T) {
	rows := []Row{
		{Command: "npm run test", Source: "package.json"},
		{Command: "make build", Source: "Makefile"},
	}
	want := "| Command | Source |\n|---|---|\n| `npm run test` | `package.json` |\n| `make build` | `Makefile` |\n"
	if got := RenderTable(rows); got != want {
		t.Errorf("RenderTable:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderTableEmpty(t *testing.T) {
	if got := RenderTable(nil); !strings.Contains(got, "_no commands detected_") {
		t.Errorf("expected empty-state row, got:\n%s", got)
	}
}

func TestFindSections(t *testing.T) {
	doc := "Intro\n" + StartMarker + "\nold\n" + EndMarker + "\nOutro\n"
	sections, err := FindSections(doc)
	if err != nil {
		t.Fatalf("FindSections: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	inner := doc[sections[0].Start+len(StartMarker) : sections[0].End-len(EndMarker)]
	if strings.TrimSpace(inner) != "old" {
		t.Errorf("inner content = %q, want %q", inner, "old")
	}
}

func TestFindSectionsNone(t *testing.T) {
	if sections, err := FindSections("plain text"); err != nil || len(sections) != 0 {
		t.Errorf("expected no sections and no error, got %d sections, err=%v", len(sections), err)
	}
}

func TestFindSectionsUnclosed(t *testing.T) {
	if _, err := FindSections("x\n" + StartMarker + "\nnever closed"); err == nil {
		t.Error("expected error for unclosed marker")
	}
}

func TestFindSectionsMultiple(t *testing.T) {
	doc := StartMarker + "\na\n" + EndMarker + "\nmiddle\n" + StartMarker + "\nb\n" + EndMarker
	sections, err := FindSections(doc)
	if err != nil {
		t.Fatalf("FindSections: %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
}

func TestUpdatePreservesHumanContent(t *testing.T) {
	doc := "# Project\n\nSome human notes.\n\n" + StartMarker + "\nstale table\n" + EndMarker + "\n\nTail."
	got, err := Update(doc, RenderTable([]Row{{Command: "go test ./...", Source: "go.mod"}}))
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	for _, keep := range []string{"# Project", "Some human notes.", "Tail."} {
		if !strings.Contains(got, keep) {
			t.Errorf("human content %q lost after update:\n%s", keep, got)
		}
	}
	if strings.Contains(got, "stale table") {
		t.Errorf("stale section was not replaced:\n%s", got)
	}
	if !strings.Contains(got, "| `go test ./...` | `go.mod` |") {
		t.Errorf("new table missing after update:\n%s", got)
	}
}

func TestUpdateMultipleSections(t *testing.T) {
	doc := StartMarker + "\nA\n" + EndMarker + "\nmid\n" + StartMarker + "\nB\n" + EndMarker
	got, err := Update(doc, RenderTable([]Row{{Command: "make test", Source: "Makefile"}}))
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if strings.Count(got, StartMarker) != 2 || strings.Count(got, "| `make test` |") != 2 {
		t.Errorf("expected both sections updated:\n%s", got)
	}
	if !strings.Contains(got, "\nmid\n") {
		t.Errorf("middle content lost:\n%s", got)
	}
}

func TestUpdateNoMarkers(t *testing.T) {
	if _, err := Update("no markers here", "x"); !errors.Is(err, ErrNoMarkers) {
		t.Errorf("expected ErrNoMarkers, got %v", err)
	}
}

func TestParseCommands(t *testing.T) {
	doc := "intro\n" + StartMarker + "\n" +
		"| Command | Source |\n|---|---|\n" +
		"| `npm run test` | `package.json` |\n" +
		"| `make build` | `Makefile` |\n" +
		EndMarker + "\noutro"
	rows, err := ParseCommands(doc)
	if err != nil {
		t.Fatalf("ParseCommands: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %v", rows)
	}
	if rows[0].Command != "npm run test" || rows[0].Source != "package.json" {
		t.Errorf("rows[0] = %+v", rows[0])
	}
	if rows[1].Command != "make build" || rows[1].Source != "Makefile" {
		t.Errorf("rows[1] = %+v", rows[1])
	}
}

func TestParseCommandsNoMarkers(t *testing.T) {
	if _, err := ParseCommands("nothing here"); !errors.Is(err, ErrNoMarkers) {
		t.Errorf("expected ErrNoMarkers, got %v", err)
	}
}
