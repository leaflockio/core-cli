// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package lang defines the Language enum and extension-to-language mapping
// used across leaf for language detection. All language comparisons in the
// codebase must use this package's typed constants — no raw string comparisons.
package lang

import "strings"

// Language identifies a programming or markup language.
// Use the typed constants below; never compare against raw strings.
type Language int

const (
	Unknown    Language = iota
	Go                  // .go
	TypeScript          // .ts, .tsx
	JavaScript          // .js, .jsx, .mjs, .cjs
	Python              // .py, .pyw
	Shell               // .sh, .bash, .zsh, .fish
	Rust                // .rs
	Java                // .java
	Kotlin              // .kt, .kts
	Swift               // .swift
	Ruby                // .rb
	CSS                 // .css, .scss, .sass, .less
	HTML                // .html, .htm
	Proto               // .proto
	Terraform           // .tf, .tfvars
	CPP                 // .cpp, .cc, .cxx, .hpp
	C                   // .c, .h
	CSharp              // .cs
	Vue                 // .vue
	Svelte              // .svelte
	Scala               // .scala, .sc
	Elixir              // .ex, .exs
	Haskell             // .hs, .lhs
	Lua                 // .lua
	Dart                // .dart
)

// names maps each Language to its canonical display string.
var names = map[Language]string{
	Unknown:    "unknown",
	Go:         "go",
	TypeScript: "typescript",
	JavaScript: "javascript",
	Python:     "python",
	Shell:      "shell",
	Rust:       "rust",
	Java:       "java",
	Kotlin:     "kotlin",
	Swift:      "swift",
	Ruby:       "ruby",
	CSS:        "css",
	HTML:       "html",
	Proto:      "proto",
	Terraform:  "terraform",
	CPP:        "cpp",
	C:          "c",
	CSharp:     "csharp",
	Vue:        "vue",
	Svelte:     "svelte",
	Scala:      "scala",
	Elixir:     "elixir",
	Haskell:    "haskell",
	Lua:        "lua",
	Dart:       "dart",
}

// String returns the canonical display name for the language.
func (l Language) String() string {
	if s, ok := names[l]; ok {
		return s
	}
	return names[Unknown]
}

// extLanguage maps lowercase file extensions (without leading dot) to Language.
// When multiple extensions map to the same language, all are listed here.
var extLanguage = map[string]Language{
	// Go
	"go": Go,
	// TypeScript
	"ts":  TypeScript,
	"tsx": TypeScript,
	// JavaScript
	"js":  JavaScript,
	"jsx": JavaScript,
	"mjs": JavaScript,
	"cjs": JavaScript,
	// Python
	"py":  Python,
	"pyw": Python,
	// Shell
	"sh":   Shell,
	"bash": Shell,
	"zsh":  Shell,
	"fish": Shell,
	// Rust
	"rs": Rust,
	// Java
	"java": Java,
	// Kotlin
	"kt":  Kotlin,
	"kts": Kotlin,
	// Swift
	"swift": Swift,
	// Ruby
	"rb": Ruby,
	// CSS
	"css":  CSS,
	"scss": CSS,
	"sass": CSS,
	"less": CSS,
	// HTML
	"html": HTML,
	"htm":  HTML,
	// Proto
	"proto": Proto,
	// Terraform
	"tf":     Terraform,
	"tfvars": Terraform,
	// C++
	"cpp": CPP,
	"cc":  CPP,
	"cxx": CPP,
	"hpp": CPP,
	// C
	"c": C,
	"h": C,
	// C#
	"cs": CSharp,
	// Vue
	"vue": Vue,
	// Svelte
	"svelte": Svelte,
	// Scala
	"scala": Scala,
	"sc":    Scala,
	// Elixir
	"ex":  Elixir,
	"exs": Elixir,
	// Haskell
	"hs":  Haskell,
	"lhs": Haskell,
	// Lua
	"lua": Lua,
	// Dart
	"dart": Dart,
}

// FromExtension returns the Language for the given file extension.
// The extension may include a leading dot (e.g. ".go") or omit it (e.g. "go").
// Lookup is case-insensitive. Returns Unknown if the extension is not recognized.
func FromExtension(ext string) Language {
	if ext == "" {
		return Unknown
	}
	if ext[0] == '.' {
		ext = ext[1:]
	}
	if l, ok := extLanguage[strings.ToLower(ext)]; ok {
		return l
	}
	return Unknown
}

// Detection is the result of scanning a repository for programming languages.
type Detection struct {
	// Primary is the language with the highest file count, or Unknown if no
	// recognized source files were found.
	Primary Language
	// All lists every detected language ordered by file count descending.
	All []Language
	// FileCounts maps each detected language to its file count.
	FileCounts map[Language]int
}
