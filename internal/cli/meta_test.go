// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli_test

import (
	"testing"

	"github.com/leaflock/core-cli/internal/cli"
	"github.com/spf13/cobra"
)

func TestNewMeta_sets_fields(t *testing.T) {
	m := cli.NewMeta("foo", "short desc", "long desc")
	if m.Use != "foo" {
		t.Errorf("Use = %q, want %q", m.Use, "foo")
	}
	if m.Short != "short desc" {
		t.Errorf("Short = %q, want %q", m.Short, "short desc")
	}
	if m.Long != "long desc" {
		t.Errorf("Long = %q, want %q", m.Long, "long desc")
	}
}

func TestNewMeta_args_defaults_to_no_args(t *testing.T) {
	m := cli.NewMeta("foo", "short", "")
	if m.Args == nil {
		t.Error("Args should default to cobra.NoArgs, got nil")
	}
	cmd := &cobra.Command{}
	if err := m.Args(cmd, []string{"unexpected"}); err == nil {
		t.Error("default Args should reject positional arguments")
	}
}

func TestMeta_WithArgs_overrides_default(t *testing.T) {
	m := cli.NewMeta("foo", "short", "").WithArgs(cobra.ArbitraryArgs)
	cmd := &cobra.Command{}
	if err := m.Args(cmd, []string{"any", "args"}); err != nil {
		t.Errorf("WithArgs: ArbitraryArgs should accept arguments, got: %v", err)
	}
}

func TestMeta_WithArgs_does_not_mutate_original(t *testing.T) {
	original := cli.NewMeta("foo", "short", "")
	_ = original.WithArgs(cobra.ArbitraryArgs)
	cmd := &cobra.Command{}
	if err := original.Args(cmd, []string{"unexpected"}); err == nil {
		t.Error("WithArgs mutated the original Meta — original should still reject arguments")
	}
}
