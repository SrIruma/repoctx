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
}

// Options configures an audit run.
type Options struct {
	Root   string            // repository root
	File   string            // path to the context file
	Actual []project.Command // commands extractable from the project
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
	report.Checks = append(report.Checks, checkPaths(text, o.Root, o.Actual))
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
	fileExt      = regexp.MustCompile(`\.[A-Za-z][A-Za-z0-9]{0,7}$`)
)

// knownRootFiles are extensionless top-level files worth treating as paths.
var knownRootFiles = map[string]bool{
	"Makefile": true, "GNUmakefile": true, "LICENSE": true, "NOTICE": true,
	"Dockerfile": true, "Jenkinsfile": true, "Gemfile": true, "Procfile": true,
}

// checkPaths flags repository paths mentioned anywhere in the context file
// that no longer exist on disk. Paths are collected from backticks and
// markdown links; commands and URLs are excluded to avoid false positives.
func checkPaths(content string, root string, actual []project.Command) Check {
	exclude := map[string]bool{}
	for _, c := range actual {
		exclude[c.Cmd] = true
		exclude[c.Name] = true
	}
	c := Check{Name: "paths", Passed: true, Detail: "all referenced paths exist"}
	for _, cand := range candidatePaths(content) {
		if exclude[cand] {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, cand)); err != nil {
			c.Passed = false
			c.Issues = append(c.Issues, Issue{
				Path:   cand,
				Detail: "stale path: not found in repository",
			})
		}
	}
	if !c.Passed {
		c.Detail = fmt.Sprintf("%d stale path(s)", len(c.Issues))
	}
	return c
}

// candidatePaths extracts path-like references from a context file.
func candidatePaths(content string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if i := strings.Index(s, "#"); i >= 0 {
			s = s[:i]
		}
		s = strings.TrimSpace(s)
		if s == "" || seen[s] || !isPathLike(s) {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, m := range backtickPath.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}
	for _, m := range linkPath.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}
	return out
}

// isPathLike filters candidates that plausibly reference a repository path:
// relative paths, known extensionless files, or files with a real extension.
func isPathLike(s string) bool {
	if strings.ContainsAny(s, " \t") || strings.Contains(s, "://") || strings.HasPrefix(s, "#") {
		return false
	}
	if strings.Contains(s, "/") {
		return true
	}
	if knownRootFiles[s] {
		return true
	}
	return fileExt.MatchString(s)
}
