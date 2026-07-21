// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package lang provides language detection for repository source trees.
// Language names follow the canonical casing used by go-enry / GitHub Linguist
// (e.g. "Go", "TypeScript", "Python").
package lang

import (
	"path/filepath"
	"sort"

	enry "github.com/go-enry/go-enry/v2"
)

// Composition is the language makeup of a repository — which languages are
// present and how many files each accounts for.
type Composition struct {
	// Primary is the language with the highest file count, or "" if no
	// recognized source files were found.
	Primary string
	// All lists every detected language ordered by file count descending.
	All []string
	// FileCounts maps each detected language name to its file count.
	FileCounts map[string]int
}

// Classify returns the canonical language name for a file path, or "" if the
// file is vendored, generated, or unrecognized.
func Classify(path string) string {
	if isExcluded(path) {
		return ""
	}
	return detectLanguage(filepath.Base(path))
}

// isExcluded reports whether a path should be skipped entirely — vendored
// directories and generated files do not count toward language composition.
func isExcluded(path string) bool {
	return enry.IsVendor(path) || enry.IsGenerated(path, nil)
}

// detectLanguage returns the programming language for a base filename.
// Filename-based detection takes precedence over extension-based detection.
// When an extension is shared by multiple languages, the first Programming-type
// candidate in the enry ranked list is used to avoid false positives from
// data/markup formats (e.g. .ts → TypeScript not XML).
func detectLanguage(base string) string {
	if l, safe := enry.GetLanguageByFilename(base); safe {
		return l
	}
	return firstProgrammingLanguage(enry.GetLanguagesByExtension(base, nil, nil))
}

// firstProgrammingLanguage returns the first Programming-type language from
// candidates, or "" if none qualify.
func firstProgrammingLanguage(candidates []string) string {
	for _, l := range candidates {
		if enry.GetLanguageType(l) == enry.Programming {
			return l
		}
	}
	return ""
}

// Count maps each recognized language name to its file count across files.
func Count(files []string) map[string]int {
	counts := make(map[string]int)
	for _, f := range files {
		if l := Classify(f); l != "" {
			counts[l]++
		}
	}
	return counts
}

// Build constructs a Composition from a language→count map, ordered by file
// count descending with ties broken alphabetically.
func Build(counts map[string]int) Composition {
	ranked := rankByCount(counts)
	c := Composition{FileCounts: counts}
	for _, e := range ranked {
		c.All = append(c.All, e.name)
	}
	if len(c.All) > 0 {
		c.Primary = c.All[0]
	}
	return c
}

// rankByCount returns the language names sorted by count descending, ties
// broken alphabetically.
func rankByCount(counts map[string]int) []struct {
	name  string
	count int
} {
	entries := make([]struct {
		name  string
		count int
	}, 0, len(counts))
	for name, c := range counts {
		entries = append(entries, struct {
			name  string
			count int
		}{name, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})
	return entries
}
