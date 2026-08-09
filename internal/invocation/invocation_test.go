// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package invocation

import (
	"errors"
	"os"
	"testing"

	"github.com/leaflockio/core-cli/internal/level"
)

func TestParseFlag(t *testing.T) {
	tests := []struct {
		name         string
		tok          string
		tokens       []string
		i            int
		wantName     string
		wantValue    string
		wantConsumed int
	}{
		{
			name:         "equals form",
			tok:          "--file=./main.go",
			tokens:       []string{"--file=./main.go"},
			i:            0,
			wantName:     "--file",
			wantValue:    "./main.go",
			wantConsumed: 1,
		},
		{
			name:         "value as next token",
			tok:          "--file",
			tokens:       []string{"--file", "./main.go"},
			i:            0,
			wantName:     "--file",
			wantValue:    "./main.go",
			wantConsumed: 2,
		},
		{
			name:         "boolean flag",
			tok:          "--all",
			tokens:       []string{"--all"},
			i:            0,
			wantName:     "--all",
			wantValue:    "",
			wantConsumed: 1,
		},
		{
			name:         "flag followed by another flag is boolean",
			tok:          "--dry-run",
			tokens:       []string{"--dry-run", "--all"},
			i:            0,
			wantName:     "--dry-run",
			wantValue:    "",
			wantConsumed: 1,
		},
		{
			name:         "flag at end of tokens is boolean",
			tok:          "--all",
			tokens:       []string{"license", "add", "--all"},
			i:            2,
			wantName:     "--all",
			wantValue:    "",
			wantConsumed: 1,
		},
		{
			name:         "shorthand boolean flag",
			tok:          "-n",
			tokens:       []string{"-n"},
			i:            0,
			wantName:     "-n",
			wantValue:    "",
			wantConsumed: 1,
		},
		{
			name:         "shorthand equals form",
			tok:          "-f=./main.go",
			tokens:       []string{"-f=./main.go"},
			i:            0,
			wantName:     "-f",
			wantValue:    "./main.go",
			wantConsumed: 1,
		},
		{
			name:         "shorthand value as next token",
			tok:          "-f",
			tokens:       []string{"-f", "./main.go"},
			i:            0,
			wantName:     "-f",
			wantValue:    "./main.go",
			wantConsumed: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, value, consumed := parseFlag(tt.tok, tt.tokens, tt.i)
			if name != tt.wantName {
				t.Errorf("name: got %q, want %q", name, tt.wantName)
			}
			if value != tt.wantValue {
				t.Errorf("value: got %q, want %q", value, tt.wantValue)
			}
			if consumed != tt.wantConsumed {
				t.Errorf("consumed: got %d, want %d", consumed, tt.wantConsumed)
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name   string
		tokens []string
		want   []Flag
	}{
		{
			name:   "empty",
			tokens: []string{},
			want:   nil,
		},
		{
			name:   "no flags",
			tokens: []string{"license", "add", "./main.go"},
			want:   nil,
		},
		{
			name:   "boolean flag",
			tokens: []string{"license", "add", "--all"},
			want:   []Flag{{Name: "--all", Value: ""}},
		},
		{
			name:   "flag with value as next token",
			tokens: []string{"license", "get", "--file", "./main.go"},
			want:   []Flag{{Name: "--file", Value: "./main.go"}},
		},
		{
			name:   "flag with value via equals",
			tokens: []string{"license", "get", "--file=./main.go"},
			want:   []Flag{{Name: "--file", Value: "./main.go"}},
		},
		{
			name:   "repeated flag preserves all entries",
			tokens: []string{"license", "add", "--all", "--var", "ORG=Acme", "--var", "YEAR=2026"},
			want: []Flag{
				{Name: "--all", Value: ""},
				{Name: "--var", Value: "ORG=Acme"},
				{Name: "--var", Value: "YEAR=2026"},
			},
		},
		{
			name:   "mixed flags and non-flags",
			tokens: []string{"license", "check", "--var", "ORG=Acme", "./cmd/main.go", "./cmd/bootstrap.go"},
			want:   []Flag{{Name: "--var", Value: "ORG=Acme"}},
		},
		{
			name:   "short flag treated as boolean when next is flag",
			tokens: []string{"--dry-run", "--all"},
			want: []Flag{
				{Name: "--dry-run", Value: ""},
				{Name: "--all", Value: ""},
			},
		},
		{
			name:   "mix of shorthand and long flags",
			tokens: []string{"license", "check", "-v", "--file", "./main.go", "-n"},
			want: []Flag{
				{Name: "-v", Value: ""},
				{Name: "--file", Value: "./main.go"},
				{Name: "-n", Value: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFlags(tt.tokens)
			if len(got) != len(tt.want) {
				t.Fatalf("len: got %d, want %d — got %v", len(got), len(tt.want), got)
			}
			for i, f := range got {
				if f.Name != tt.want[i].Name || f.Value != tt.want[i].Value {
					t.Errorf("flags[%d]: got {%q, %q}, want {%q, %q}",
						i, f.Name, f.Value, tt.want[i].Name, tt.want[i].Value)
				}
			}
		})
	}
}

func TestFlagMap(t *testing.T) {
	inv := &Invocation{
		Flags: []Flag{
			{Name: "--file", Value: "./main.go"},
			{Name: "--var", Value: "ORG=Acme"},
			{Name: "--var", Value: "YEAR=2026"},
		},
	}
	m := inv.FlagMap()
	if m["--file"] != "./main.go" {
		t.Errorf("--file: got %q, want %q", m["--file"], "./main.go")
	}
	// Last value wins for repeated flags.
	if m["--var"] != "YEAR=2026" {
		t.Errorf("--var: got %q, want %q", m["--var"], "YEAR=2026")
	}
}

func TestFlatFlags(t *testing.T) {
	inv := &Invocation{
		Flags: []Flag{
			{Name: "--all", Value: ""},
			{Name: "--var", Value: "ORG=Acme"},
			{Name: "--var", Value: "YEAR=2026"},
		},
	}
	got := inv.FlatFlags()
	want := []string{"--all", "--var", "ORG=Acme", "--var", "YEAR=2026"}
	if len(got) != len(want) {
		t.Fatalf("len: got %d, want %d — got %v", len(got), len(want), got)
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, v, want[i])
		}
	}
}

func TestFlatFlags_booleanFlagOmitsValue(t *testing.T) {
	inv := &Invocation{Flags: []Flag{{Name: "--all", Value: ""}}}
	got := inv.FlatFlags()
	if len(got) != 1 || got[0] != "--all" {
		t.Errorf("got %v, want [--all]", got)
	}
}

func TestCaptureCWD_error(t *testing.T) {
	orig := getwd
	defer func() { getwd = orig }()
	getwd = func() (string, error) { return "", errors.New("mock") }

	if got := captureCWD(); got != "" {
		t.Errorf("expected empty string on error, got %q", got)
	}
}

func TestFromArgs(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"leaf", "license", "add", "--all", "--var", "ORG=Acme"}
	inv := FromArgs()

	wantRaw := []string{"license", "add", "--all", "--var", "ORG=Acme"}
	if len(inv.Raw) != len(wantRaw) {
		t.Fatalf("Raw len: got %d, want %d", len(inv.Raw), len(wantRaw))
	}
	for i, v := range inv.Raw {
		if v != wantRaw[i] {
			t.Errorf("Raw[%d]: got %q, want %q", i, v, wantRaw[i])
		}
	}

	wantFlags := []Flag{
		{Name: "--all", Value: ""},
		{Name: "--var", Value: "ORG=Acme"},
	}
	if len(inv.Flags) != len(wantFlags) {
		t.Fatalf("Flags len: got %d, want %d", len(inv.Flags), len(wantFlags))
	}
	for i, f := range inv.Flags {
		if f.Name != wantFlags[i].Name || f.Value != wantFlags[i].Value {
			t.Errorf("Flags[%d]: got {%q,%q}, want {%q,%q}", i, f.Name, f.Value, wantFlags[i].Name, wantFlags[i].Value)
		}
	}

	if inv.CWD == "" {
		t.Error("CWD should not be empty")
	}
}

func TestCommandAt(t *testing.T) {
	tests := []struct {
		name string
		raw  []string
		lvl  level.Level
		want string
	}{
		{"root is never represented in Raw", []string{"license", "add"}, level.LevelRoot, ""},
		{"top-level command", []string{"license", "add", "--all"}, level.LevelTop, "license"},
		{"top-level with leading flag", []string{"--verbose=1", "license", "add"}, level.LevelTop, "license"},
		{"nested command", []string{"license", "add", "--all"}, level.LevelNested, "add"},
		{"nested command, multiple segments", []string{"license", "add", "sub"}, level.LevelNested, "add sub"},
		{"nested with interleaved flag", []string{"license", "--all=true", "add"}, level.LevelNested, "add"},
		{"top-level only, no nested token", []string{"license", "--all"}, level.LevelNested, ""},
		{"empty Raw", []string{}, level.LevelTop, ""},
		{"flags only, no positional token", []string{"--all", "--var", "ORG=Acme"}, level.LevelTop, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invocation{Raw: tt.raw}
			if got := inv.CommandAt(tt.lvl); got != tt.want {
				t.Errorf("CommandAt(%v) = %q, want %q", tt.lvl, got, tt.want)
			}
		})
	}
}
