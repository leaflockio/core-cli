// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package preamble

import "testing"

func TestDetect_noPreamble(t *testing.T) {
	lines := []string{"package main", "", "func main() {}"}
	if got := Detect(lines, "main.go"); got != 0 {
		t.Errorf("Detect = %d, want 0", got)
	}
}

func TestDetect_unscopedMarkerAppliesToAnyExtension(t *testing.T) {
	lines := []string{"#!/usr/bin/env node", "console.log('hi')"}
	if got := Detect(lines, "script.js"); got != 1 {
		t.Errorf("Detect(shebang, .js) = %d, want 1", got)
	}
	if got := Detect(lines, "script.custom"); got != 1 {
		t.Errorf("Detect(shebang, unknown ext) = %d, want 1", got)
	}
}

func TestDetect_scopedMarkerAppliesOnlyToItsExtension(t *testing.T) {
	lines := []string{"# encoding: utf-8", "puts 'hi'"}
	if got := Detect(lines, "script.rb"); got != 1 {
		t.Errorf("Detect(encoding, .rb) = %d, want 1", got)
	}
	if got := Detect(lines, "script.py"); got != 0 {
		t.Errorf("Detect(encoding, .py) = %d, want 0 (out of scope)", got)
	}
}

func TestDetect_bareFilenameMarker(t *testing.T) {
	lines := []string{"# syntax=docker/dockerfile:1", "FROM golang"}
	if got := Detect(lines, "Dockerfile"); got != 1 {
		t.Errorf("Detect(syntax, Dockerfile) = %d, want 1", got)
	}
	if got := Detect(lines, "path/to/Dockerfile"); got != 1 {
		t.Errorf("Detect(syntax, path/to/Dockerfile) = %d, want 1", got)
	}
}

func TestDetect_goBuildTag(t *testing.T) {
	lines := []string{"//go:build linux", "// +build linux", "", "package main"}
	if got := Detect(lines, "main.go"); got != 3 {
		t.Errorf("Detect(go build tag) = %d, want 3", got)
	}
}

func TestDetect_firstMatchingMarkerWins(t *testing.T) {
	// An XML declaration is unscoped, so it applies to .xml too; confirm
	// the file's own content still drives the result, not the extension.
	lines := []string{`<?xml version="1.0"?>`, "<root/>"}
	if got := Detect(lines, "config.xml"); got != 1 {
		t.Errorf("Detect(xml) = %d, want 1", got)
	}
}

func TestNormalizeExt(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"go", "go"},
		{".go", "go"},
		{"GO", "go"},
		{".GO", "go"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := normalizeExt(tt.in); got != tt.want {
			t.Errorf("normalizeExt(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
