// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/cli"
)

func TestGroupProject_value(t *testing.T) {
	if cli.GroupProject != "project" {
		t.Errorf("GroupProject = %q, want %q", cli.GroupProject, "project")
	}
}

func TestGroupAccount_value(t *testing.T) {
	if cli.GroupAccount != "account" {
		t.Errorf("GroupAccount = %q, want %q", cli.GroupAccount, "account")
	}
}

func TestGroups_containsProjectAndAccount(t *testing.T) {
	want := []cli.Group{
		{ID: cli.GroupProject, Title: "project"},
		{ID: cli.GroupAccount, Title: "account"},
	}
	if len(cli.Groups) != len(want) {
		t.Fatalf("Groups = %v, want %v", cli.Groups, want)
	}
	for i, g := range want {
		if cli.Groups[i] != g {
			t.Errorf("Groups[%d] = %v, want %v", i, cli.Groups[i], g)
		}
	}
}
