// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import (
	"testing"

	"github.com/leaflock/core-cli/internal/cli/flags"
)

func TestFlag_interface_satisfied_by_CommandFlag(t *testing.T) {
	var _ flags.Flag = flags.CommandFlag[flags.BoolValue]{}
	var _ flags.Flag = flags.CommandFlag[flags.StringValue]{}
	var _ flags.Flag = flags.CommandFlag[flags.StringSliceValue]{}
}

func TestFlag_Definition_returns_pointer(t *testing.T) {
	f := flags.CommandFlag[flags.BoolValue]{Value: flags.Bool("verbose", "enable verbose output")}
	d := f.Definition()
	if d == nil {
		t.Fatal("Definition() must not return nil")
	}
}

func TestFlag_Definition_sets_kind_command(t *testing.T) {
	f := flags.CommandFlag[flags.BoolValue]{Value: flags.Bool("verbose", "enable verbose output")}
	d := f.Definition()
	if d.Meta.Kind != flags.KindCommand {
		t.Errorf("Meta.Kind = %q, want %q", d.Meta.Kind, flags.KindCommand)
	}
}

func TestFlag_Definition_carries_name_and_usage(t *testing.T) {
	f := flags.CommandFlag[flags.StringValue]{Value: flags.String("output", "output path")}
	d := f.Definition()
	if d.Meta.Name != "output" {
		t.Errorf("Meta.Name = %q, want %q", d.Meta.Name, "output")
	}
	if d.Meta.Usage != "output path" {
		t.Errorf("Meta.Usage = %q, want %q", d.Meta.Usage, "output path")
	}
}
