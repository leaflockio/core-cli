// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"fmt"
	"os"
	"time"
)

// printDuration reports how long the invocation took. Shows both a
// rounded, readable form and the exact duration.
func printDuration(d time.Duration) {
	fmt.Fprintf(os.Stderr, "\ndone in %s (%s)\n", roundForDisplay(d), d)
}

// roundForDisplay rounds d to a granularity proportional to its own
// magnitude, so it always prints with roughly 2-3 significant digits
// regardless of scale.
func roundForDisplay(d time.Duration) time.Duration {
	switch {
	case d < time.Millisecond:
		return d.Round(time.Microsecond)
	case d < time.Second:
		return d.Round(time.Millisecond)
	default:
		return d.Round(10 * time.Millisecond)
	}
}
