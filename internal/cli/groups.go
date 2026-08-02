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
)

// Group pairs a GroupID with its display title.
type Group struct {
	ID    GroupID
	Title string
}

// Groups is the reusable group set, in display order.
var Groups = []Group{
	{ID: GroupProject, Title: "project"},
	{ID: GroupAccount, Title: "account"},
}
