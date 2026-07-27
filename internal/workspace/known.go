// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package workspace

import (
	"path/filepath"

	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/paths"
)

// KnownPaths is the static set of paths a Workspace resolves — paths that
// exist exactly once, independent of any specific command.
type KnownPaths struct {
	Manifest    paths.KnownPath
	UserConfig  paths.KnownPath
	Credentials paths.KnownPath
}

// All returns every path in kp as a flat slice. Returns an error if two
// entries share the same Name.
func (kp *KnownPaths) All() ([]paths.KnownPath, error) {
	r := paths.NewRegistry()
	if err := r.Add(kp.Manifest, kp.UserConfig, kp.Credentials); err != nil {
		return nil, err
	}
	return r.All(), nil
}

// Known returns the static set of paths this Workspace resolves.
func (w *Workspace) Known() KnownPaths {
	return KnownPaths{
		Manifest: paths.KnownPath{
			Name:  "manifest",
			Path:  filepath.Join(w.repoRoot, config.ManifestFile),
			Desc:  "combined config for every command",
			Scope: paths.ScopeProject,
		},
		UserConfig: paths.KnownPath{
			Name:  "user-config",
			Path:  filepath.Join(w.userRoot, config.UserConfigFile),
			Desc:  "user-level config override",
			Scope: paths.ScopeUser,
		},
		Credentials: paths.KnownPath{
			Name:      "credentials",
			Path:      w.CredentialsPath(),
			Desc:      "stored credentials",
			Scope:     paths.ScopeUser,
			Generated: true,
		},
	}
}
