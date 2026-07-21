// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package fstree provides gitignore-aware filesystem tree walking and
// glob-based file filtering. Walk enumerates candidate source files under a
// repository root; Filter narrows that list by include/exclude glob patterns.
// Both functions are scope-agnostic and are shared across commands.
package fstree

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// git ls-files arguments.
const (
	gitArgCached          = "--cached"
	gitArgOthers          = "--others"
	gitArgExcludeStandard = "--exclude-standard"
)

// Glob and regexp building constants.
const (
	globDoubleStar = "**"
	reAny          = ".*"     // regexp expansion for **
	reNonSep       = "[^/]*"  // regexp expansion for *
	reNonSepSingle = "[^/]"   // regexp expansion for ?
	reLiteralDot   = `\.`     // regexp expansion for .
	reMatchNothing = `^\x00$` // fallback pattern that matches nothing
)

// Mockable runner for git ls-files enumeration.
var gitLsFilesCmd = func(dir string) ([]byte, error) {
	cmd := exec.Command("git", "ls-files",
		gitArgCached,          // tracked files
		gitArgOthers,          // untracked files
		gitArgExcludeStandard, // honor .gitignore
	)
	cmd.Dir = dir
	return cmd.Output()
}

// Mockable runner for directory tree walking.
var walkDir = filepath.WalkDir

// Directory name patterns always skipped during the fallback filesystem walk.
// Only the git internals directory is unconditionally excluded here; all other
// filtering (vendor, node_modules, dist, etc.) is handled by Filter using the
// config exclude patterns, keeping the user in full control.
var fsSkipPatterns = []string{".git"}

// Walk returns all candidate source files under repoRoot.
//
// When useGitignore is true and the directory is a git repository, git ls-files
// is used so that .gitignore rules are honored and untracked-but-ignored files
// are excluded. When useGitignore is false, or when git is unavailable, a plain
// filesystem walk is used instead.
//
// Walk does not apply include/exclude filtering; use Filter for that.
func Walk(repoRoot string, useGitignore bool) ([]string, error) {
	if useGitignore {
		if files, err := gitWalk(repoRoot); err == nil {
			return files, nil
		}
	}
	return fsWalk(repoRoot)
}

// gitWalk enumerates files via git ls-files. Returns absolute OS paths.
func gitWalk(repoRoot string) ([]string, error) {
	out, err := gitLsFilesCmd(repoRoot)
	if err != nil {
		return nil, err
	}
	return parseGitLines(out, repoRoot), nil
}

// parseGitLines converts raw git ls-files output into absolute OS paths rooted
// at repoRoot. Blank and whitespace-only lines are skipped.
func parseGitLines(out []byte, repoRoot string) []string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, filepath.Join(repoRoot, filepath.FromSlash(line)))
	}
	return files
}

// shouldSkipDir reports whether a directory name matches any fsSkipPattern.
func shouldSkipDir(name string) bool {
	for _, pattern := range fsSkipPatterns {
		if globMatch(pattern, name) {
			return true
		}
	}
	return false
}

// fsWalk is a plain filesystem walk that skips directories matching fsSkipPatterns.
func fsWalk(repoRoot string) ([]string, error) {
	var files []string
	err := walkDir(repoRoot, func(path string, d os.DirEntry, entryErr error) error {
		if entryErr == nil {
			if d.IsDir() {
				if shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
			} else {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// Filter applies include and exclude glob patterns, returning only files that
// match at least one include pattern and no exclude pattern.
// Exclude always wins when both patterns match the same file.
func Filter(files, include, exclude []string) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if !matchesAny(f, include) {
			continue
		}
		if matchesAny(f, exclude) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// matchesAny reports whether path (or its base name) matches any pattern.
func matchesAny(path string, patterns []string) bool {
	base := filepath.Base(path)
	norm := filepath.ToSlash(path)
	for _, pattern := range patterns {
		if globMatch(pattern, norm) || globMatch(pattern, base) {
			return true
		}
	}
	return false
}

// globMatch reports whether path matches pattern with ** support.
func globMatch(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	if !strings.Contains(pattern, globDoubleStar) {
		matched, err := filepath.Match(pattern, path)
		return err == nil && matched
	}
	return doublestarRegexp(pattern).MatchString(path)
}

// buildDoublestarPattern converts a glob pattern containing ** into a regexp
// source string. Rules:
//
//	**  → .*      (any characters including path separators)
//	*   → [^/]*   (any characters except path separator)
//	?   → [^/]    (any single character except path separator)
//	.   → \.      (literal dot)
//	other → literal character
func buildDoublestarPattern(pattern string) string {
	var sb strings.Builder
	sb.WriteByte('^')

	i := 0
	for i < len(pattern) {
		switch {
		case i+1 < len(pattern) && pattern[i] == '*' && pattern[i+1] == '*':
			sb.WriteString(reAny)
			i += 2
			// Consume an optional trailing slash so "**/foo" works.
			if i < len(pattern) && pattern[i] == '/' {
				i++
			}
		case pattern[i] == '*':
			sb.WriteString(reNonSep)
			i++
		case pattern[i] == '?':
			sb.WriteString(reNonSepSingle)
			i++
		case pattern[i] == '.':
			sb.WriteString(reLiteralDot)
			i++
		default:
			sb.WriteByte(pattern[i])
			i++
		}
	}

	sb.WriteByte('$')
	return sb.String()
}

// doublestarRegexp compiles a glob pattern containing ** into a *regexp.Regexp.
// If the compiled pattern is invalid, it returns a regexp that matches nothing.
func doublestarRegexp(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(buildDoublestarPattern(pattern))
	if err != nil {
		re = regexp.MustCompile(reMatchNothing)
	}
	return re
}
