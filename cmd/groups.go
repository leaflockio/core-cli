// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

// Command group IDs.
const (
	groupGeneral = "general"
	groupTools   = "tools"
)

// groups defines the display order of command groups in the help output.
var groups = []string{groupTools, groupGeneral}
