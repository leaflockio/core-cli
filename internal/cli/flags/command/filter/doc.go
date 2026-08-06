// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package filter holds reusable command flags for narrowing a file list by
// glob pattern.
//
// # Include
//
// Include (--include) adds a glob pattern a file must match to be kept.
// Repeatable — a file matching any one of the given patterns qualifies.
//
// # Exclude
//
// Exclude (--exclude) adds a glob pattern that removes a file even if it
// matched Include. Repeatable, and order-sensitive: a "!pattern" entry can
// cancel an earlier exclude match.
package filter
