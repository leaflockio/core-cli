// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package paths

import (
	"errors"
	"fmt"

	"github.com/leaflockio/core-cli/internal/errs"
)

// errDuplicateName is the static base error wrapped when Add finds a
// duplicate Name.
var errDuplicateName = errors.New("paths: duplicate KnownPath name")

// Registry accumulates KnownPath entries, rejecting a duplicate Name at
// registration time.
type Registry struct {
	paths []KnownPath
	names map[string]bool
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{names: make(map[string]bool)}
}

// Add registers each of paths. Returns an error on the first duplicate
// Name found.
func (r *Registry) Add(paths ...KnownPath) error {
	for _, p := range paths {
		if r.names[p.Name] {
			return errs.Unexpected(fmt.Errorf("%w: %q", errDuplicateName, p.Name))
		}
		r.names[p.Name] = true
		r.paths = append(r.paths, p)
	}
	return nil
}

// All returns every path registered so far.
func (r *Registry) All() []KnownPath {
	return r.paths
}
