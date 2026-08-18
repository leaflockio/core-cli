// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package preamble

import "testing"

func TestScopeApplies(t *testing.T) {
	tests := []struct {
		name string
		exts []string
		ext  string
		base string
		want bool
	}{
		{name: "nil exts applies to everything", exts: nil, ext: "xyz", base: "whatever", want: true},
		{name: "matches by extension", exts: []string{"go"}, ext: "go", base: "main.go", want: true},
		{name: "matches by bare filename", exts: []string{"dockerfile"}, ext: "", base: "dockerfile", want: true},
		{name: "entry normalized: uppercase", exts: []string{"GO"}, ext: "go", base: "main.go", want: true},
		{name: "entry normalized: leading dot", exts: []string{".go"}, ext: "go", base: "main.go", want: true},
		{
			name: "entry normalized: mixed case bare name",
			exts: []string{"Dockerfile"}, ext: "", base: "dockerfile", want: true,
		},
		{name: "no match", exts: []string{"rb"}, ext: "go", base: "main.go", want: false},
		{name: "empty exts slice matches nothing", exts: []string{}, ext: "go", base: "main.go", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scopeApplies(tt.exts, tt.ext, tt.base); got != tt.want {
				t.Errorf("scopeApplies(%v, %q, %q) = %v, want %v", tt.exts, tt.ext, tt.base, got, tt.want)
			}
		})
	}
}

func TestMarker_appliesTo(t *testing.T) {
	m := marker{exts: []string{"rb"}}
	if !m.appliesTo("rb", "foo.rb") {
		t.Error("appliesTo(rb, foo.rb) = false, want true")
	}
	if m.appliesTo("py", "foo.py") {
		t.Error("appliesTo(py, foo.py) = true, want false")
	}
}

func TestMarker_matchSingle(t *testing.T) {
	m := marker{prefixes: []string{"#!"}}
	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{name: "no lines", lines: []string{}, want: 0},
		{name: "matching prefix", lines: []string{"#!/usr/bin/env bash", "echo hi"}, want: 1},
		{name: "case-insensitive match", lines: []string{"#!/USR/BIN/ENV BASH"}, want: 1},
		{name: "no match", lines: []string{"package main"}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.matchSingle(tt.lines); got != tt.want {
				t.Errorf("matchSingle(%v) = %d, want %d", tt.lines, got, tt.want)
			}
		})
	}
}

func TestMarker_matchSingle_multiplePrefixes(t *testing.T) {
	m := marker{prefixes: []string{"# escape", "# syntax"}}
	if got := m.matchSingle([]string{"# escape=backtick"}); got != 1 {
		t.Errorf("matchSingle(escape) = %d, want 1", got)
	}
	if got := m.matchSingle([]string{"# syntax=docker/dockerfile:1"}); got != 1 {
		t.Errorf("matchSingle(syntax) = %d, want 1", got)
	}
}

func TestMarker_matchRepeating(t *testing.T) {
	m := marker{prefixes: []string{"//go:build", "// +build"}}
	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{
			name:  "with trailing blank line",
			lines: []string{"//go:build linux", "// +build linux", "", "package main"},
			want:  3,
		},
		{
			name:  "no trailing blank line",
			lines: []string{"//go:build linux", "package main"},
			want:  1,
		},
		{
			name:  "build tag is the entire file",
			lines: []string{"//go:build linux"},
			want:  1,
		},
		{
			name:  "extra blank line only consumes one",
			lines: []string{"//go:build linux", "", "", "package main"},
			want:  2,
		},
		{
			name:  "no match at all",
			lines: []string{"package main", "", "func main() {}"},
			want:  0,
		},
		{
			name:  "blank line between two matching lines stops early",
			lines: []string{"//go:build linux", "", "// +build linux", "package main"},
			want:  2,
		},
		{
			name:  "no lines",
			lines: []string{},
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.matchRepeating(tt.lines); got != tt.want {
				t.Errorf("matchRepeating(%v) = %d, want %d", tt.lines, got, tt.want)
			}
		})
	}
}

func TestMarker_match_dispatchesByKind(t *testing.T) {
	single := marker{prefixes: []string{"#!"}, kind: SingleLine}
	if got := single.match([]string{"#!/bin/sh", "", "// +build linux"}); got != 1 {
		t.Errorf("SingleLine match = %d, want 1", got)
	}

	repeating := marker{prefixes: []string{"//go:build"}, kind: RepeatingLines}
	if got := repeating.match([]string{"//go:build linux", "", "package main"}); got != 2 {
		t.Errorf("RepeatingLines match = %d, want 2", got)
	}
}

// TestBuiltinMarkers exercises every entry in the real markers table
// against a realistic example, confirming both that it matches when in
// scope and that ext-scoping actually excludes out-of-scope files.
func TestBuiltinMarkers(t *testing.T) {
	tests := []struct {
		name       string
		lines      []string
		ext        string
		base       string
		wantMatch  int
		outOfScope string // an ext that should NOT match, empty if unscoped
	}{
		{name: "shebang", lines: []string{"#!/usr/bin/env python3"}, ext: "py", base: "run.py", wantMatch: 1},
		{name: "xml declaration", lines: []string{`<?xml version="1.0"?>`}, ext: "xml", base: "f.xml", wantMatch: 1},
		{name: "html doctype", lines: []string{"<!doctype html>"}, ext: "html", base: "f.html", wantMatch: 1},
		{
			name: "ruby encoding", lines: []string{"# encoding: utf-8"},
			ext: "rb", base: "f.rb", wantMatch: 1, outOfScope: "py",
		},
		{
			name: "ruby frozen string literal", lines: []string{"# frozen_string_literal: true"},
			ext: "rb", base: "f.rb", wantMatch: 1, outOfScope: "py",
		},
		{name: "php open tag", lines: []string{"<?php"}, ext: "php", base: "f.php", wantMatch: 1, outOfScope: "html"},
		{
			name: "dockerfile escape directive", lines: []string{"# escape=`"},
			ext: "", base: "dockerfile", wantMatch: 1, outOfScope: "py",
		},
		{
			name: "dockerfile syntax directive", lines: []string{"# syntax=docker/dockerfile:1"},
			ext: "", base: "containerfile", wantMatch: 1, outOfScope: "py",
		},
		{
			name: "go build tag", lines: []string{"//go:build linux", "", "package main"},
			ext: "go", base: "f.go", wantMatch: 2, outOfScope: "js",
		},
		{
			name: "legacy go build tag", lines: []string{"// +build linux", "", "package main"},
			ext: "go", base: "f.go", wantMatch: 2, outOfScope: "js",
		},
		{
			name: "rack directive", lines: []string{`#\ -p 8080`},
			ext: "", base: "config.ru", wantMatch: 1, outOfScope: "rb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var found *marker
			for _, m := range markers {
				if m.appliesTo(tt.ext, tt.base) && m.match(tt.lines) == tt.wantMatch {
					found = &m
					break
				}
			}
			if found == nil {
				t.Fatalf("no built-in marker matched %v for ext=%q base=%q with want=%d",
					tt.lines, tt.ext, tt.base, tt.wantMatch)
			}
			if tt.outOfScope != "" && found.appliesTo(tt.outOfScope, "f."+tt.outOfScope) {
				t.Errorf("marker %v unexpectedly applies to out-of-scope ext %q", found.prefixes, tt.outOfScope)
			}
		})
	}
}
