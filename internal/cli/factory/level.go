// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

// level classifies a command's position in the tree. Take a generic
// three-level command line: "tool group action" — tool is the root, group is
// top-level (a direct child of the root), and action is nested (below
// top-level). Validation rules only ever care about which of these three
// buckets a command falls into, never its exact nesting depth — a
// grandchild of action is also levelNested, indistinguishable from action
// itself — so level saturates there instead of counting arbitrarily deep.
type level int

const (
	// Root is the single command passed directly to Factory.Build.
	levelRoot level = iota

	// Top is a direct child of the root.
	levelTop

	// Nested is anything below top-level, no matter how many levels deep.
	levelNested
)

const (
	levelRootName    = "root"
	levelTopName     = "top"
	levelNestedName  = "nested"
	levelUnknownName = "unknown"
)

// String returns the level's name, for use in error messages.
func (l level) String() string {
	switch l {
	case levelRoot:
		return levelRootName
	case levelTop:
		return levelTopName
	case levelNested:
		return levelNestedName
	default:
		return levelUnknownName
	}
}

// next returns the level for this node's children.
func (l level) next() level {
	if l == levelRoot {
		return levelTop
	}
	return levelNested
}
