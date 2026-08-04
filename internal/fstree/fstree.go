// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package fstree provides plain filesystem tree walking and glob-based file
// filtering, shared across commands.
package fstree

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

var walkDir = filepath.WalkDir

// Directory name patterns always skipped during the walk.
var fsSkipPatterns = []string{".git"}

// Walk returns all candidate source files under root via a plain filesystem walk.
func Walk(root string) ([]string, error) {
	return fsWalk(root)
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
