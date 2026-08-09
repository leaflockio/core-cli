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

func TestNewMeta_args_defaults_to_rejecting_unrecognized_commands(t *testing.T) {
	m := cli.NewMeta("foo", "short", "")
	if m.Args == nil {
		t.Error("Args should default to a validator, got nil")
	}
	cmd := &cobra.Command{}
	if err := m.Args(cmd, []string{"unexpected"}); err == nil {
		t.Error("default Args should reject an unrecognized positional command name")
	}
}

func TestNewMeta_disableSuggestions_defaultsToFalse(t *testing.T) {
	m := cli.NewMeta("foo", "short", "")
	if m.DisableSuggestions {
		t.Error("DisableSuggestions should default to false")
	}
}

func TestNewMeta_argsUsage_defaultsToEmpty(t *testing.T) {
	m := cli.NewMeta("foo", "short", "")
	if m.ArgsUsage != "" {
		t.Errorf("ArgsUsage should default to empty, got %q", m.ArgsUsage)
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

func TestMeta_WithArgsUsage_setsField(t *testing.T) {
	m := cli.NewMeta("check", "short", "").WithArgsUsage("[file...]")
	if m.ArgsUsage != "[file...]" {
		t.Errorf("ArgsUsage = %q, want %q", m.ArgsUsage, "[file...]")
	}
}

func TestMeta_WithArgsUsage_does_not_mutate_original(t *testing.T) {
	original := cli.NewMeta("check", "short", "")
	_ = original.WithArgsUsage("[file...]")
	if original.ArgsUsage != "" {
		t.Errorf("WithArgsUsage mutated the original Meta — should still be empty, got %q", original.ArgsUsage)
	}
}

func TestMeta_WithDisableSuggestions_setsField(t *testing.T) {
	m := cli.NewMeta("foo", "short", "").WithDisableSuggestions()
	if !m.DisableSuggestions {
		t.Error("WithDisableSuggestions should set DisableSuggestions to true")
	}
}

func TestMeta_WithDisableSuggestions_does_not_mutate_original(t *testing.T) {
	original := cli.NewMeta("foo", "short", "")
	_ = original.WithDisableSuggestions()
	if original.DisableSuggestions {
		t.Error("WithDisableSuggestions mutated the original Meta — original should still be false")
	}
}
