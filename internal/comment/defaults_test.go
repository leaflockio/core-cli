// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import "testing"

func TestLanguageForExtension(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		want Language
		ok   bool
	}{
		{name: "known extension without dot", ext: "go", want: CStyle, ok: true},
		{name: "known extension with dot", ext: ".go", want: CStyle, ok: true},
		{name: "uppercase extension", ext: "GO", want: CStyle, ok: true},
		{name: "mixed case extension", ext: ".Py", want: ShellStyle, ok: true},
		{name: "unknown extension", ext: "xyz", ok: false},
		{name: "empty extension", ext: "", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := LanguageForExtension(tt.ext)
			if ok != tt.ok {
				t.Fatalf("LanguageForExtension(%q) ok = %v, want %v", tt.ext, ok, tt.ok)
			}
			if ok && got.Default != tt.want.Default {
				t.Errorf("LanguageForExtension(%q) = %v, want %v", tt.ext, got, tt.want)
			}
		})
	}
}

func TestLanguageForFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want Language
		ok   bool
	}{
		{name: "bare filename, exact case", path: "Makefile", want: ShellStyle, ok: true},
		{name: "bare filename, lowercase", path: "dockerfile", want: ShellStyle, ok: true},
		{name: "bare filename with directory", path: "path/to/Dockerfile", want: ShellStyle, ok: true},
		{name: "bare filename as dotfile", path: ".dockerfile", want: ShellStyle, ok: true},
		{name: "extension fallback", path: "main.go", want: CStyle, ok: true},
		{name: "extension fallback with directory", path: "path/to/main.go", want: CStyle, ok: true},
		{name: "uppercase extension fallback", path: "MAIN.GO", want: CStyle, ok: true},
		{name: "bare name takes priority over extension lookup", path: "Gemfile", want: ShellStyle, ok: true},
		{name: "unknown file", path: "file.unknown", ok: false},
		{name: "no extension, not a known bare name", path: "README", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := LanguageForFile(tt.path)
			if ok != tt.ok {
				t.Fatalf("LanguageForFile(%q) ok = %v, want %v", tt.path, ok, tt.ok)
			}
			if ok && got.Default != tt.want.Default {
				t.Errorf("LanguageForFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestForExtension(t *testing.T) {
	s, ok := ForExtension(".go")
	if !ok {
		t.Fatal("ForExtension(.go) ok = false, want true")
	}
	if s != CStyle.Default {
		t.Errorf("ForExtension(.go) = %v, want CStyle.Default", s)
	}

	if _, ok := ForExtension("nope"); ok {
		t.Error("ForExtension(nope) ok = true, want false")
	}
}

func TestForFile(t *testing.T) {
	s, ok := ForFile("Makefile")
	if !ok {
		t.Fatal("ForFile(Makefile) ok = false, want true")
	}
	if s != ShellStyle.Default {
		t.Errorf("ForFile(Makefile) = %v, want ShellStyle.Default", s)
	}

	if _, ok := ForFile("file.unknown"); ok {
		t.Error("ForFile(file.unknown) ok = true, want false")
	}
}
