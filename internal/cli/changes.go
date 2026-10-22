package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// changedLines returns the 1-based line numbers of file (relative to root)
// that were added or modified in the working tree relative to base. It is the
// input of an incremental audit (audit --since): the checks are unchanged, but
// only findings on these lines count as new.
//
// The engine itself never shells out; this lives in the CLI so that
// internal/audit stays a pure function of the filesystem. git is invoked with
// a zero-context diff, so the hunk headers carry exactly the ranges to mark.
//
// A file unknown to the base (added, renamed, or untracked) makes every line
// changed, which is what "new since base" means for a fresh context file.
func changedLines(root, file, base string) (map[int]bool, error) {
	rel, err := filepath.Rel(root, file)
	if err != nil {
		rel = file
	}
	rel = filepath.ToSlash(rel)

	if !isGitRepo(root) {
		return nil, fmt.Errorf("--since needs a git repository; %s is not inside one", root)
	}
	if err := gitCheckRev(root, base); err != nil {
		return nil, err
	}
	// A path absent from the base tree is entirely new.
	if !pathInBase(root, base, rel) {
		return allLines(file)
	}

	out, err := runGit(root, "diff", "--unified=0", "--no-color", base, "--", rel)
	if err != nil {
		return nil, err
	}
	return parseDiffLines(out), nil
}

// parseDiffLines extracts the added/changed line numbers (new-side) from a
// unified diff with zero context. Each hunk header has the form
// "@@ -a,b +c,d @@"; the new-side range [c, c+d-1] is what the change touched.
func parseDiffLines(diff string) map[int]bool {
	lines := map[int]bool{}
	for _, line := range strings.Split(diff, "\n") {
		if !strings.HasPrefix(line, "@@ ") {
			continue
		}
		plus := strings.Index(line, " +")
		if plus < 0 {
			continue
		}
		rest := line[plus+2:]
		if end := strings.Index(rest, " @@"); end >= 0 {
			rest = rest[:end]
		}
		rest = strings.TrimSpace(rest)
		startPart, countPart, _ := strings.Cut(rest, ",")
		start, err := strconv.Atoi(strings.TrimSpace(startPart))
		if err != nil {
			continue
		}
		count := 1
		if countPart != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(countPart)); err == nil {
				count = n
			}
		}
		for i := 0; i < count; i++ {
			lines[start+i] = true
		}
	}
	return lines
}

// allLines marks every line of file as changed.
func allLines(file string) (map[int]bool, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines := map[int]bool{}
	for i := 1; i <= strings.Count(string(b), "\n")+1; i++ {
		lines[i] = true
	}
	return lines, nil
}

func isGitRepo(dir string) bool {
	_, err := runGit(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func gitCheckRev(root, rev string) error {
	if _, err := runGit(root, "rev-parse", "--verify", "--quiet", rev+"^{commit}"); err != nil {
		return fmt.Errorf("--since: cannot resolve revision %q", rev)
	}
	return nil
}

// pathInBase reports whether rel exists in the base tree.
func pathInBase(root, base, rel string) bool {
	out, err := runGit(root, "ls-tree", "-r", base, "--", rel)
	return err == nil && strings.TrimSpace(out) != ""
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
		}
		return "", err
	}
	return stdout.String(), nil
}
