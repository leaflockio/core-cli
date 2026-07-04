// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

import "github.com/leaflock/core-cli/internal/app"

// SystemFlag is a typed flag with a side effect applied at the infrastructure level.
type SystemFlag[V ValueType] struct {
	Sub    FlagSubcategory
	Value  V
	Effect func(*app.App)
}

// Definition returns the complete declaration of this flag.
func (f SystemFlag[V]) Definition() *Definition {
	m := f.Value.meta()
	m.Kind = KindSystem
	m.Sub = f.Sub
	return &Definition{Meta: m}
}
