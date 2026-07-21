// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"testing"

	"github.com/spf13/pflag"
)

// — boolRegistrar —

func TestBoolRegistrar_no_shorthand_no_dest(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	boolRegistrar("verbose", "", "enable verbose output", true, nil)(fs)

	f := fs.Lookup("verbose")
	if f == nil {
		t.Fatal("flag 'verbose' not registered")
	}
	if f.Usage != "enable verbose output" {
		t.Errorf("Usage = %q, want %q", f.Usage, "enable verbose output")
	}
	if f.DefValue != "true" {
		t.Errorf("DefValue = %q, want %q", f.DefValue, "true")
	}
}

func TestBoolRegistrar_with_shorthand(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	boolRegistrar("verbose", "v", "enable verbose output", false, nil)(fs)

	if fs.ShorthandLookup("v") == nil {
		t.Error("shorthand 'v' not registered")
	}
}

func TestBoolRegistrar_with_dest(t *testing.T) {
	var dest bool
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	boolRegistrar("verbose", "", "enable verbose output", false, &dest)(fs)

	if err := fs.Parse([]string{"--verbose"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !dest {
		t.Error("dest must be true after flag is set")
	}
}

func TestBoolRegistrar_with_shorthand_and_dest(t *testing.T) {
	var dest bool
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	boolRegistrar("verbose", "v", "enable verbose output", false, &dest)(fs)

	if err := fs.Parse([]string{"-v"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !dest {
		t.Error("dest must be true after shorthand is set")
	}
}

// — stringRegistrar —

func TestStringRegistrar_no_shorthand_no_dest(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringRegistrar("output", "", "output path", "default.txt", nil)(fs)

	f := fs.Lookup("output")
	if f == nil {
		t.Fatal("flag 'output' not registered")
	}
	if f.Usage != "output path" {
		t.Errorf("Usage = %q, want %q", f.Usage, "output path")
	}
	if f.DefValue != "default.txt" {
		t.Errorf("DefValue = %q, want %q", f.DefValue, "default.txt")
	}
}

func TestStringRegistrar_with_shorthand(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringRegistrar("output", "o", "output path", "", nil)(fs)

	if fs.ShorthandLookup("o") == nil {
		t.Error("shorthand 'o' not registered")
	}
}

func TestStringRegistrar_with_dest(t *testing.T) {
	var dest string
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringRegistrar("output", "", "output path", "", &dest)(fs)

	if err := fs.Parse([]string{"--output", "out.txt"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if dest != "out.txt" {
		t.Errorf("dest = %q, want %q", dest, "out.txt")
	}
}

func TestStringRegistrar_with_shorthand_and_dest(t *testing.T) {
	var dest string
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringRegistrar("output", "o", "output path", "", &dest)(fs)

	if err := fs.Parse([]string{"-o", "out.txt"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if dest != "out.txt" {
		t.Errorf("dest = %q, want %q", dest, "out.txt")
	}
}

// — stringArrayRegistrar —

func TestStringArrayRegistrar_no_shorthand_no_dest(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringArrayRegistrar("tags", "", "list of tags", []string{"a", "b"}, nil)(fs)

	f := fs.Lookup("tags")
	if f == nil {
		t.Fatal("flag 'tags' not registered")
	}
	if f.Usage != "list of tags" {
		t.Errorf("Usage = %q, want %q", f.Usage, "list of tags")
	}
}

func TestStringArrayRegistrar_with_shorthand(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringArrayRegistrar("tags", "t", "list of tags", nil, nil)(fs)

	if fs.ShorthandLookup("t") == nil {
		t.Error("shorthand 't' not registered")
	}
}

func TestStringArrayRegistrar_with_dest(t *testing.T) {
	var dest []string
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringArrayRegistrar("tags", "", "list of tags", nil, &dest)(fs)

	if err := fs.Parse([]string{"--tags", "foo", "--tags", "bar"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(dest) != 2 || dest[0] != "foo" || dest[1] != "bar" {
		t.Errorf("dest = %v, want [foo bar]", dest)
	}
}

func TestStringArrayRegistrar_with_shorthand_and_dest(t *testing.T) {
	var dest []string
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	stringArrayRegistrar("tags", "t", "list of tags", nil, &dest)(fs)

	if err := fs.Parse([]string{"-t", "foo", "-t", "bar"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(dest) != 2 || dest[0] != "foo" || dest[1] != "bar" {
		t.Errorf("dest = %v, want [foo bar]", dest)
	}
}
