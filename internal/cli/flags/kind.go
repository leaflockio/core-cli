// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// FlagKind identifies who owns a flag.
type FlagKind string

// FlagSubcategory narrows the behavior within a kind.
type FlagSubcategory string

// System flag kind and its subcategories.
const (
	KindSystem  FlagKind        = "system"
	SubImplicit FlagSubcategory = "implicit" // factory registers automatically on every command
	SubExplicit FlagSubcategory = "explicit" // developer opts the command in
)

// Command flag kind. Behavior is driven by the Resolver interface, not subcategory.
const (
	KindCommand FlagKind = "command"
)
