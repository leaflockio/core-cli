// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli

// GroupID identifies a command display group.
type GroupID string

// Command group IDs.
const (
	GroupProject GroupID = "project"
	GroupAccount GroupID = "account"
	GroupCLI     GroupID = "cli"
)

// Groups defines the display order of command groups in the help output.
var Groups = []GroupID{GroupProject, GroupAccount, GroupCLI}
