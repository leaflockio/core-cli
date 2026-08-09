// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"testing"

	"github.com/spf13/cobra"
)

// stubTreeReady additionally implements ontreeready.OnTreeReady, so
// discoverHooks/discoverTreeReady/fireTreeReady can be exercised directly.
type stubTreeReady struct {
	stubCommand
	gotSelf   *cobra.Command
	gotRoot   *cobra.Command
	callCount int
}

func (s *stubTreeReady) OnTreeReady(self, root *cobra.Command) {
	s.gotSelf = self
	s.gotRoot = root
	s.callCount++
}

// — discoverHooks —

func TestDiscoverHooks_appendsTreeReadyImplementer(t *testing.T) {
	self := &cobra.Command{Use: "child"}
	cmd := &stubTreeReady{stubCommand: stubCommand{use: "child"}}
	var hooks []hookRecord

	discoverHooks(cmd, self, &hooks)

	if len(hooks) != 1 {
		t.Fatalf("len(hooks) = %d, want 1", len(hooks))
	}
	if hooks[0].self != self {
		t.Error("hookRecord.self must be the given self")
	}
	if hooks[0].treeReady == nil {
		t.Error("hookRecord.treeReady must be set")
	}
}

func TestDiscoverHooks_skipsCommandNotImplementingAnyHook(t *testing.T) {
	self := &cobra.Command{Use: "child"}
	cmd := &stubCommand{use: "child"}
	var hooks []hookRecord

	discoverHooks(cmd, self, &hooks)

	if len(hooks) != 0 {
		t.Errorf("len(hooks) = %d, want 0", len(hooks))
	}
}

// — discoverTreeReady —

func TestDiscoverTreeReady_pairsSelfWithHook(t *testing.T) {
	self := &cobra.Command{Use: "child"}
	cmd := &stubTreeReady{stubCommand: stubCommand{use: "child"}}
	var hooks []hookRecord

	discoverTreeReady(cmd, self, &hooks)

	if len(hooks) != 1 || hooks[0].self != self {
		t.Fatalf("expected one hookRecord paired with self, got %+v", hooks)
	}
}

func TestDiscoverTreeReady_noAppendWhenNotImplemented(t *testing.T) {
	self := &cobra.Command{Use: "child"}
	cmd := &stubCommand{use: "child"}
	var hooks []hookRecord

	discoverTreeReady(cmd, self, &hooks)

	if len(hooks) != 0 {
		t.Errorf("len(hooks) = %d, want 0", len(hooks))
	}
}

// — fireTreeReady —

func TestFireTreeReady_callsOnTreeReadyOnEachRecord(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	self := &cobra.Command{Use: "child"}
	cmd := &stubTreeReady{stubCommand: stubCommand{use: "child"}}

	fireTreeReady([]hookRecord{{self: self, treeReady: cmd}}, root)

	if cmd.callCount != 1 {
		t.Fatalf("callCount = %d, want 1", cmd.callCount)
	}
	if cmd.gotSelf != self {
		t.Error("expected self to be passed through")
	}
	if cmd.gotRoot != root {
		t.Error("expected root to be passed through")
	}
}

func TestFireTreeReady_skipsRecordWithNoTreeReady(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	// Must not panic on a record with no treeReady set.
	fireTreeReady([]hookRecord{{self: &cobra.Command{}}}, root)
}

func TestFireTreeReady_callsEveryRecord(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	a := &stubTreeReady{stubCommand: stubCommand{use: "a"}}
	b := &stubTreeReady{stubCommand: stubCommand{use: "b"}}

	fireTreeReady([]hookRecord{
		{self: &cobra.Command{Use: "a"}, treeReady: a},
		{self: &cobra.Command{Use: "b"}, treeReady: b},
	}, root)

	if a.callCount != 1 {
		t.Error("expected a to be fired")
	}
	if b.callCount != 1 {
		t.Error("expected b to be fired")
	}
}
