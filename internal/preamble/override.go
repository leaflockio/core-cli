// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package preamble

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/leaflockio/core-cli/internal/errs"
)

// largePreserveLinesThreshold is the PreserveLines value above which
// ApplyOverrides warns that an entry is unusually large for a preamble.
const largePreserveLinesThreshold = 10

type PatternOverride struct {
	Pattern string
	Exts    []string
}

// compiledPattern is a PatternOverride with its regex already compiled.
type compiledPattern struct {
	re   *regexp.Regexp
	exts []string
}

// patterns holds every pattern override added via ApplyOverrides, checked
// in addition to the built-in markers.
var patterns []compiledPattern

// preserveLines maps a doublestar glob pattern to the number of leading
// lines to treat as preamble unconditionally for files matching it,
// bypassing markers and patterns.
var preserveLines map[string]int

// ApplyOverrides compiles overrides and adds them to the set Detect
// checks, alongside the built-in markers, and sets perFile as the
// PreserveLines exceptions Detect resolves per file. Every value in
// perFile must be a positive integer. Nothing from this call takes
// effect unless every pattern compiles and every perFile value is
// positive.
//
// The returned warnings report every perFile entry above
// largePreserveLinesThreshold — unusually large for a preamble, likely a
// misconfiguration — sorted by pattern for determinism. These entries
// are still applied; a warning is advisory, not a rejection.
func ApplyOverrides(overrides []PatternOverride, perFile map[string]int) ([]string, error) {
	compiled := make([]compiledPattern, 0, len(overrides))
	for _, o := range overrides {
		re, err := regexp.Compile(o.Pattern)
		if err != nil {
			return nil, errs.Caller(errs.PRE001,
				"preamble pattern is not a valid regular expression",
				err,
				errs.Context{
					Cause:      "pattern " + o.Pattern + " could not be compiled as a regular expression",
					Resolution: "fix the pattern's regular expression syntax",
				},
			)
		}
		compiled = append(compiled, compiledPattern{re: re, exts: o.Exts})
	}

	for pattern, n := range perFile {
		if n <= 0 {
			return nil, errs.Caller(errs.PRE002,
				"preamble PreserveLines value is not a positive integer",
				nil,
				errs.Context{
					Cause:      "per_file entry " + pattern + " has PreserveLines " + strconv.Itoa(n),
					Resolution: "set PreserveLines to a positive integer, or remove the entry",
				},
			)
		}
	}

	names := make([]string, 0, len(perFile))
	for pattern := range perFile {
		names = append(names, pattern)
	}
	sort.Strings(names)
	var warnings []string
	for _, pattern := range names {
		if n := perFile[pattern]; n > largePreserveLinesThreshold {
			warnings = append(warnings, fmt.Sprintf(
				"per_file entry %s has PreserveLines %d, unusually large for a preamble", pattern, n,
			))
		}
	}

	patterns = append(patterns, compiled...)
	preserveLines = perFile
	return warnings, nil
}

// matchPatterns consumes every leading line matching any pattern in
// patterns that's in scope for ext and base.
func matchPatterns(lines []string, ext, base string) int {
	i := 0
	for i < len(lines) && matchesAnyPattern(lines[i], ext, base) {
		i++
	}
	return i
}

func matchesAnyPattern(line, ext, base string) bool {
	for _, p := range patterns {
		if scopeApplies(p.exts, ext, base) && p.re.MatchString(line) {
			return true
		}
	}
	return false
}

// resolvePreserveLines returns the PreserveLines value that applies to
// path: the first entry in preserveLines — in alphabetical pattern
// order, for determinism, since map iteration order isn't — whose
// doublestar glob pattern matches path. Returns 0 (unset) when nothing
// matches. An invalid pattern matches nothing rather than erroring.
func resolvePreserveLines(path string) int {
	names := make([]string, 0, len(preserveLines))
	for p := range preserveLines {
		names = append(names, p)
	}
	sort.Strings(names)

	slashPath := filepath.ToSlash(path)
	for _, p := range names {
		if matched, err := doublestar.Match(filepath.ToSlash(p), slashPath); err == nil && matched {
			return preserveLines[p]
		}
	}
	return 0
}
