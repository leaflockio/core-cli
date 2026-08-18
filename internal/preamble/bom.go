// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package preamble

import "bytes"

// bom is the 3-byte UTF-8 byte-order-mark signature some editors —
// notably on Windows — prepend to a text file.
var bom = []byte{0xEF, 0xBB, 0xBF}

// HasBOM reports whether content begins with a UTF-8 byte-order-mark.
func HasBOM(content []byte) bool {
	return bytes.HasPrefix(content, bom)
}

// StripBOM returns content with a leading UTF-8 byte-order-mark removed,
// if present. Safe to call on content that has none.
func StripBOM(content []byte) []byte {
	return bytes.TrimPrefix(content, bom)
}
