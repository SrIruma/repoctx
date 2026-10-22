// Package audit detects context rot in AI coding-agent context files:
// ghost commands claimed by the file that no longer exist in the project, and
// stale paths referenced anywhere in the file that are gone from the repo.
package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SrIruma/repoctx/internal/markdown"
	"github.com/SrIruma/repoctx/internal/project"
)

// Issue is a single rot finding.
type Issue struct {
	Command string `json:"command,omitempty"`
	Path    string `json:"path,omitempty"`
	Detail  string `json:"detail"`
	// Line is the 1-based line in the context file the finding came from.
	// It is zero when the finding cannot be tied to a line.
	Line int `json:"line,omitempty"`
	// Scope classifies the finding against the diff base, when the caller
	// asked for one (audit --since). It is empty when no base was given,
	// "new" when the finding's line was changed since the base, and
	// "pre-existing" otherwise.
	Scope string `json:"scope,omitempty"`
}

// Check is the result of one audit check.
type Check struct {
	Name   string  `json:"name"`
	Passed bool    `json:"passed"`
	Detail string  `json:"detail"`
	Issues []Issue `json:"issues,omitempty"`
}

// Report is the audit result for one context file.
type Report struct {
	File   string  `json:"file"`
	Checks []Check `json:"checks"`
	Score  int     `json:"score"`
	Passed bool    `json:"passed"`
	// NewIssues and PreExisting are set only for an incremental audit
	// (audit --since). NewIssues counts findings introduced since the base and
	// drives Passed; PreExisting counts inherited rot, reported but not gated.
	NewIssues   int `json:"new_issues,omitempty"`
	PreExisting int `json:"pre_existing,omitempty"`
}

// Options configures an audit run.
type Options struct {
	Root   string            // repository root
	File   string            // path to the context file
	Actual []project.Command // commands extractable from the project

	// Since, when non-zero, is the set of 1-based line numbers of File that
	// were added or changed since a diff base. Each finding whose line is in
	// the set is classified "new"; every other finding is "pre-existing".
	// When the set is nil (the default) no classification happens and every
	// finding counts as a failure, which is the historical behaviour.
	//
	// Since is the only thing that differs between a whole-repo audit and an
	// incremental one; the checks themselves are unchanged, so the engine
	// never has to know how the set was computed (git, or anything else).
	Since map[int]bool
}

// Run audits the context file at Options.File and returns its report.
func Run(o Options) (*Report, error) {
	content, err := os.ReadFile(o.File)
	if err != nil {
		return nil, err
	}
	text := string(content)
	report := &Report{File: o.File}
	report.Checks = append(report.Checks, checkCommands(text, o.Actual))
	report.Checks = append(report.Checks, checkPaths(text, o.Root, filepath.Dir(o.File), o.Actual))
	classify(report, o.Since)
	passed := 0
	for _, c := range report.Checks {
		if c.Passed {
			passed++
		}
	}
	report.Score = passed * 100 / len(report.Checks)
	report.Passed = report.Score == 100
	return report, nil
}

// classify tags each issue as "new" or "pre-existing" against the changed-line
// set and recomputes which checks fail. With no set (the default whole-repo
// audit) issues keep an empty scope and every check that found one fails.
//
// When a set is given, a check fails only on its "new" issues: rot the caller
// inherited is reported but does not gate. The per-check and per-report counts
// are refreshed so the score reflects the new-only verdict.
func classify(report *Report, since map[int]bool) {
	if since == nil {
		return
	}
	report.NewIssues = 0
	report.PreExisting = 0
	for i := range report.Checks {
		c := &report.Checks[i]
		newCount, preCount := 0, 0
		for j := range c.Issues {
			if c.Issues[j].Line != 0 && since[c.Issues[j].Line] {
				c.Issues[j].Scope = "new"
				newCount++
			} else {
				c.Issues[j].Scope = "pre-existing"
				preCount++
			}
		}
		report.NewIssues += newCount
		report.PreExisting += preCount
		c.Passed = newCount == 0
		switch {
		case newCount > 0 && preCount > 0:
			c.Detail = fmt.Sprintf("%d new and %d pre-existing issue(s) since base", newCount, preCount)
		case newCount > 0:
			c.Detail = fmt.Sprintf("%d new issue(s) since base", newCount)
		case preCount > 0:
			c.Detail = fmt.Sprintf("%d pre-existing issue(s); none new since base", preCount)
		}
	}
}

// checkCommands flags commands claimed between the markers that cannot be
// extracted from the project anymore.
func checkCommands(content string, actual []project.Command) Check {
	if !markdown.HasMarkers(content) {
		return Check{Name: "commands", Passed: true, Detail: "no repoctx markers; commands check skipped"}
	}
	rows, err := markdown.ParseCommands(content)
	if err != nil {
		return Check{Name: "commands", Passed: false, Detail: err.Error()}
	}
	available := map[string]bool{}
	for _, c := range actual {
		available[c.Cmd] = true
	}
	c := Check{Name: "commands", Passed: true, Detail: fmt.Sprintf("%d commands claimed", len(rows))}
	for _, r := range rows {
		if !available[r.Command] {
			c.Passed = false
			c.Issues = append(c.Issues, Issue{
				Command: r.Command,
				Line:    r.Line,
				Detail:  "ghost command: no longer present in the project",
			})
		}
	}
	if !c.Passed {
		c.Detail = fmt.Sprintf("%d ghost command(s) out of %d claimed", len(c.Issues), len(rows))
	}
	return c
}

var (
	backtickPath = regexp.MustCompile("`([^`]+)`")
	linkPath     = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)\)`)
)

// contextFileNames are the files repoctx itself manages. Mentioning them is
// never "rot", so they are excluded from stale-path candidates.
var contextFileNames = map[string]bool{
	"AGENTS.md": true,
	"CLAUDE.md": true,
}

// knownRootFiles are extensionless top-level files worth treating as paths.
var knownRootFiles = map[string]bool{
	"Makefile": true, "GNUmakefile": true, "LICENSE": true, "NOTICE": true,
	"Dockerfile": true, "Jenkinsfile": true, "Gemfile": true, "Procfile": true,
}

// knownExtensions are file extensions considered path-like. The list is
// deliberately finite: an unknown suffix is not treated as a file, so a Go
// symbol such as "internal/cli.version" is never flagged as a stale path.
var knownExtensions = map[string]bool{
	"go": true, "rs": true, "py": true, "md": true, "txt": true,
	"toml": true, "json": true, "yaml": true, "yml": true, "xml": true,
	"sh": true, "bash": true, "ts": true, "tsx": true, "js": true, "jsx": true,
	"mod": true, "sum": true, "lock": true, "env": true, "html": true, "css": true,
}

// extensionOf returns the part of s after the last dot, or "" when s has no dot
// or ends with one.
func extensionOf(s string) string {
	i := strings.LastIndex(s, ".")
	if i < 0 || i == len(s)-1 {
		return ""
	}
	return s[i+1:]
}

// checkPaths flags repository paths mentioned anywhere in the context file
// that no longer exist on disk. Paths are collected from backticks and
// markdown links; commands and URLs are excluded to avoid false positives.
// A path is resolved relative to the context file's own directory first (so a
// nested AGENTS.md refers to its own subtree), then relative to the repository
// root (so a file can still cite root-relative paths). It is stale only when
// neither base resolves it.
func checkPaths(content string, root, fileDir string, actual []project.Command) Check {
	exclude := map[string]bool{}
	for _, c := range actual {
		exclude[c.Cmd] = true
		exclude[c.Name] = true
	}
	c := Check{Name: "paths", Passed: true, Detail: "all referenced paths exist"}
	for _, cand := range candidatePaths(content) {
		if exclude[cand.value] {
			continue
		}
		if !pathExists(fileDir, cand.value) && !pathExists(root, cand.value) {
			c.Passed = false
			c.Issues = append(c.Issues, Issue{
				Path:   cand.value,
				Line:   cand.line,
				Detail: "stale path: not found in repository",
			})
		}
	}
	if !c.Passed {
		c.Detail = fmt.Sprintf("%d stale path(s)", len(c.Issues))
	}
	return c
}

// pathExists reports whether cand exists under base. Absolute candidates are
// checked directly.
func pathExists(base, cand string) bool {
	p := cand
	if !filepath.IsAbs(p) {
		p = filepath.Join(base, cand)
	}
	_, err := os.Stat(p)
	return err == nil
}

// pathCandidate is a path-like reference found in a context file, together
// with the 1-based line number it appeared on.
type pathCandidate struct {
	value string
	line  int
}

// candidatePaths extracts path-like references from a context file, with the
// line each was found on.
//
// Candidates are collected line by line. The backtick pattern matches
// whatever sits between two backticks, so scanning the whole file at once
// also matches the gap between two separate inline spans ("... `code`.\n- `other`")
// and the body of a fenced block. Neither is a path, and both used to be
// reported as stale paths that do not exist.
func candidatePaths(content string) []pathCandidate {
	seen := map[string]bool{}
	var out []pathCandidate
	add := func(s string, line int) {
		s = strings.TrimSpace(s)
		if i := strings.Index(s, "#"); i >= 0 {
			s = s[:i]
		}
		s = strings.TrimSpace(s)
		if s == "" || seen[s] || contextFileNames[s] || !isPathLike(s) {
			return
		}
		seen[s] = true
		out = append(out, pathCandidate{value: s, line: line})
	}
	for i, line := range strings.Split(content, "\n") {
		lineNo := i + 1
		for _, m := range backtickPath.FindAllStringSubmatch(line, -1) {
			add(m[1], lineNo)
		}
		for _, m := range linkPath.FindAllStringSubmatch(line, -1) {
			add(m[1], lineNo)
		}
	}
	return out
}

// isPathLike filters candidates that plausibly reference a repository path:
// directories, known extensionless files, dotfiles, and files with a known
// extension. Unknown-suffix symbols such as "internal/cli.version" are not
// treated as paths.
func isPathLike(s string) bool {
	if strings.ContainsAny(s, " \t") || strings.Contains(s, "://") || strings.HasPrefix(s, "#") {
		return false
	}
	s = strings.TrimSuffix(s, "/")
	if strings.HasPrefix(s, ".") {
		return true
	}
	if strings.Contains(s, "/") {
		last := s[strings.LastIndex(s, "/")+1:]
		if !strings.Contains(last, ".") {
			return true // directory path
		}
		return knownExtensions[extensionOf(last)]
	}
	if knownRootFiles[s] {
		return true
	}
	return knownExtensions[extensionOf(s)]
}
