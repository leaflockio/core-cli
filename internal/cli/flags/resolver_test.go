// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import (
	"testing"

	"github.com/leaflock/core-cli/internal/cli/flags"
)

type stubResolver struct{}

func (s stubResolver) IsResolver() {}

func TestResolver_satisfied_by_implementation(t *testing.T) {
	var _ flags.Resolver = stubResolver{}
}

func TestResolver_IsResolver_callable(t *testing.T) {
	var r flags.Resolver = stubResolver{}
	r.IsResolver()
}
