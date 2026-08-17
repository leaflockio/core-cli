// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

// lineOnly builds a Language for a language with only a line form.
func lineOnly(prefix string) Language {
	s := lineStyle{prefix: prefix}
	return Language{Line: s, Default: s}
}

// blockOnly builds a Language for a language with only a block form.
func blockOnly(start, middle, end string) Language {
	s := blockStyle{start: start, middle: middle, end: end}
	return Language{Block: s, Default: s}
}

// lineDefault builds a Language for a language with both forms, defaulting
// to line.
func lineDefault(prefix, start, middle, end string) Language {
	l := lineStyle{prefix: prefix}
	b := blockStyle{start: start, middle: middle, end: end}
	return Language{Line: l, Block: b, Default: l}
}

// blockDefault builds a Language for a language with both forms,
// defaulting to block.
func blockDefault(prefix, start, middle, end string) Language {
	l := lineStyle{prefix: prefix}
	b := blockStyle{start: start, middle: middle, end: end}
	return Language{Line: l, Block: b, Default: b}
}

var (
	CStyle          = lineDefault("//", "/*", "", "*/")
	ShellStyle      = lineOnly("#")
	CSSStyle        = blockOnly("/*", " * ", " */")
	SassStyle       = blockDefault("//", "/*", " * ", " */")
	HTMLStyle       = blockOnly("<!--", "", "-->")
	JavaStyle       = blockDefault("//", "/*", " * ", " */")
	DashStyle       = lineOnly("--")
	LispStyle       = lineOnly(";;")
	ErlangStyle     = lineOnly("%")
	VimStyle        = lineOnly(`"`)
	JinjaStyle      = blockOnly("{#", "", "#}")
	OCamlStyle      = blockOnly("(**", "   ", "*)")
	PowerShellStyle = blockDefault("#", "<#", " ", "#>")
	HandlebarsStyle = blockOnly("{{!--", "  ", "--}}")
)
