// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package builtin embeds the built-in license header templates and exposes
// lookup and classification helpers for the template loader.
package builtin

import (
	_ "embed"
	"strings"
)

//go:embed templates/proprietary.txt
var proprietary string

//go:embed templates/open-source.txt
var openSource string

// canonicalNames is the ordered list of canonical built-in template names
// shown in error messages and help output. Aliases are not included.
var canonicalNames = []string{"proprietary", "open-source"}

// templates maps every accepted name (canonical + alias) to its content.
var templates = map[string]string{
	"proprietary": proprietary,
	"private":     proprietary, // alias for proprietary
	"open-source": openSource,
	"public":      openSource, // alias for open-source
}

// projects. Format: spdx is invalid with any of these.
var privateBuiltins = map[string]struct{}{
	"proprietary": {},
	"private":     {},
}

// IsPrivate reports whether name is a built-in private/proprietary template.
func IsPrivate(name string) bool {
	_, ok := privateBuiltins[strings.ToLower(name)]
	return ok
}

// Lookup returns the embedded template content for name, and whether it was
// found. Name is matched case-insensitively.
func Lookup(name string) (string, bool) {
	content, ok := templates[strings.ToLower(name)]
	return content, ok
}

// Names returns the canonical built-in template names (aliases excluded).
// Used in error messages and help output.
func Names() []string {
	return canonicalNames
}
