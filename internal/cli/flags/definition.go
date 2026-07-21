// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// Definition is the complete declaration of a flag — its identity and metadata.
type Definition struct {
	// Meta holds the flag's identity — its kind, subcategory, name, shorthand, and usage.
	Meta Meta
}
