// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package strutil

import "strings"

// Title converts a string to title case by capitalising the first character
// (e.g. "tools" → "Tools"). Returns the input unchanged if it is empty.
func Title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
