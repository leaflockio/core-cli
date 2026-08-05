// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package fstree provides plain filesystem tree walking and glob-based file
// filtering, shared across commands.
//
// Example tree used below, with cwd /repo:
//
//	/repo
//	├── main.go
//	├── config
//	│   ├── settings.json
//	│   └── settings.yaml
//	└── src
//	    ├── index.js
//	    └── App.jsx
package fstree

import (
	"os"
	"path/filepath"

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

// Filter keeps only the files that match at least one include pattern and no
// exclude pattern. Exclude always wins when a file matches both.
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
//
// Matching is always case-sensitive, and an invalid pattern matches nothing
// rather than erroring.
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

// matchesAny(path, patterns) reports whether path matches by its full path
// or its base name:
//
//	matchesAny("config/settings.json", ["*.json"]) → true (matched by base name)
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
