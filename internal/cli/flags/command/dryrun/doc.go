// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package dryrun declares the canonical --dry-run command flag.
//
// It only belongs on a command that actually writes to disk — resolving
// files and reporting what would be written, without writing it.
package dryrun
