// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// Resolver is the marker interface for flags that handle their own value
// writing after parsing. Any flag kind can implement this.
type Resolver interface {
	IsResolver()
}
