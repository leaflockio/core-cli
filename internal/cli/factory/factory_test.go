// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory_test

import (
	"testing"

	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/cli"
	"github.com/leaflock/core-cli/internal/cli/factory"
)

func TestFactory_Build_returns_error_for_nil_command(t *testing.T) {
	_, err := factory.New().Build(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil command, got nil")
	}
}

func TestFactory_Build_returns_cobra_command(t *testing.T) {
	cmd, err := factory.New().Build(&stubCmd{use: "test"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Build must return a non-nil cobra.Command")
	}
	if cmd.Use != "test" {
		t.Errorf("Use = %q, want %q", cmd.Use, "test")
	}
}

func TestFactory_Build_returns_error_when_definition_invalid(t *testing.T) {
	_, err := factory.New().Build(&nilHandlerCmd{}, nil)
	if err == nil {
		t.Fatal("expected error for invalid definition, got nil")
	}
}

type stubCmd struct{ use string }

func (c *stubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    cli.Meta{Use: c.use},
		Handler: func(_ *app.App, _ []string) error { return nil },
	}
}

type nilHandlerCmd struct{}

func (c *nilHandlerCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{Meta: cli.Meta{Use: "bad"}}
}
