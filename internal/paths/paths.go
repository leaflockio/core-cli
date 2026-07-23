// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package paths defines the shared vocabulary for describing filesystem
// paths the tool cares about.
package paths

import "fmt"

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

// Scope label constants, returned by String.
const (
	scopeUserLabel    = "user"
	scopeProjectLabel = "project"
	scopeCacheLabel   = "cache"
	scopeUnknownLabel = "unknown"
)

// String returns the human-readable label for s.
func (s Scope) String() string {
	switch s {
	case ScopeUser:
		return scopeUserLabel
	case ScopeProject:
		return scopeProjectLabel
	case ScopeCache:
		return scopeCacheLabel
	default:
		return scopeUnknownLabel
	}
}

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

// Print returns a human-readable representation of p. When brief is true,
// only Path is returned; otherwise Name, Scope, Path, and Desc are included.
func (p KnownPath) Print(brief bool) string {
	if brief {
		return p.Path
	}
	return fmt.Sprintf("%-14s %-8s %-40s %s", p.Name, p.Scope, p.Path, p.Desc)
}
