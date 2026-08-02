// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/spf13/cobra"
)

// availableChild returns a *cobra.Command that satisfies IsAvailableCommand,
// so it's eligible for cobra's own suggestion matching.
func availableChild(use string) *cobra.Command {
	return &cobra.Command{Use: use, Run: func(_ *cobra.Command, _ []string) {}}
}

func rootWithChildren(use string, children ...string) *cobra.Command {
	root := &cobra.Command{Use: use}
	for _, c := range children {
		root.AddCommand(availableChild(c))
	}
	return root
}

// rootWithArgsValidator is like rootWithChildren, but also sets Args on the
// root to Meta's real default validator — needed to exercise ShowHelp, since
// found.ValidateArgs only rejects an unrecognized topic when Args is
// actually configured (a bare *cobra.Command's Args defaults to nil, which
// cobra treats as ArbitraryArgs).
func rootWithArgsValidator(use string, children ...string) *cobra.Command {
	root := rootWithChildren(use, children...)
	root.Args = cli.NewMeta(use, "", "").Args
	return root
}

func asErrsError(t *testing.T, err error) *errs.Error {
	t.Helper()
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error = %v, want an *errs.Error", err)
	}
	return e
}

func TestMetaArgs_noPositionalArgs_returnsNil(t *testing.T) {
	m := cli.NewMeta("leaf", "", "")
	cmd := rootWithChildren("leaf", "version")

	if err := m.Args(cmd, nil); err != nil {
		t.Errorf("Args(nil) = %v, want nil", err)
	}
}

func TestMetaArgs_unrecognizedCommand_withoutCloseMatch_isCLI001(t *testing.T) {
	m := cli.NewMeta("leaf", "", "")
	cmd := rootWithChildren("leaf", "version")

	err := m.Args(cmd, []string{"zzzzzzzzzz"})
	e := asErrsError(t, err)

	if e.Code != errs.CLI001 {
		t.Errorf("Code = %q, want %q", e.Code, errs.CLI001)
	}
	if len(e.Contexts) != 1 {
		t.Fatalf("Contexts = %v, want exactly one entry", e.Contexts)
	}
	if !strings.Contains(e.Contexts[0].Resolution, "--help") {
		t.Errorf("Resolution = %q, want a hint to run --help", e.Contexts[0].Resolution)
	}
}

func TestMetaArgs_unrecognizedCommand_withCloseMatch_suggestsExactCommand(t *testing.T) {
	m := cli.NewMeta("leaf", "", "")
	cmd := rootWithChildren("leaf", "license")

	err := m.Args(cmd, []string{"lice"})
	e := asErrsError(t, err)

	want := `did you mean "license"? run "leaf license" instead`
	if e.Contexts[0].Resolution != want {
		t.Errorf("Resolution = %q, want %q", e.Contexts[0].Resolution, want)
	}
}

func TestMetaArgs_unrecognizedCommand_withMultipleCloseMatches_listsAll(t *testing.T) {
	m := cli.NewMeta("leaf", "", "")
	cmd := rootWithChildren("leaf", "add", "addr")

	err := m.Args(cmd, []string{"ad"})
	e := asErrsError(t, err)

	resolution := e.Contexts[0].Resolution
	if !strings.Contains(resolution, "add") || !strings.Contains(resolution, "addr") {
		t.Errorf("Resolution = %q, want it to list both add and addr", resolution)
	}
	if !strings.Contains(resolution, "did you mean one of") {
		t.Errorf("Resolution = %q, want the multi-candidate phrasing", resolution)
	}
}

func TestMetaArgs_unrecognizedCommand_disableSuggestions_fallsBackToGenericResolution(t *testing.T) {
	m := cli.NewMeta("leaf", "", "")
	cmd := rootWithChildren("leaf", "license")
	cmd.DisableSuggestions = true

	err := m.Args(cmd, []string{"lice"})
	e := asErrsError(t, err)

	if strings.Contains(e.Contexts[0].Resolution, "did you mean") {
		t.Errorf("Resolution = %q, DisableSuggestions should suppress suggestions", e.Contexts[0].Resolution)
	}
	if !strings.Contains(e.Contexts[0].Resolution, "--help") {
		t.Errorf("Resolution = %q, want the generic --help hint instead", e.Contexts[0].Resolution)
	}
}

func TestShowHelp_noArgs_showsRootHelp(t *testing.T) {
	root := rootWithArgsValidator("leaf", "version")
	buf := &bytes.Buffer{}
	root.SetOut(buf)

	if err := cli.ShowHelp(root, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected help output to be written")
	}
}

func TestShowHelp_knownCommand_showsItsOwnHelp(t *testing.T) {
	root := rootWithArgsValidator("leaf", "version")
	buf := &bytes.Buffer{}
	root.SetOut(buf)

	if err := cli.ShowHelp(root, []string{"version"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "version") {
		t.Errorf("expected help output to mention version, got %q", buf.String())
	}
}

func TestShowHelp_unknownTopic_rejectsInsteadOfShowingHelp(t *testing.T) {
	root := rootWithArgsValidator("leaf", "version")
	buf := &bytes.Buffer{}
	root.SetOut(buf)

	err := cli.ShowHelp(root, []string{"banana"})
	if err == nil {
		t.Fatal("expected an error for an unrecognized help topic, got nil")
	}
	if buf.Len() != 0 {
		t.Errorf("expected no help output to be written for an unrecognized topic, got %q", buf.String())
	}
}
