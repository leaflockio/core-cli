// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package lang

import (
	"path/filepath"
	"testing"
)

// --- Classify ---

func TestClassify_goFile(t *testing.T) {
	if got := Classify(filepath.Join("src", "main.go")); got != "Go" {
		t.Errorf("expected Go, got %q", got)
	}
}

func TestClassify_typescriptFile(t *testing.T) {
	if got := Classify(filepath.Join("src", "app.ts")); got != "TypeScript" {
		t.Errorf("expected TypeScript, got %q", got)
	}
}

func TestClassify_pythonFile(t *testing.T) {
	if got := Classify(filepath.Join("src", "main.py")); got != "Python" {
		t.Errorf("expected Python, got %q", got)
	}
}

func TestClassify_rustFile(t *testing.T) {
	// .rs is shared by RenderScript and Rust; without file content enry cannot
	// make a definitive call. We only verify a Programming-type language is returned.
	if got := Classify(filepath.Join("src", "main.rs")); got == "" {
		t.Error("expected a non-empty language for .rs file")
	}
}

func TestClassify_filenameBasedMakefile(t *testing.T) {
	if got := Classify("Makefile"); got == "" {
		t.Error("expected non-empty language for Makefile")
	}
}

func TestClassify_unknownExtension(t *testing.T) {
	if got := Classify(filepath.Join("repo", "data.bin")); got != "" {
		t.Errorf("expected empty for unknown extension, got %q", got)
	}
}

func TestClassify_vendoredPathSkipped(t *testing.T) {
	if got := Classify(filepath.Join("vendor", "lib", "main.go")); got != "" {
		t.Errorf("expected empty for vendored path, got %q", got)
	}
}

// --- Count ---

func TestCount_empty(t *testing.T) {
	counts := Count(nil)
	if len(counts) != 0 {
		t.Errorf("expected empty counts, got %v", counts)
	}
}

func TestCount_knownExtensions(t *testing.T) {
	files := []string{
		filepath.Join("src", "main.go"),
		filepath.Join("src", "util.go"),
		filepath.Join("src", "app.ts"),
	}
	counts := Count(files)
	if counts["Go"] != 2 {
		t.Errorf("expected Go count=2, got %d", counts["Go"])
	}
	if counts["TypeScript"] != 1 {
		t.Errorf("expected TypeScript count=1, got %d", counts["TypeScript"])
	}
}

func TestCount_unknownExtensionIgnored(t *testing.T) {
	counts := Count([]string{filepath.Join("repo", "data.bin")})
	if len(counts) != 0 {
		t.Errorf("expected no counts for unknown extension, got %v", counts)
	}
}

// --- Build ---

func TestBuild_empty(t *testing.T) {
	det := Build(map[string]int{})
	if det.Primary != "" {
		t.Errorf("expected empty Primary for empty counts, got %q", det.Primary)
	}
	if len(det.All) != 0 {
		t.Errorf("expected empty All, got %v", det.All)
	}
}

func TestBuild_singleLanguage(t *testing.T) {
	det := Build(map[string]int{"Go": 3})
	if det.Primary != "Go" {
		t.Errorf("expected Primary=Go, got %q", det.Primary)
	}
	if len(det.All) != 1 || det.All[0] != "Go" {
		t.Errorf("expected All=[Go], got %v", det.All)
	}
}

func TestBuild_orderedByCount(t *testing.T) {
	// Go: 3, TypeScript: 1, Python: 1 — tie broken alphabetically.
	counts := map[string]int{
		"Go":         3,
		"TypeScript": 1,
		"Python":     1,
	}
	det := Build(counts)
	if det.Primary != "Go" {
		t.Errorf("expected Primary=Go, got %q", det.Primary)
	}
	if len(det.All) != 3 {
		t.Errorf("expected 3 languages, got %v", det.All)
	}
	if det.All[0] != "Go" {
		t.Errorf("expected Go first, got %q", det.All[0])
	}
	// Tie between Python and TypeScript — Python comes first alphabetically.
	if det.All[1] != "Python" {
		t.Errorf("expected Python second (alphabetical tie-break), got %q", det.All[1])
	}
}

func TestBuild_fileCountsPreserved(t *testing.T) {
	counts := map[string]int{"Go": 5, "Rust": 2}
	det := Build(counts)
	if det.FileCounts["Go"] != 5 {
		t.Errorf("expected Go FileCounts=5, got %d", det.FileCounts["Go"])
	}
	if det.FileCounts["Rust"] != 2 {
		t.Errorf("expected Rust FileCounts=2, got %d", det.FileCounts["Rust"])
	}
}
