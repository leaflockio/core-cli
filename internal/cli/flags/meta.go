// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// Meta holds the shared identity of every flag — its kind, subcategory,
// name, optional shorthand, and usage text.
type Meta struct {
	Kind      FlagKind
	Sub       FlagSubcategory
	Name      string
	Shorthand string
	Usage     string
}
