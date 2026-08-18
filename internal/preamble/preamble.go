// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package preamble detects leading file content that must stay before any
// inserted license header — a shebang line, an XML declaration, Go build
// constraints, and similar constructs that would break the file if a
// header were inserted above them.
package preamble

import (
	"path/filepath"
	"strings"
)

// Detect returns the number of leading lines in lines that make up a
// known preamble for the file at path — content that must stay before
// any inserted license header. Returns 0 if none is found. It only uses
// path to scope markers that apply to specific file types, path need
// not exist on disk.
//
// Resolution order: a PreserveLines override matching path wins outright,
// bypassing markers and patterns entirely, otherwise the built-in markers
// are tried first, then config-supplied patterns.
func Detect(lines []string, path string) int {
	if n := resolvePreserveLines(path); n > 0 {
		if n > len(lines) {
			return len(lines)
		}
		return n
	}

	ext := normalizeExt(filepath.Ext(path))
	base := strings.ToLower(strings.TrimPrefix(filepath.Base(path), "."))

	for _, m := range markers {
		if !m.appliesTo(ext, base) {
			continue
		}
		if n := m.match(lines); n > 0 {
			return n
		}
	}
	return matchPatterns(lines, ext, base)
}

func normalizeExt(ext string) string {
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}
