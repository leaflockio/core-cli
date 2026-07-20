// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
)

func TestDefinition_zero_value(t *testing.T) {
	var d flags.Definition
	if d.Meta.Name != "" || d.Meta.Kind != "" {
		t.Error("zero Definition should have zero Meta")
	}
}

func TestDefinition_holds_meta(t *testing.T) {
	d := flags.Definition{
		Meta: flags.Meta{
			Kind:      flags.KindCommand,
			Name:      "output",
			Shorthand: "o",
			Usage:     "output path",
		},
	}
	if d.Meta.Kind != flags.KindCommand {
		t.Errorf("Meta.Kind = %q, want %q", d.Meta.Kind, flags.KindCommand)
	}
	if d.Meta.Name != "output" {
		t.Errorf("Meta.Name = %q, want %q", d.Meta.Name, "output")
	}
	if d.Meta.Shorthand != "o" {
		t.Errorf("Meta.Shorthand = %q, want %q", d.Meta.Shorthand, "o")
	}
	if d.Meta.Usage != "output path" {
		t.Errorf("Meta.Usage = %q, want %q", d.Meta.Usage, "output path")
	}
}
