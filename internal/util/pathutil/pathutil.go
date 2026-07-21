// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ModuleRootMarker is the filename used to identify a module root.
// Override in tests to avoid depending on a real go.mod.
var ModuleRootMarker = "go.mod"

// FallbackDepth is the number of slash-separated path components returned
// by RelPath when no module root is found.
var FallbackDepth = 2

// RelPath returns file's path relative to the nearest module root.
// Falls back to the last FallbackDepth slash-separated components if no
// module root is found.
func RelPath(file string) string {
	file = filepath.ToSlash(filepath.Clean(file))
	if root, ok := FindModuleRoot(filepath.Dir(file)); ok {
		root = filepath.ToSlash(root)
		return strings.TrimPrefix(file, root+"/")
	}
	parts := strings.Split(file, "/")
	if len(parts) >= FallbackDepth {
		return strings.Join(parts[len(parts)-FallbackDepth:], "/")
	}
	return file
}

// FindModuleRoot walks up from dir until it finds a directory containing
// ModuleRootMarker. Returns the directory and true if found.
func FindModuleRoot(dir string) (string, bool) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ModuleRootMarker)); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
