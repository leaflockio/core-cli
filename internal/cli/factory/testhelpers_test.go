// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/cli"
	"github.com/leaflock/core-cli/internal/cli/flags"
)

// nopHandler is a minimal handler for definitions that require one.
var nopHandler = func(_ *app.App, _ []string) error { return nil }

// minDef returns a minimal valid Definition for the given command name.
func minDef(use string) *cli.Definition {
	return &cli.Definition{
		Meta:    cli.Meta{Use: use},
		Handler: nopHandler,
	}
}

// assembledPlan builds a minimal assembled struct for testing.
func assembledPlan(use, short, long string) *assembled {
	return &assembled{
		def: cli.Definition{
			Meta:    cli.Meta{Use: use, Short: short, Long: long},
			Handler: func(_ *app.App, _ []string) error { return nil },
		},
	}
}

// mustPlanFlag calls planFlag and panics on error — only for test setup.
func mustPlanFlag(f flags.Flag) assembledFlag {
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
		Meta:    cli.Meta{Use: c.use},
		Handler: func(_ *app.App, _ []string) error { return nil },
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
	return &cli.Definition{Meta: cli.Meta{Use: "bad"}}
}

// errStringSliceResolver always returns an error from Resolve.
type errStringSliceResolver struct{ err error }

func (r *errStringSliceResolver) IsResolver()              {}
func (r *errStringSliceResolver) Resolve(_ []string) error { return r.err }
