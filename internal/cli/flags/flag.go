// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// Flag is the interface implemented by all concrete flag kinds.
// Definition returns the complete declaration of the flag — its identity,
// kind, subcategory, name, shorthand, and usage text.
type Flag interface {
	Definition() *Definition
}

// ValueType constrains the set of value descriptors that can be held by a flag.
// All supported value types must be added to this union.
type ValueType interface {
	*BoolValue | *StringValue | *StringSliceValue
	meta() Meta
	Validate() error
}
