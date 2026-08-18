// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package preamble

import "strings"

// Kind is the shape of a marker's match.
type Kind int

const (
	// SingleLine: only line 0 is checked; a match consumes exactly 1
	// line.
	SingleLine Kind = iota
	// RepeatingLines: consecutive matching lines are consumed.
	RepeatingLines
)

// marker is a known construct that, when found at the top of a file,
// must stay before any inserted license header.
type marker struct {
	// prefixes are checked case-insensitively against the start of a
	// line.
	prefixes []string
	// exts restricts the marker to specific file extensions or bare
	// filenames. A nil exts applies to every file, unscoped.
	exts []string
	kind Kind
}

var markers = []marker{
	{prefixes: []string{"#!"}, kind: SingleLine},
	{prefixes: []string{"<?xml"}, kind: SingleLine},
	{prefixes: []string{"<!doctype"}, kind: SingleLine},
	{prefixes: []string{"# encoding:"}, kind: SingleLine, exts: []string{"rb"}},
	{prefixes: []string{"# frozen_string_literal:"}, kind: SingleLine, exts: []string{"rb"}},
	{prefixes: []string{"<?php"}, kind: SingleLine, exts: []string{"php"}},
	{prefixes: []string{"# escape", "# syntax"}, kind: SingleLine, exts: []string{"dockerfile", "containerfile"}},
	{prefixes: []string{"//go:build", "// +build"}, kind: RepeatingLines, exts: []string{"go"}},
	{prefixes: []string{`#\`}, kind: SingleLine, exts: []string{"config.ru"}},
}

// appliesTo reports whether m's scope includes ext or base.
// A nil exts applies to every file.
func (m marker) appliesTo(ext, base string) bool {
	return scopeApplies(m.exts, ext, base)
}

// scopeApplies reports whether exts includes ext or base.
// A nil exts applies to every file.
func scopeApplies(exts []string, ext, base string) bool {
	if exts == nil {
		return true
	}
	for _, e := range exts {
		e = normalizeExt(e)
		if e == ext || e == base {
			return true
		}
	}
	return false
}

// match reports how many leading lines of lines m consumes.
func (m marker) match(lines []string) int {
	if m.kind == RepeatingLines {
		return m.matchRepeating(lines)
	}
	return m.matchSingle(lines)
}

// matchSingle checks only line 0 against m's prefixes.
func (m marker) matchSingle(lines []string) int {
	if len(lines) == 0 {
		return 0
	}
	first := strings.ToLower(lines[0])
	for _, p := range m.prefixes {
		if strings.HasPrefix(first, p) {
			return 1
		}
	}
	return 0
}

// matchRepeating consumes every leading line matching any of m's
// prefixes.
func (m marker) matchRepeating(lines []string) int {
	i := 0
	sawMatch := false
	for i < len(lines) {
		if m.hasAnyPrefix(lines[i]) {
			sawMatch = true
			i++
			continue
		}
		if sawMatch && lines[i] == "" {
			return i + 1
		}
		break
	}
	return i
}

func (m marker) hasAnyPrefix(line string) bool {
	for _, p := range m.prefixes {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}
