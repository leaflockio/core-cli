// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package invocation captures how leaf was called. It is built once from
// os.Args before cobra runs and injected into App so that any command,
// domain service, or lock entry can inspect the full invocation context
// without re-parsing os.Args.
package invocation

import (
	"os"
	"strings"
)

const (
	flagPrefix    = "-"
	flagValuesSep = '='
)

// Flag is a single parsed flag with its value.
// Boolean flags have an empty Value.
// Repeated flags (e.g. --var ORG=X --var YEAR=Y) appear as separate entries.
type Flag struct {
	Name  string // e.g. "--file", "--var", "--all"
	Value string // e.g. "./cmd/main.go", "ORG=Acme", "" for boolean flags
}

// Invocation is an immutable snapshot of how leaf was launched.
type Invocation struct {
	// Raw is os.Args[1:] exactly as the OS provided it.
	Raw []string
	// Flags are the parsed flag-value pairs in declaration order.
	// Boolean flags have an empty Value. Repeated flags appear as multiple entries.
	Flags []Flag
	// CWD is the working directory at the time leaf was launched.
	CWD string
}

// FromArgs constructs an Invocation from os.Args. Call once at program start,
// before cobra parses anything.
func FromArgs() *Invocation {
	raw := captureArgs()
	cwd := captureCWD()
	return &Invocation{
		Raw:   raw,
		Flags: parseFlags(raw),
		CWD:   cwd,
	}
}

// captureArgs returns a copy of os.Args[1:].
func captureArgs() []string {
	raw := make([]string, len(os.Args)-1)
	copy(raw, os.Args[1:])
	return raw
}

// getwd is the os.Getwd function, overridable in tests.
var getwd = os.Getwd

// captureCWD returns the current working directory, empty string on error.
func captureCWD() string {
	cwd, err := getwd()
	if err != nil {
		return ""
	}
	return cwd
}

// parseFlags scans a token list and returns all flag-value pairs.
// Non-flag tokens are skipped. Supports three forms:
//
//	--flag=value   split on first '='
//	--flag value   next token consumed as value when it does not start with '-'
//	--flag         boolean flag, Value is empty string
func parseFlags(tokens []string) []Flag {
	flags := make([]Flag, 0, len(tokens))
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		if !isFlag(tok) {
			i++
			continue
		}
		name, value, consumed := parseFlag(tok, tokens, i)
		flags = append(flags, Flag{Name: name, Value: value})
		i += consumed
	}
	return flags
}

// isFlag reports whether a token is a flag (starts with '-').
func isFlag(tok string) bool {
	return strings.HasPrefix(tok, flagPrefix)
}

// parseFlag parses a single flag token starting at position i in tokens.
// Returns the flag name, its value, and how many tokens were consumed (1 or 2).
func parseFlag(tok string, tokens []string, i int) (string, string, int) {
	// --flag=value form.
	if eq := strings.IndexByte(tok, flagValuesSep); eq != -1 {
		return tok[:eq], tok[eq+1:], 1
	}
	// --flag value form: next token exists and is not itself a flag.
	if i+1 < len(tokens) && !isFlag(tokens[i+1]) {
		return tok, tokens[i+1], 2
	}
	// Boolean flag.
	return tok, "", 1
}

// FlagMap returns the flags as a name→value map for quick lookups.
// When the same flag appears more than once only the last value is kept.
// Use Flags directly when order or repetition matters.
func (inv *Invocation) FlagMap() map[string]string {
	m := make(map[string]string, len(inv.Flags))
	for _, f := range inv.Flags {
		m[f.Name] = f.Value
	}
	return m
}

// FlatFlags returns flags as a flat string slice:
// ["--all", "--var", "ORG=Acme", "--var", "YEAR=2026"].
func (inv *Invocation) FlatFlags() []string {
	out := make([]string, 0, len(inv.Flags)*2)
	for _, f := range inv.Flags {
		out = append(out, f.Name)
		if f.Value != "" {
			out = append(out, f.Value)
		}
	}
	return out
}
