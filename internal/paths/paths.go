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

// Kind identifies what a KnownPath's content is.
type Kind int

const (
	// KindConfig paths hold user-authored configuration.
	KindConfig Kind = iota
	// KindArtifact paths hold tool-generated, non-disposable content.
	KindArtifact
	// KindCache paths hold tool-generated, disposable content.
	KindCache
)

// Kind label constants, returned by String.
const (
	kindConfigLabel   = "config"
	kindArtifactLabel = "artifact"
	kindCacheLabel    = "cache"
	kindUnknownLabel  = "unknown"
)

// String returns the human-readable label for k.
func (k Kind) String() string {
	switch k {
	case KindConfig:
		return kindConfigLabel
	case KindArtifact:
		return kindArtifactLabel
	case KindCache:
		return kindCacheLabel
	default:
		return kindUnknownLabel
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
	// Kind identifies what this path's content is.
	Kind Kind
}

// Print returns a human-readable representation of p. When brief is true,
// only Path is returned; otherwise Name, Scope, Kind, Path, and Desc are
// included.
func (p KnownPath) Print(brief bool) string {
	if brief {
		return p.Path
	}
	return fmt.Sprintf("%-14s %-8s %-8s %-40s %s", p.Name, p.Scope, p.Kind, p.Path, p.Desc)
}
