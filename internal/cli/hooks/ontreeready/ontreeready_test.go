// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package ontreeready_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/hooks/ontreeready"
	"github.com/spf13/cobra"
)

// stubCommand is a plain cli.Command with no hook.
type stubCommand struct {
	children []cli.Command
}

func (s *stubCommand) Define(_ *app.App) *cli.Definition {
	return cli.NewDefinition(cli.NewMeta("stub", "", "")).WithChildren(s.children)
}

// readyCommand additionally implements ontreeready.OnTreeReady.
type readyCommand struct {
	stubCommand
	calledWith *cobra.Command
	callCount  int
}

func (r *readyCommand) OnTreeReady(root *cobra.Command) {
	r.calledWith = root
	r.callCount++
}

var _ ontreeready.OnTreeReady = (*readyCommand)(nil)

func TestNotify_callsOnTreeReady(t *testing.T) {
	root := &cobra.Command{Use: "leaf"}
	r := &readyCommand{}

	ontreeready.Notify(r, nil, root)

	if r.callCount != 1 {
		t.Fatalf("callCount = %d, want 1", r.callCount)
	}
	if r.calledWith != root {
		t.Error("OnTreeReady was not passed the given root")
	}
}

func TestNotify_skipsCommandsNotImplementingHook(t *testing.T) {
	root := &cobra.Command{Use: "leaf"}
	s := &stubCommand{}

	// Must not panic — the whole point of the interface check.
	ontreeready.Notify(s, nil, root)
}

func TestNotify_recursesIntoChildren(t *testing.T) {
	root := &cobra.Command{Use: "leaf"}
	child := &readyCommand{}
	parent := &stubCommand{children: []cli.Command{child}}

	ontreeready.Notify(parent, nil, root)

	if child.callCount != 1 {
		t.Fatalf("child.callCount = %d, want 1", child.callCount)
	}
}

func TestNotify_callsEveryImplementerInTree(t *testing.T) {
	root := &cobra.Command{Use: "leaf"}
	grandchild := &readyCommand{}
	childWithGrandchild := &readyCommand{stubCommand: stubCommand{children: []cli.Command{grandchild}}}
	plainChild := &readyCommand{}
	parent := &stubCommand{children: []cli.Command{childWithGrandchild, plainChild}}

	ontreeready.Notify(parent, nil, root)

	if childWithGrandchild.callCount != 1 {
		t.Error("expected childWithGrandchild to be notified")
	}
	if plainChild.callCount != 1 {
		t.Error("expected plainChild to be notified")
	}
	if grandchild.callCount != 1 {
		t.Error("expected a nested grandchild to be notified too")
	}
}
