// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/spf13/cobra"
)

// nopHandler is a minimal handler for definitions that require one.
var nopHandler = func(_ *app.App, _ *cobra.Command, _ []string) error { return nil }

// causeOf extracts the Cause text errs.Unexpected attaches — the detail an
// *errs.Error carries separately from its (deliberately generic) Error()
// message.
func causeOf(t *testing.T, err error) string {
	t.Helper()
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error = %v, want an *errs.Error", err)
	}
	if len(e.Contexts) == 0 {
		t.Fatalf("error = %v, want at least one Context", err)
	}
	return e.Contexts[0].Cause
}

// stubConfigLoader is a minimal cmdconfig.ConfigLoader for Definition.Config tests.
type stubConfigLoader struct{}

func (stubConfigLoader) Load(_ map[string]any) error { return nil }
func (stubConfigLoader) Validate() error             { return nil }

// minDef returns a minimal valid Definition for the given command name.
func minDef(use string) *cli.Definition {
	return &cli.Definition{
		Meta:    &cli.Meta{Use: use},
		Handler: nopHandler,
	}
}

// minBlueprint builds a minimal blueprint struct for testing.
func minBlueprint(use, short, long string) *blueprint {
	return &blueprint{
		def: cli.Definition{
			Meta:    &cli.Meta{Use: use, Short: short, Long: long},
			Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
		},
	}
}

// boolFlags returns one flags.Flag per name, for building
// Definition.FlagRules in tests.
func boolFlags(names ...string) []flags.Flag {
	fs := make([]flags.Flag, len(names))
	for i, n := range names {
		fs[i] = flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool(n, "")}
	}
	return fs
}

// mustPlanFlag calls planFlag and panics on error — only for test setup.
func mustPlanFlag(f flags.Flag) flagSpec {
	p, err := planFlag(f)
	if err != nil {
		panic(err)
	}
	return p
}

// unknownFlag satisfies flags.Flag but is not a recognized concrete type.
type unknownFlag struct{}

func (f unknownFlag) Definition() *flags.Definition {
	return &flags.Definition{Meta: flags.Meta{Name: "unknown", Kind: flags.KindCommand}}
}

// stubCommand is a minimal cli.Command for child registration tests.
type stubCommand struct{ use string }

func (c *stubCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    &cli.Meta{Use: c.use},
		Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
	}
}

// groupedStubCommand is a minimal cli.Command whose Group is settable, for
// group-validation tests.
type groupedStubCommand struct {
	use   string
	group cli.GroupID
}

func (c *groupedStubCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    &cli.Meta{Use: c.use},
		Group:   c.group,
		Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
	}
}

// stubBoolResolver records calls to Resolve.
type stubBoolResolver struct {
	called bool
	got    bool
}

func (r *stubBoolResolver) IsResolver() {}
func (r *stubBoolResolver) Resolve(raw bool) error {
	r.called = true
	r.got = raw
	return nil
}

// stubStringResolver records calls to Resolve.
type stubStringResolver struct {
	called bool
	got    string
}

func (r *stubStringResolver) IsResolver() {}
func (r *stubStringResolver) Resolve(raw string) error {
	r.called = true
	r.got = raw
	return nil
}

// stubStringSliceResolver records calls to Resolve.
type stubStringSliceResolver struct {
	called bool
	got    []string
}

func (r *stubStringSliceResolver) IsResolver() {}
func (r *stubStringSliceResolver) Resolve(raw []string) error {
	r.called = true
	r.got = raw
	return nil
}

// nilHandlerCommand returns a definition with no handler — used to trigger build errors.
type nilHandlerCommand struct{}

func (c *nilHandlerCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{Meta: &cli.Meta{Use: "bad"}}
}

// errStringSliceResolver always returns an error from Resolve.
type errStringSliceResolver struct{ err error }

func (r *errStringSliceResolver) IsResolver()              {}
func (r *errStringSliceResolver) Resolve(_ []string) error { return r.err }
