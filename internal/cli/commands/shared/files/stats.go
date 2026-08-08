// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package files

// Stats reports how many files survived each narrowing stage of Resolve.
type Stats struct {
	Discovered int // before any narrowing
	Query      int // after positional-arg narrowing; equals Discovered with no args
	Included   int // after --include narrowing
	Excluded   int // Included files dropped by --exclude
	Final      int // what Resolve actually returns
}
