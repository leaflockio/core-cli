// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package comment describes and applies comment syntax for source files.
package comment

// Style renders plain text lines as a comment.
type Style interface {
	wrap(lines []string) []string
}

// Wrap renders lines of plain text into style's comment syntax.
func Wrap(lines []string, style Style) []string {
	return style.wrap(lines)
}
