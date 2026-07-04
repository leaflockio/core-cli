// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// CommandFlag is a typed flag whose value is consumed directly by its command.
type CommandFlag[V ValueType] struct {
	Value V
}

// Definition returns the complete declaration of this flag.
func (f CommandFlag[V]) Definition() *Definition {
	m := f.Value.meta()
	m.Kind = KindCommand
	return &Definition{Meta: m}
}
