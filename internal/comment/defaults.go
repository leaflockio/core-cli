// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import "path/filepath"

// Language records every comment form a language supports.
type Language struct {
	Line, Block Style
	Default     Style
}

var languages = map[string]Language{
	"go":     CStyle,
	"ts":     CStyle,
	"tsx":    CStyle,
	"cts":    CStyle,
	"mts":    CStyle,
	"js":     CStyle,
	"jsx":    CStyle,
	"mjs":    CStyle,
	"cjs":    CStyle,
	"swift":  CStyle,
	"rs":     CStyle,
	"cpp":    CStyle,
	"cc":     CStyle,
	"cxx":    CStyle,
	"hpp":    CStyle,
	"cs":     CStyle,
	"proto":  CStyle,
	"dart":   CStyle,
	"sc":     CStyle,
	"groovy": CStyle,
	"jsonc":  CStyle,
	"json5":  CStyle,
	"hcl":    CStyle,
	"m":      CStyle,
	"mm":     CStyle,
	"v":      CStyle,
	"sv":     CStyle,

	"java":   JavaStyle,
	"kt":     JavaStyle,
	"kts":    JavaStyle,
	"scala":  JavaStyle,
	"c":      JavaStyle,
	"h":      JavaStyle,
	"gv":     JavaStyle,
	"thrift": JavaStyle,
	"php":    JavaStyle,

	"sh":     ShellStyle,
	"bash":   ShellStyle,
	"zsh":    ShellStyle,
	"fish":   ShellStyle,
	"py":     ShellStyle,
	"pyw":    ShellStyle,
	"rb":     ShellStyle,
	"tf":     ShellStyle,
	"tfvars": ShellStyle,
	"r":      ShellStyle,
	"ex":     ShellStyle,
	"exs":    ShellStyle,
	"toml":   ShellStyle,
	"yaml":   ShellStyle,
	"yml":    ShellStyle,
	"nix":    ShellStyle,
	"pl":     ShellStyle,
	"pp":     ShellStyle,
	"tcl":    ShellStyle,
	"mk":     ShellStyle,
	"bzl":    ShellStyle,
	"bazel":  ShellStyle,

	"css": CSSStyle,

	"scss": SassStyle,
	"sass": SassStyle,
	"less": SassStyle,

	"html":     HTMLStyle,
	"htm":      HTMLStyle,
	"vue":      HTMLStyle,
	"svelte":   HTMLStyle,
	"xml":      HTMLStyle,
	"markdown": HTMLStyle,
	"md":       HTMLStyle,

	"hs":  DashStyle,
	"sql": DashStyle,
	"sdl": DashStyle,
	"lua": DashStyle,

	"el":   LispStyle,
	"lisp": LispStyle,
	"scm":  LispStyle,
	"clj":  LispStyle,
	"cljs": LispStyle,
	"cljc": LispStyle,

	"erl": ErlangStyle,

	"vim": VimStyle,

	"j2":     JinjaStyle,
	"jinja":  JinjaStyle,
	"jinja2": JinjaStyle,
	"twig":   JinjaStyle,

	"ml":  OCamlStyle,
	"mli": OCamlStyle,
	"mll": OCamlStyle,
	"mly": OCamlStyle,

	"ps1":  PowerShellStyle,
	"psm1": PowerShellStyle,

	"hbs":        HandlebarsStyle,
	"handlebars": HandlebarsStyle,
}

// bareNames maps a lowercase, extension-less filename (or dotfile name
// without its leading dot) to its known comment Language — for files
// conventionally named without an extension, like Makefile and
// Dockerfile, that filepath.Ext can never match.
var bareNames = map[string]Language{
	"makefile":      ShellStyle,
	"dockerfile":    ShellStyle,
	"containerfile": ShellStyle,
	"build":         ShellStyle,
	"buckconfig":    ShellStyle,
	"gemfile":       ShellStyle,
}

// LanguageForExtension returns the known Language for ext (with or
// without a leading dot, case-insensitive), and whether one was found.
func LanguageForExtension(ext string) (Language, bool) {
	l, ok := languages[normalizeKey(ext)]
	return l, ok
}

// LanguageForFile returns the known Language for path, and whether one
// was found. It checks path's base filename against bareNames first —
// for a file conventionally named without an extension, like Makefile —
// before falling back to LanguageForExtension.
func LanguageForFile(path string) (Language, bool) {
	base := normalizeKey(filepath.Base(path))
	if l, ok := bareNames[base]; ok {
		return l, true
	}
	return LanguageForExtension(filepath.Ext(path))
}

// ForExtension returns the default Style for ext (with or without a
// leading dot, case-insensitive), and whether one was found.
func ForExtension(ext string) (Style, bool) {
	l, ok := LanguageForExtension(ext)
	if !ok {
		return nil, false
	}
	return l.Default, true
}

// ForFile returns the default Style for path, and whether one was found.
func ForFile(path string) (Style, bool) {
	l, ok := LanguageForFile(path)
	if !ok {
		return nil, false
	}
	return l.Default, true
}
