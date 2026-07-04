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

func TestCommandFlag_satisfies_Flag_interface(t *testing.T) {
	var _ flags.Flag = flags.CommandFlag[*flags.BoolValue]{}
	var _ flags.Flag = flags.CommandFlag[*flags.StringValue]{}
	var _ flags.Flag = flags.CommandFlag[*flags.StringSliceValue]{}
}

func TestCommandFlag_Definition_not_nil(t *testing.T) {
	f := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "enable verbose output")}
	if f.Definition() == nil {
		t.Fatal("Definition() must not return nil")
	}
}

func TestCommandFlag_Definition_kind_is_command(t *testing.T) {
	f := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "enable verbose output")}
	if got := f.Definition().Meta.Kind; got != flags.KindCommand {
		t.Errorf("Meta.Kind = %q, want %q", got, flags.KindCommand)
	}
}

func TestCommandFlag_Definition_carries_name_and_usage(t *testing.T) {
	f := flags.CommandFlag[*flags.StringValue]{Value: flags.String("output", "output path")}
	d := f.Definition()
	if d.Meta.Name != "output" {
		t.Errorf("Meta.Name = %q, want %q", d.Meta.Name, "output")
	}
	if d.Meta.Usage != "output path" {
		t.Errorf("Meta.Usage = %q, want %q", d.Meta.Usage, "output path")
	}
}

func TestCommandFlag_Definition_carries_shorthand(t *testing.T) {
	f := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "").WithShorthand("v")}
	if got := f.Definition().Meta.Shorthand; got != "v" {
		t.Errorf("Meta.Shorthand = %q, want %q", got, "v")
	}
}

func TestCommandFlag_Definition_sub_is_empty(t *testing.T) {
	f := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "")}
	if got := f.Definition().Meta.Sub; got != "" {
		t.Errorf("Meta.Sub = %q, want empty", got)
	}
}
