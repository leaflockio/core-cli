// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package fstree

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// MatchAll is the pattern that matches every file, at any depth.
const MatchAll = "**/*"

var walkDir = filepath.WalkDir

// fsSkipPatterns are directory names never descended into during a walk.
var fsSkipPatterns = []string{".git"}

// Walk lists every file under root, recursively, skipping fsSkipPatterns.
//
//	Walk("config") → ["config/settings.json", "config/settings.yaml"]
//	Walk(".")       → every file in the tree
func Walk(root string) ([]string, error) {
	return fsWalk(root)
}

// shouldSkipDir reports whether a directory name is in fsSkipPatterns.
func shouldSkipDir(name string) bool {
	for _, pattern := range fsSkipPatterns {
		if globMatch(pattern, name) {
			return true
		}
	}
	return false
}

// fsWalk is Walk's implementation.
func fsWalk(dir string) ([]string, error) {
	var files []string
	err := walkDir(dir, func(path string, d os.DirEntry, entryErr error) error {
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

// Filter keeps only the files that match at least one include pattern and
// aren't excluded.
//
//	Filter(all, ["**/*.go"], nil)      → ["main.go"]
//	Filter(all, ["**/*"], ["*.json"])  → everything except config/settings.json
//	Filter(all, nil, nil)              → [] (no include patterns means nothing matches)
//
// If a caller wants "everything" as the default instead of "nothing," pass
// MatchAll rather than an empty include list.
//
//	no "/" in the pattern   matches at any depth, by file name alone
//	                            "*.json"            → config/settings.json, a/config/settings.json, ...
//	has a "/" in the pattern   anchored — must match the full path exactly
//	                            "config/*.json"     → config/settings.json, not a/config/settings.json
//	*                       any characters except "/"
//	**                      any characters, including "/" — the only way to cross directories
//	                            "**/config/*.json"  → config/settings.json and a/config/settings.json
//	?                       exactly one character except "/"
//	[abc] [a-z]             one character from a set or range
//	[^abc] or [!abc]        one character NOT in the set
//	{a,b,c}                 matches if any comma-separated alternative matches (nestable)
//	                            "config/*.{json,yaml}" → config/settings.json, config/settings.yaml
//	\                       escapes the next character to match it literally
//	                            `main\.go` matches "main.go" but not "mainXgo"
//	trailing "/"            matches only a directory, never a file of the same name
//	                            "build/" excludes src/build/output.js, not a file named build
//
// exclude is order-sensitive: a "!pattern" entry un-excludes a file matched
// by an earlier entry, so the entry that appears *last* in exclude wins,
// This lets you exclude a whole directory and carve out one exception:
//
//	Filter(all, ["**/*"], ["vendor/**", "!vendor/utils/patched.go"])
//	    → everything except vendor, but vendor/utils/patched.go is kept
//
// include has no such order-sensitivity or negation — a file either matches
// one of its patterns or it doesn't; there's nothing to un-include that
// exclude doesn't already cover.
//
// Matching is always case-sensitive, and an invalid pattern matches nothing
// rather than erroring.
func Filter(files, include, exclude []string) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if !matchesAny(f, include) {
			continue
		}
		if isExcluded(f, exclude) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// matchesAny(path, patterns) reports whether path matches by its full path
// or its base name:
//
//	matchesAny("config/settings.json", ["*.json"]) → true (matched by base name)
func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesOne(path, pattern) {
			return true
		}
	}
	return false
}

// isExcluded(path, exclude) evaluates exclude in order, letting a "!pattern"
// entry cancel a match from an earlier entry — the last entry in exclude
// that matches path decides the outcome:
//
//	isExcluded("vendor/patched.go", ["vendor/**"])                          → true
//	isExcluded("vendor/patched.go", ["vendor/**", "!vendor/patched.go"])    → false
//	isExcluded("vendor/patched.go", ["!vendor/patched.go", "vendor/**"])    → true (later entry wins)
func isExcluded(path string, exclude []string) bool {
	excluded := false
	for _, pattern := range exclude {
		if negated, ok := strings.CutPrefix(pattern, "!"); ok {
			if matchesOne(path, negated) {
				excluded = false
			}
			continue
		}
		if matchesOne(path, pattern) {
			excluded = true
		}
	}
	return excluded
}

// matchesOne(path, pattern) reports whether path matches pattern by its full
// path or its base name. A trailing "/" on pattern means "directory only" —
// path itself is never a directory , so that case matches against path's
// ancestor directories instead:
//
//	matchesOne("config/settings.json", "*.json")   → true (matched by base name)
//	matchesOne("src/build/output.js", "build/")    → true (build is an ancestor dir)
//	matchesOne("build", "build/")                  → false (build here is a file, not a dir)
func matchesOne(path, pattern string) bool {
	if dirPattern, ok := strings.CutSuffix(pattern, "/"); ok {
		return matchesAncestorDir(path, dirPattern)
	}
	base := filepath.Base(path)
	norm := filepath.ToSlash(path)
	return globMatch(pattern, norm) || globMatch(pattern, base)
}

// matchesAncestorDir(path, dirPattern) reports whether any ancestor
// directory of path matches dirPattern — e.g. for "src/build/output.js",
// the ancestors checked are "src" and "src/build".
func matchesAncestorDir(path, dirPattern string) bool {
	segments := strings.Split(filepath.ToSlash(path), "/")
	for i := 1; i < len(segments); i++ {
		ancestor := strings.Join(segments[:i], "/")
		if matchesOne(ancestor, dirPattern) {
			return true
		}
	}
	return false
}

// globMatch(pattern, path) reports whether path matches pattern
//
//	globMatch("*.{js,jsx}", "index.js")    → true
//	globMatch("*.{js,jsx}", "settings.json") → false
//
// An invalid pattern matches nothing rather than erroring.
func globMatch(pattern, path string) bool {
	matched, err := doublestar.Match(filepath.ToSlash(pattern), path)
	return err == nil && matched
}
