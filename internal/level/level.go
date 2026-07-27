// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package level classifies a command's position in a command tree.
package level

// Level classifies a command's position in the tree. Take a generic
// three-level command line: "tool group action" — tool is the root, group is
// top-level (a direct child of the root), and action is nested (below
// top-level). A grandchild of action is also LevelNested, indistinguishable
// from action itself — Level saturates there instead of counting arbitrarily
// deep.
type Level int

const (
	// LevelRoot is the root of the tree.
	LevelRoot Level = iota

	// LevelTop is a direct child of the root.
	LevelTop

	// LevelNested is anything below top-level, no matter how many levels deep.
	LevelNested
)

const (
	rootName    = "root"
	topName     = "top"
	nestedName  = "nested"
	unknownName = "unknown"
)

// String returns the level's name, for use in error messages.
func (l Level) String() string {
	switch l {
	case LevelRoot:
		return rootName
	case LevelTop:
		return topName
	case LevelNested:
		return nestedName
	default:
		return unknownName
	}
}

// Next returns the level for this node's children.
func (l Level) Next() Level {
	if l == LevelRoot {
		return LevelTop
	}
	return LevelNested
}
