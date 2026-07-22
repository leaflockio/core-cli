// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

// Scope identifies which of Workspace's root scopes a KnownPath belongs to.
type Scope int

const (
	// ScopeUser paths live under Workspace's user root (~/<EntityFolder>/<AppName>/).
	ScopeUser Scope = iota
	// ScopeProject paths live under Workspace's repo root (<repo-root>/<AppName>/).
	ScopeProject
	// ScopeCache paths live under Workspace's cache root (<UserCacheDir>/<AppName>/).
	ScopeCache
)

// KnownPath describes a single path the tool resolves.
type KnownPath struct {
	// Name is a stable identifier for this path.
	Name string
	// Path is the resolved filesystem path.
	Path string
	// Desc explains what this path is for.
	Desc string
	// Scope is the root scope this path was resolved under.
	Scope Scope
	// Generated is true when the tool fully owns and writes this path, and
	// false when the path is user-authored and only ever read.
	Generated bool
}
