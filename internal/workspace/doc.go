// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package workspace owns the filesystem layout for the CLI on the user's
// machine. It is the single source of truth for where every command reads and
// writes data — no other package constructs filesystem paths independently.
//
// # Problem
//
// Without a central authority, each package invents its own path:
//
//   - internal/license/template uses os.UserCacheDir()/<cli>/templates
//   - internal/license/config uses ~/.config/<cli>/config.yaml
//   - internal/config/log uses ~/.local/state/<cli>/
//
// These are inconsistent, untestable, and invisible to each other. Adding a
// new command means inventing another path convention with no guarantee it
// follows the same layout.
//
// # Solution
//
// workspace defines two root directories and derives all paths from them:
//
//	User scope  — ~/<EntityFolder>/<AppName>/   (per machine, never committed)
//	Repo scope  — <repo-root>/<AppName>/        (per repository, committed to git)
//
// The entity folder is a directory in the user home directory named after the
// parent entity. It acts as a namespace so all products under the same entity
// share a common root and do not collide with other tools on the machine.
//
// Full layout example:
//
//	~/<EntityFolder>/<AppName>/
//	  license/
//	    templates/     ← remote and SPDX template cache
//	  config.yaml      ← user-level config override
//	  logs/            ← application logs
//
//	<repo-root>/<AppName>/
//	  license/
//	    license.lock   ← template history lock file (committed)
//
// # Usage
//
// Workspace is constructed once in main and stored on app.App. Commands access
// it through the app context — they never construct paths themselves:
//
//	// user-scoped: template cache, rooted at the top-level command directory
//	dir, err := a.Workspace.ForUser(cmd, workspace.DepthCommand).Dir("templates")
//
//	// repo-scoped: lock file, rooted at the top-level command directory
//	file, err := a.Workspace.ForRepo(cmd, workspace.DepthCommand).File("pr.lock")
//
// # Scopes
//
//	a.Workspace.ForUser(cmd, depth)  — reads/writes to the user's machine only
//	a.Workspace.ForRepo(cmd, depth)  — reads/writes to the repository (may be committed)
//
// # Error codes
//
// Workspace operations return errors using the WSP domain:
//
//	WSP001 — home directory could not be resolved at construction time
//
// # Testing
//
// Workspace accepts a home directory override at construction time so tests
// can redirect all paths to a temporary directory without touching the real
// home directory or repository root.
package workspace
