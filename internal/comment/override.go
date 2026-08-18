// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import "strings"

// ApplyOverrides sets extensions and files into the package's known
// extension and bare-filename tables. A key already known to the package
// has its Language replaced; an unknown key is added. Keys are normalized
// the same way LanguageForExtension and LanguageForFile normalize their
// own lookup input (lowercased, leading dot trimmed).
//
// Not safe for concurrent use with lookups (LanguageForExtension,
// LanguageForFile, ForExtension, ForFile) — call it during startup,
// before any lookup can run concurrently with it.
func ApplyOverrides(extensions, files map[string]Language) {
	for ext, lang := range extensions {
		languages[normalizeKey(ext)] = lang
	}
	for name, lang := range files {
		bareNames[normalizeKey(name)] = lang
	}
}

// normalizeKey lowercases s and trims a single leading dot, matching how
// LanguageForExtension and LanguageForFile normalize their own lookup keys.
func normalizeKey(s string) string {
	return strings.ToLower(strings.TrimPrefix(s, "."))
}
